package prepare

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type tpccWorkload struct {
	name        string
	sql         string
	boundSQL    string
	marker      string
	bindingHint tpccBindingHint
	tables      []string
	weight      int
}

type tpccBindingHint string

const (
	tpccBindingHintUseIndex tpccBindingHint = "use_index"
	tpccBindingHintTiFlash  tpccBindingHint = "read_from_storage_tiflash"
)

func drainQuery(ctx context.Context, conn *sql.Conn, statement string) error {
	rows, err := conn.QueryContext(ctx, statement)
	if err != nil {
		return err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	values := make([]sql.RawBytes, len(columns))
	destinations := make([]any, len(columns))
	for i := range values {
		destinations[i] = &values[i]
	}
	for rows.Next() {
		if err := rows.Scan(destinations...); err != nil {
			return err
		}
	}
	return rows.Err()
}

type planOperator struct {
	id   string
	task string
}

func explainPlanOperators(ctx context.Context, conn *sql.Conn, statement string) ([]planOperator, error) {
	rows, err := conn.QueryContext(ctx, "EXPLAIN "+statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	idColumn, taskColumn := -1, -1
	for i, column := range columns {
		switch strings.ToLower(column) {
		case "id":
			idColumn = i
		case "task":
			taskColumn = i
		}
	}
	if idColumn < 0 || taskColumn < 0 {
		return nil, fmt.Errorf("EXPLAIN result must contain id and task columns: %v", columns)
	}
	var operators []planOperator
	for rows.Next() {
		values := make([]sql.NullString, len(columns))
		destinations := make([]any, len(columns))
		for i := range values {
			destinations[i] = &values[i]
		}
		if err := rows.Scan(destinations...); err != nil {
			return nil, err
		}
		operators = append(operators, planOperator{
			id:   normalizedOperatorID(values[idColumn].String),
			task: values[taskColumn].String,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(operators) == 0 {
		return nil, errors.New("EXPLAIN returned no operators")
	}
	return operators, nil
}

func normalizedOperatorID(id string) string {
	separator := strings.LastIndexByte(id, '_')
	if separator < 0 || separator == len(id)-1 {
		return id
	}
	for _, char := range id[separator+1:] {
		if char < '0' || char > '9' {
			return id
		}
	}
	return id[:separator]
}
