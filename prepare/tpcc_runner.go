package prepare

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

const workloadCleanupTimeout = 10 * time.Second

type weightedTPCCPicker struct {
	cumulativeWeights []int
	totalWeight       int
}

func newWeightedTPCCPicker(workloads []tpccWorkload) (weightedTPCCPicker, error) {
	picker := weightedTPCCPicker{cumulativeWeights: make([]int, len(workloads))}
	for i, workload := range workloads {
		if workload.weight <= 0 {
			return weightedTPCCPicker{}, fmt.Errorf("workload %s has invalid weight %d", workload.name, workload.weight)
		}
		picker.totalWeight += workload.weight
		picker.cumulativeWeights[i] = picker.totalWeight
	}
	if picker.totalWeight == 0 {
		return weightedTPCCPicker{}, errors.New("TPC-C workload list is empty")
	}
	return picker, nil
}

func (p weightedTPCCPicker) pick(draw int) int {
	return sort.Search(len(p.cumulativeWeights), func(i int) bool {
		return draw < p.cumulativeWeights[i]
	})
}

type tpccWorkloadWorker struct {
	id   int
	conn *sql.Conn
}

type tpccWorkerResult struct {
	id         int
	executions uint64
	err        error
}

func (p *preparer) prepareTPCCStatements(ctx context.Context) error {
	if err := p.validateTPCCBindings(ctx); err != nil {
		return err
	}
	picker, err := newWeightedTPCCPicker(tpccWorkloads)
	if err != nil {
		return err
	}
	workers, err := p.createTPCCWorkers(ctx)
	if err != nil {
		return err
	}

	executionCounts := make([]atomic.Uint64, len(tpccWorkloads))
	runCtx, cancel := context.WithTimeout(ctx, p.duration)
	results := make(chan tpccWorkerResult, len(workers))
	var waitGroup sync.WaitGroup
	for _, worker := range workers {
		waitGroup.Add(1)
		go func(worker tpccWorkloadWorker) {
			defer waitGroup.Done()
			executions, runErr := runTPCCWorker(runCtx, worker, picker, executionCounts)
			results <- tpccWorkerResult{id: worker.id, executions: executions, err: runErr}
			if runErr != nil {
				cancel()
			}
		}(worker)
	}
	waitGroup.Wait()
	cancel()
	close(results)

	var runErr error
	var totalExecutions uint64
	for result := range results {
		totalExecutions += result.executions
		if result.err != nil {
			runErr = errors.Join(runErr, fmt.Errorf("worker %d: %w", result.id, result.err))
		}
	}
	if err := ctx.Err(); err != nil {
		runErr = errors.Join(runErr, err)
	}

	cleanupCtx, cleanupCancel := context.WithTimeout(context.WithoutCancel(ctx), workloadCleanupTimeout)
	cleanupErr := closeTPCCWorkers(cleanupCtx, workers)
	cleanupCancel()

	for i, workload := range tpccWorkloads {
		p.logger.Info("TPC-C workload execution summary",
			zap.String("workload", workload.name),
			zap.Int("weight", workload.weight),
			zap.String("binding_hint", string(workload.bindingHint)),
			zap.Uint64("executions", executionCounts[i].Load()),
		)
	}
	p.logger.Info("completed concurrent TPC-C workload",
		zap.Int("concurrency", p.concurrency),
		zap.Duration("duration", p.duration),
		zap.Uint64("executions", totalExecutions),
	)
	return errors.Join(runErr, cleanupErr)
}

func (p *preparer) validateTPCCBindings(ctx context.Context) error {
	for _, workload := range tpccWorkloads {
		naturalPlan, err := explainPlanOperators(ctx, p.conn, workload.sql)
		if err != nil {
			return fmt.Errorf("explain natural plan for workload %s: %w", workload.name, err)
		}
		boundPlan, err := p.explainWithSessionBinding(ctx, workload)
		if err != nil {
			return fmt.Errorf("validate session binding for workload %s: %w", workload.name, err)
		}
		if slices.Equal(boundPlan, naturalPlan) {
			return fmt.Errorf("workload %s has the same bound and natural id/task operators", workload.name)
		}
	}
	return nil
}

func (p *preparer) explainWithSessionBinding(ctx context.Context, workload tpccWorkload) (plan []planOperator, returnErr error) {
	createBinding := "CREATE SESSION BINDING FOR " + workload.sql + " USING " + workload.boundSQL
	dropBinding := "DROP SESSION BINDING FOR " + workload.sql
	if _, err := p.conn.ExecContext(ctx, createBinding); err != nil {
		return nil, fmt.Errorf("create session binding: %w", err)
	}
	bindingActive := true
	defer func() {
		if !bindingActive {
			return
		}
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		defer cancel()
		if _, err := p.conn.ExecContext(cleanupCtx, dropBinding); err != nil {
			returnErr = errors.Join(returnErr, fmt.Errorf("drop validation binding: %w", err))
		}
	}()

	plan, err := explainPlanOperators(ctx, p.conn, workload.sql)
	if err != nil {
		return nil, fmt.Errorf("explain bound statement: %w", err)
	}
	if _, err := p.conn.ExecContext(ctx, dropBinding); err != nil {
		return nil, fmt.Errorf("drop session binding: %w", err)
	}
	bindingActive = false
	return plan, nil
}

func (p *preparer) createTPCCWorkers(ctx context.Context) ([]tpccWorkloadWorker, error) {
	workers := make([]tpccWorkloadWorker, 0, p.concurrency)
	for workerID := 1; workerID <= p.concurrency; workerID++ {
		conn, err := p.db.Conn(ctx)
		if err != nil {
			closeConnections(workers)
			return nil, fmt.Errorf("acquire connection for worker %d: %w", workerID, err)
		}
		if _, err := conn.ExecContext(ctx, "USE "+quoteIdentifier(p.schema)); err != nil {
			_ = conn.Close()
			closeConnections(workers)
			return nil, fmt.Errorf("select schema for worker %d: %w", workerID, err)
		}
		for _, workload := range tpccWorkloads {
			statement := "CREATE SESSION BINDING FOR " + workload.sql + " USING " + workload.boundSQL
			if _, err := conn.ExecContext(ctx, statement); err != nil {
				_ = conn.Close()
				closeConnections(workers)
				return nil, fmt.Errorf("create binding for worker %d workload %s: %w", workerID, workload.name, err)
			}
		}
		workers = append(workers, tpccWorkloadWorker{id: workerID, conn: conn})
	}
	return workers, nil
}

func runTPCCWorker(ctx context.Context, worker tpccWorkloadWorker, picker weightedTPCCPicker, executionCounts []atomic.Uint64) (uint64, error) {
	var executions uint64
	checkedBinding := false
	for {
		if ctx.Err() != nil {
			return executions, nil
		}
		workloadIndex := picker.pick(rand.IntN(picker.totalWeight))
		workload := tpccWorkloads[workloadIndex]
		if err := drainQuery(ctx, worker.conn, workload.sql); err != nil {
			if ctx.Err() != nil {
				return executions, nil
			}
			return executions, fmt.Errorf("execute workload %s: %w", workload.name, err)
		}
		executions++
		executionCounts[workloadIndex].Add(1)

		if !checkedBinding {
			var usedBinding int
			if err := worker.conn.QueryRowContext(ctx, "SELECT @@last_plan_from_binding").Scan(&usedBinding); err != nil {
				return executions, fmt.Errorf("read last_plan_from_binding: %w", err)
			}
			if usedBinding != 1 {
				return executions, fmt.Errorf("workload %s did not use its session binding", workload.name)
			}
			checkedBinding = true
		}
	}
}

func closeTPCCWorkers(ctx context.Context, workers []tpccWorkloadWorker) error {
	var cleanupErr error
	for _, worker := range workers {
		for _, workload := range tpccWorkloads {
			statement := "DROP SESSION BINDING FOR " + workload.sql
			if _, err := worker.conn.ExecContext(ctx, statement); err != nil {
				cleanupErr = errors.Join(cleanupErr, fmt.Errorf("drop worker %d binding for workload %s: %w", worker.id, workload.name, err))
				break
			}
		}
		if err := worker.conn.Close(); err != nil {
			cleanupErr = errors.Join(cleanupErr, fmt.Errorf("close worker %d connection: %w", worker.id, err))
		}
	}
	return cleanupErr
}

func closeConnections(workers []tpccWorkloadWorker) {
	for _, worker := range workers {
		_ = worker.conn.Close()
	}
}
