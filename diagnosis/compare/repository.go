package compare

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"strings"

	"github.com/xzhangxian1008/OptimizerCanaryService/diagnosis/util"
	"go.uber.org/zap"
)

type SQLRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewSQLRepository(db *sql.DB, logger *zap.Logger) *SQLRepository {
	return &SQLRepository{db: db, logger: logger}
}

func (r *SQLRepository) Bindings(ctx context.Context, digest, schema string) (string, error) {
	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	r.logger.Info("execute SQL", zap.String("sql", compareBindingsQuery), zap.String("digest", digest), zap.String("schema", schema))
	rows, err := r.db.QueryContext(queryCtx, compareBindingsQuery, digest, schema)
	if err != nil {
		return "", util.NewStackErrorf("query bindings: %w", err)
	}
	defer rows.Close()
	columns, values, err := readStringRows(rows)
	if err != nil {
		return "", fmt.Errorf("read bindings: %w", err)
	}
	var binding strings.Builder
	for i, row := range values {
		if i > 0 {
			binding.WriteByte('\n')
		}
		for j, value := range row {
			fmt.Fprintf(&binding, "%s: %s\n", columns[j], value)
		}
	}
	return binding.String(), nil
}

func (r *SQLRepository) GetSQLInfo(ctx context.Context) ([]SQLInfo, error) {
	queryCtx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	r.logger.Info("execute SQL", zap.String("sql", compareSqlInfoQuery))
	rows, err := r.db.QueryContext(queryCtx, compareSqlInfoQuery)
	if err != nil {
		return nil, util.NewStackErrorf("sql infos: %w", err)
	}
	defer rows.Close()
	var sqlInfos []SQLInfo
	for rows.Next() {
		var sqlInfo SQLInfo
		if err := rows.Scan(&sqlInfo.Schema, &sqlInfo.SQLDigest, &sqlInfo.SQL, &sqlInfo.ExecCount, &sqlInfo.PlanDigest, &sqlInfo.Plan, &sqlInfo.ExecTime, &sqlInfo.PlanHint); err != nil {
			return nil, util.NewStackErrorf("scan sql info: %w", err)
		}
		if sqlInfo.SQLDigest == "" || sqlInfo.PlanDigest == "" || strings.TrimSpace(sqlInfo.SQL) == "" {
			return nil, util.NewStackErrorf("cluster sql info is missing SQL or digest (SQL digest %q)", sqlInfo.SQLDigest)
		}
		sqlInfos = append(sqlInfos, sqlInfo)
	}
	if err := rows.Err(); err != nil {
		return nil, util.NewStackErrorf("read sql infos: %w", err)
	}
	return sqlInfos, nil
}

func (r *SQLRepository) ExplainPlan(ctx context.Context, sample Sample) (string, error) {
	plan, _, err := r.ExplainPlanWithDigest(ctx, sample)
	return plan, err
}

func (r *SQLRepository) ExplainPlanWithDigest(ctx context.Context, sample Sample) (string, string, error) {
	explainCtx, cancel := context.WithTimeout(ctx, explainTimeout)
	defer cancel()
	conn, err := r.db.Conn(explainCtx)
	if err != nil {
		return "", "", util.NewStackErrorf("acquire connection: %w", err)
	}
	defer conn.Close()
	if sample.Schema != "" {
		useSQL := "USE `" + strings.ReplaceAll(sample.Schema, "`", "``") + "`"
		r.logger.Info("execute SQL", zap.String("sql", useSQL))
		if _, err := conn.ExecContext(explainCtx, useSQL); err != nil {
			return "", "", util.NewStackErrorf("select schema %q: %w", sample.Schema, err)
		}
	}
	statement, arguments, err := explainableSample(sample.SQL)
	if err != nil {
		return "", "", fmt.Errorf("parse prepared statement sample: %w", err)
	}
	if statement == "" {
		return "", "", util.NewStackErrorf("cannot explain an empty statement")
	}
	explainSQL := "EXPLAIN " + statement
	r.logger.Info("execute SQL", zap.String("schema", sample.Schema), zap.String("sql", explainSQL))
	rows, err := conn.QueryContext(explainCtx, explainSQL, arguments...)
	if err != nil {
		return "", "", util.NewStackErrorf("explain statement: %w", err)
	}
	defer rows.Close()
	columns, values, err := readStringRows(rows)
	if err != nil {
		return "", "", fmt.Errorf("read EXPLAIN result: %w", err)
	}
	var plan strings.Builder
	plan.WriteString(strings.Join(columns, "\t"))
	for _, row := range values {
		plan.WriteByte('\n')
		plan.WriteString(strings.Join(row, "\t"))
	}
	planText := plan.String()
	planDigest, err := r.explainPlanDigest(explainCtx, conn, sample.Schema, explainSQL, planText)
	if err != nil {
		return "", "", err
	}
	return planText, planDigest, nil
}

func (r *SQLRepository) explainPlanDigest(ctx context.Context, conn *sql.Conn, schema, explainSQL, newPlan string) (string, error) {
	rows, err := conn.QueryContext(ctx, explainPlanDigestQuery, schema, schema)
	if err != nil {
		return "", util.NewStackErrorf("query EXPLAIN plan digest: %w", err)
	}
	defer rows.Close()
	newOperators, err := planOperators(newPlan)
	if err != nil {
		return "", fmt.Errorf("read EXPLAIN plan for digest lookup: %w", err)
	}
	newDigestOperators := digestLookupOperators(newOperators)
	normalizedExplainSQL := normalizeExplainSQL(explainSQL)
	for rows.Next() {
		var querySample, digest, plan string
		if err := rows.Scan(&querySample, &digest, &plan); err != nil {
			return "", util.NewStackErrorf("scan EXPLAIN plan digest: %w", err)
		}
		if normalizeExplainSQL(querySample) != normalizedExplainSQL {
			continue
		}
		operators, err := planOperators(plan)
		if err == nil && slices.Equal(newDigestOperators, digestLookupOperators(operators)) {
			return digest, nil
		}
	}
	if err := rows.Err(); err != nil {
		return "", util.NewStackErrorf("read EXPLAIN plan digests: %w", err)
	}
	return "", util.NewStackErrorf("no plan digest found for EXPLAIN of SQL in schema %q", schema)
}

var _ CompareRepository = (*SQLRepository)(nil)
