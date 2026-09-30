package prepare

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	tpccTiFlashReplicaCount = 1
	tiFlashWaitTimeout      = 10 * time.Minute
	tiFlashPollInterval     = 2 * time.Second
)

type tiFlashReplicaStatus struct {
	available bool
	progress  float64
}

func (p *preparer) prepareTPCCTiFlashReplicas(ctx context.Context) error {
	for _, table := range tpccTableDDLs {
		p.logger.Info("set TPC-C TiFlash replica",
			zap.String("table", table.name),
			zap.Int("replica_count", tpccTiFlashReplicaCount),
		)
		statement := fmt.Sprintf(
			"ALTER TABLE %s SET TIFLASH REPLICA %d",
			quoteIdentifier(table.name),
			tpccTiFlashReplicaCount,
		)
		if _, err := p.conn.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("set TiFlash replica for table %s: %w", table.name, err)
		}
	}

	waitCtx, cancel := context.WithTimeout(ctx, tiFlashWaitTimeout)
	defer cancel()
	return p.waitForTPCCTiFlashReplicas(waitCtx)
}

func (p *preparer) waitForTPCCTiFlashReplicas(ctx context.Context) error {
	var pending []string
	for {
		statuses, err := p.loadTiFlashReplicaStatuses(ctx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) && len(pending) > 0 {
				return fmt.Errorf("wait for TiFlash replicas timed out; pending tables: %s", strings.Join(pending, ", "))
			}
			return err
		}

		pending = pendingTiFlashTables(statuses)
		if len(pending) == 0 {
			p.logger.Info("all TPC-C TiFlash replicas are available",
				zap.Int("tables", len(tpccTableDDLs)),
			)
			return nil
		}
		p.logger.Info("waiting for TPC-C TiFlash replicas",
			zap.Strings("pending_tables", pending),
		)

		timer := time.NewTimer(tiFlashPollInterval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {
				return fmt.Errorf("wait for TiFlash replicas timed out; pending tables: %s", strings.Join(pending, ", "))
			}
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func (p *preparer) loadTiFlashReplicaStatuses(ctx context.Context) (map[string]tiFlashReplicaStatus, error) {
	const query = `SELECT table_name, available, progress
	FROM information_schema.tiflash_replica
	WHERE table_schema = ?`
	rows, err := p.conn.QueryContext(ctx, query, p.schema)
	if err != nil {
		return nil, fmt.Errorf("query TiFlash replica status: %w", err)
	}
	defer rows.Close()

	statuses := make(map[string]tiFlashReplicaStatus, len(tpccTableDDLs))
	for rows.Next() {
		var (
			tableName string
			available int
			progress  float64
		)
		if err := rows.Scan(&tableName, &available, &progress); err != nil {
			return nil, fmt.Errorf("scan TiFlash replica status: %w", err)
		}
		statuses[tableName] = tiFlashReplicaStatus{
			available: available == 1,
			progress:  progress,
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate TiFlash replica status: %w", err)
	}
	return statuses, nil
}

func pendingTiFlashTables(statuses map[string]tiFlashReplicaStatus) []string {
	pending := make([]string, 0, len(tpccTableDDLs))
	for _, table := range tpccTableDDLs {
		status, exists := statuses[table.name]
		if !exists || !status.available {
			pending = append(pending, table.name)
		}
	}
	sort.Strings(pending)
	return pending
}
