package prepare

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

const (
	tpccWarehouseCount      = 100
	tpccDistrictCount       = 10
	tpccCustomerCount       = 1000
	tpccOrderCount          = 1000
	tpccOrderLinesPerOrder  = 5
	tpccItemCount           = 1000
	tpccInsertBatchSize     = 1000
	tpccFixtureTimestamp    = "2026-01-01 00:00:00"
	tpccTargetStockQuantity = 15
)

var tpccTableDDLs = []struct {
	name string
	sql  string
}{
	{"warehouse", `CREATE TABLE IF NOT EXISTS warehouse (
		w_id INT NOT NULL,
		w_name VARCHAR(10),
		w_street_1 VARCHAR(20),
		w_street_2 VARCHAR(20),
		w_city VARCHAR(20),
		w_state CHAR(2),
		w_zip CHAR(9),
		w_tax DECIMAL(4, 4),
		w_ytd DECIMAL(12, 2),
		PRIMARY KEY (w_id)
	)`},
	{"district", `CREATE TABLE IF NOT EXISTS district (
		d_id INT NOT NULL,
		d_w_id INT NOT NULL,
		d_name VARCHAR(10),
		d_street_1 VARCHAR(20),
		d_street_2 VARCHAR(20),
		d_city VARCHAR(20),
		d_state CHAR(2),
		d_zip CHAR(9),
		d_tax DECIMAL(4, 4),
		d_ytd DECIMAL(12, 2),
		d_next_o_id INT,
		PRIMARY KEY (d_w_id, d_id)
	)`},
	{"customer", `CREATE TABLE IF NOT EXISTS customer (
		c_id INT NOT NULL,
		c_d_id INT NOT NULL,
		c_w_id INT NOT NULL,
		c_first VARCHAR(16),
		c_middle CHAR(2),
		c_last VARCHAR(16),
		c_street_1 VARCHAR(20),
		c_street_2 VARCHAR(20),
		c_city VARCHAR(20),
		c_state CHAR(2),
		c_zip CHAR(9),
		c_phone CHAR(16),
		c_since DATETIME,
		c_credit CHAR(2),
		c_credit_lim DECIMAL(12, 2),
		c_discount DECIMAL(4, 4),
		c_balance DECIMAL(12, 2),
		c_ytd_payment DECIMAL(12, 2),
		c_payment_cnt INT,
		c_delivery_cnt INT,
		c_data VARCHAR(500),
		PRIMARY KEY (c_w_id, c_d_id, c_id),
		INDEX idx_customer (c_w_id, c_d_id, c_last, c_first)
	)`},
	{"history", `CREATE TABLE IF NOT EXISTS history (
		h_c_id INT NOT NULL,
		h_c_d_id INT NOT NULL,
		h_c_w_id INT NOT NULL,
		h_d_id INT NOT NULL,
		h_w_id INT NOT NULL,
		h_date DATETIME,
		h_amount DECIMAL(6, 2),
		h_data VARCHAR(24),
		INDEX idx_h_w_id (h_w_id),
		INDEX idx_h_c_w_id (h_c_w_id)
	)`},
	{"new_order", `CREATE TABLE IF NOT EXISTS new_order (
		no_o_id INT NOT NULL,
		no_d_id INT NOT NULL,
		no_w_id INT NOT NULL,
		PRIMARY KEY (no_w_id, no_d_id, no_o_id)
	)`},
	{"orders", `CREATE TABLE IF NOT EXISTS orders (
		o_id INT NOT NULL,
		o_d_id INT NOT NULL,
		o_w_id INT NOT NULL,
		o_c_id INT,
		o_entry_d DATETIME,
		o_carrier_id INT,
		o_ol_cnt INT,
		o_all_local INT,
		PRIMARY KEY (o_w_id, o_d_id, o_id),
		INDEX idx_order (o_w_id, o_d_id, o_c_id, o_id)
	)`},
	{"order_line", `CREATE TABLE IF NOT EXISTS order_line (
		ol_o_id INT NOT NULL,
		ol_d_id INT NOT NULL,
		ol_w_id INT NOT NULL,
		ol_number INT NOT NULL,
		ol_i_id INT NOT NULL,
		ol_supply_w_id INT,
		ol_delivery_d DATETIME,
		ol_quantity INT,
		ol_amount DECIMAL(6, 2),
		ol_dist_info CHAR(24),
		PRIMARY KEY (ol_w_id, ol_d_id, ol_o_id, ol_number)
	)`},
	{"stock", `CREATE TABLE IF NOT EXISTS stock (
		s_i_id INT NOT NULL,
		s_w_id INT NOT NULL,
		s_quantity INT,
		s_dist_01 CHAR(24),
		s_dist_02 CHAR(24),
		s_dist_03 CHAR(24),
		s_dist_04 CHAR(24),
		s_dist_05 CHAR(24),
		s_dist_06 CHAR(24),
		s_dist_07 CHAR(24),
		s_dist_08 CHAR(24),
		s_dist_09 CHAR(24),
		s_dist_10 CHAR(24),
		s_ytd INT,
		s_order_cnt INT,
		s_remote_cnt INT,
		s_data VARCHAR(50),
		PRIMARY KEY (s_w_id, s_i_id)
	)`},
	{"item", `CREATE TABLE IF NOT EXISTS item (
		i_id INT NOT NULL,
		i_im_id INT,
		i_name VARCHAR(24),
		i_price DECIMAL(5, 2),
		i_data VARCHAR(50),
		PRIMARY KEY (i_id)
	)`},
}

// prepareTPCCSchema creates the nine core tables used by PingCAP go-tpc and
// loads a compact, deterministic data set suitable for optimizer experiments.
func (p *preparer) prepareTPCCSchema(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	quotedSchema := quoteIdentifier(p.schema)
	if _, err := p.conn.ExecContext(ctx, "CREATE DATABASE IF NOT EXISTS "+quotedSchema); err != nil {
		return fmt.Errorf("create schema %q: %w", p.schema, err)
	}
	if _, err := p.conn.ExecContext(ctx, "USE "+quotedSchema); err != nil {
		return fmt.Errorf("select schema %q: %w", p.schema, err)
	}
	for _, table := range tpccTableDDLs {
		p.logger.Info("create TPC-C table", zap.String("table", table.name))
		if _, err := p.conn.ExecContext(ctx, table.sql); err != nil {
			return fmt.Errorf("create table %s: %w", table.name, err)
		}
	}
	if err := p.seedTPCCData(ctx); err != nil {
		return err
	}
	if err := p.prepareTPCCTiFlashReplicas(ctx); err != nil {
		return err
	}
	for _, table := range []string{"warehouse", "district", "customer", "orders", "order_line", "stock", "item"} {
		if _, err := p.conn.ExecContext(ctx, "ANALYZE TABLE "+quoteIdentifier(table)); err != nil {
			return fmt.Errorf("analyze table %s: %w", table, err)
		}
	}
	return nil
}

func quoteIdentifier(identifier string) string {
	return "`" + strings.ReplaceAll(identifier, "`", "``") + "`"
}

func (p *preparer) seedTPCCData(ctx context.Context) error {
	started := time.Now()
	warehouses := make([][]any, 0, tpccWarehouseCount)
	for warehouseID := 1; warehouseID <= tpccWarehouseCount; warehouseID++ {
		warehouses = append(warehouses, []any{warehouseID, fmt.Sprintf("W%d", warehouseID), "Street 1", "Street 2", "Portland", "OR", "972010000", 0.1, 300000.0})
	}
	if err := p.insertRows(ctx, "warehouse",
		[]string{"w_id", "w_name", "w_street_1", "w_street_2", "w_city", "w_state", "w_zip", "w_tax", "w_ytd"},
		warehouses); err != nil {
		return err
	}

	districts := make([][]any, 0, tpccWarehouseCount*tpccDistrictCount)
	for warehouseID := 1; warehouseID <= tpccWarehouseCount; warehouseID++ {
		for districtID := 1; districtID <= tpccDistrictCount; districtID++ {
			districts = append(districts, []any{districtID, warehouseID, fmt.Sprintf("D%d", districtID), "Street 1", "Street 2", "Portland", "OR", "972010000", 0.1, 30000.0, tpccOrderCount + 1})
		}
	}
	if err := p.insertRows(ctx, "district",
		[]string{"d_id", "d_w_id", "d_name", "d_street_1", "d_street_2", "d_city", "d_state", "d_zip", "d_tax", "d_ytd", "d_next_o_id"}, districts); err != nil {
		return err
	}

	items := make([][]any, 0, tpccItemCount)
	for id := 1; id <= tpccItemCount; id++ {
		items = append(items, []any{id, id % 100, fmt.Sprintf("ITEM%04d", id), 1.0 + float64(id%100), "optimizer canary item"})
	}
	if err := p.insertRows(ctx, "item", []string{"i_id", "i_im_id", "i_name", "i_price", "i_data"}, items); err != nil {
		return err
	}

	for warehouseID := 1; warehouseID <= tpccWarehouseCount; warehouseID++ {
		if err := p.seedTPCCWarehouse(ctx, warehouseID); err != nil {
			return err
		}
		if warehouseID%10 == 0 || warehouseID == tpccWarehouseCount {
			p.logger.Info("seeded TPC-C warehouses",
				zap.Int("completed", warehouseID),
				zap.Int("total", tpccWarehouseCount),
			)
		}
	}
	p.logger.Info("seeded compact TPC-C fixture",
		zap.Int("warehouses", tpccWarehouseCount),
		zap.Int("rows", tpccFixtureRowCount()),
		zap.Duration("elapsed", time.Since(started)),
	)
	return nil
}

func (p *preparer) seedTPCCWarehouse(ctx context.Context, warehouseID int) error {
	customers := make([][]any, 0, tpccCustomerCount)
	orders := make([][]any, 0, tpccOrderCount)
	newOrders := make([][]any, 0, tpccOrderCount*3/10)
	orderLines := make([][]any, 0, tpccOrderCount*tpccOrderLinesPerOrder)
	for id := 1; id <= tpccCustomerCount; id++ {
		last := fmt.Sprintf("LAST%03d", id%100)
		customers = append(customers, []any{id, 1, warehouseID, fmt.Sprintf("C%04d", id), "OE", last, "Street 1", "Street 2", "Portland", "OR", "972010000", fmt.Sprintf("%016d", id), tpccFixtureTimestamp, "GC", 50000.0, 0.05, -10.0, 10.0, 1, 0, "optimizer canary TPC-C fixture"})
		orders = append(orders, []any{id, 1, warehouseID, id, tpccFixtureTimestamp, (id % 10) + 1, tpccOrderLinesPerOrder, 1})
		if id > tpccOrderCount*7/10 {
			newOrders = append(newOrders, []any{id, 1, warehouseID})
		}
		for line := 1; line <= tpccOrderLinesPerOrder; line++ {
			itemID := ((id-1)*tpccOrderLinesPerOrder+line-1)%tpccItemCount + 1
			orderLines = append(orderLines, []any{id, 1, warehouseID, line, itemID, warehouseID, tpccFixtureTimestamp, 5, 1.0 + float64(line), "optimizer canary dist"})
		}
	}
	if err := p.insertRows(ctx, "customer",
		[]string{"c_id", "c_d_id", "c_w_id", "c_first", "c_middle", "c_last", "c_street_1", "c_street_2", "c_city", "c_state", "c_zip", "c_phone", "c_since", "c_credit", "c_credit_lim", "c_discount", "c_balance", "c_ytd_payment", "c_payment_cnt", "c_delivery_cnt", "c_data"}, customers); err != nil {
		return err
	}
	if err := p.insertRows(ctx, "orders",
		[]string{"o_id", "o_d_id", "o_w_id", "o_c_id", "o_entry_d", "o_carrier_id", "o_ol_cnt", "o_all_local"}, orders); err != nil {
		return err
	}
	if err := p.insertRows(ctx, "new_order", []string{"no_o_id", "no_d_id", "no_w_id"}, newOrders); err != nil {
		return err
	}
	if err := p.insertRows(ctx, "order_line",
		[]string{"ol_o_id", "ol_d_id", "ol_w_id", "ol_number", "ol_i_id", "ol_supply_w_id", "ol_delivery_d", "ol_quantity", "ol_amount", "ol_dist_info"}, orderLines); err != nil {
		return err
	}

	stocks := make([][]any, 0, tpccItemCount)
	for id := 1; id <= tpccItemCount; id++ {
		dist := fmt.Sprintf("DIST%04d", id)
		stocks = append(stocks, []any{id, warehouseID, tpccTargetStockQuantity + id%20, dist, dist, dist, dist, dist, dist, dist, dist, dist, dist, 0, 0, 0, "optimizer canary stock"})
	}
	return p.insertRows(ctx, "stock",
		[]string{"s_i_id", "s_w_id", "s_quantity", "s_dist_01", "s_dist_02", "s_dist_03", "s_dist_04", "s_dist_05", "s_dist_06", "s_dist_07", "s_dist_08", "s_dist_09", "s_dist_10", "s_ytd", "s_order_cnt", "s_remote_cnt", "s_data"}, stocks)
}

func tpccFixtureRowCount() int {
	perWarehouse := 1 + tpccDistrictCount + tpccCustomerCount + tpccOrderCount +
		tpccOrderCount*3/10 + tpccOrderCount*tpccOrderLinesPerOrder + tpccItemCount
	return tpccWarehouseCount*perWarehouse + tpccItemCount
}

func (p *preparer) insertRows(ctx context.Context, table string, columns []string, rows [][]any) error {
	for start := 0; start < len(rows); start += tpccInsertBatchSize {
		end := min(start+tpccInsertBatchSize, len(rows))
		var query strings.Builder
		query.WriteString("INSERT IGNORE INTO ")
		query.WriteString(quoteIdentifier(table))
		query.WriteString(" (")
		for i, column := range columns {
			if i > 0 {
				query.WriteByte(',')
			}
			query.WriteString(quoteIdentifier(column))
		}
		query.WriteString(") VALUES ")
		arguments := make([]any, 0, (end-start)*len(columns))
		for i, row := range rows[start:end] {
			if len(row) != len(columns) {
				return fmt.Errorf("insert %s: row has %d values for %d columns", table, len(row), len(columns))
			}
			if i > 0 {
				query.WriteByte(',')
			}
			query.WriteByte('(')
			for j := range row {
				if j > 0 {
					query.WriteByte(',')
				}
				query.WriteByte('?')
			}
			query.WriteByte(')')
			arguments = append(arguments, row...)
		}
		if _, err := p.conn.ExecContext(ctx, query.String(), arguments...); err != nil {
			return fmt.Errorf("insert %s rows %d-%d: %w", table, start, end-1, err)
		}
	}
	return nil
}
