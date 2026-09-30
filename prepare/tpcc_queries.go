package prepare

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

// tpccWorkloads contains every query executed by prepare mode. Different
// projections, predicates, ordering, aggregation, and joins keep their
// normalized SQL digests distinct.
var tpccWorkloads = assignRandomWeights([]tpccWorkload{
	newTPCCWorkload(
		"new_order_customer",
		`SELECT c_discount, c_last, c_credit, w_tax
		 FROM customer, warehouse
		 WHERE w_id = 1 AND c_w_id = w_id AND c_d_id = 1 AND c_id = 7`,
		"customer", "warehouse",
	),
	newTPCCWorkload(
		"order_status_customer_by_last",
		`SELECT c_balance, c_first, c_middle, c_id
		 FROM customer
		 WHERE c_w_id = 1 AND c_d_id = 1 AND c_last = 'LAST007'
		 ORDER BY c_first`,
		"customer",
	),
	newTPCCWorkload(
		"order_status_latest_order",
		`SELECT o_id, o_carrier_id, o_entry_d
		 FROM orders
		 WHERE o_w_id = 1 AND o_d_id = 1 AND o_c_id = 7
		 ORDER BY o_id DESC LIMIT 1`,
		"orders",
	),
	newTPCCWorkload(
		"order_status_order_lines",
		`SELECT ol_i_id, ol_supply_w_id, ol_quantity, ol_amount, ol_delivery_d
		 FROM order_line
		 WHERE ol_w_id = 1 AND ol_d_id = 1 AND ol_o_id = 907`,
		"order_line",
	),
	newTPCCWorkload(
		"new_order_items",
		`SELECT i_price, i_name, i_data, i_id
		 FROM item
		 WHERE i_id IN (7, 17, 27, 37, 47)`,
		"item",
	),
	newTPCCWorkload(
		"stock_level",
		`SELECT COUNT(DISTINCT s_i_id) AS stock_count
		 FROM order_line, stock
		 WHERE ol_w_id = 1 AND ol_d_id = 1
		   AND ol_o_id < 1001 AND ol_o_id >= 981
		   AND s_w_id = 1 AND s_i_id = ol_i_id AND s_quantity < 25`,
		"order_line", "stock",
	),
	newTPCCWorkload(
		"warehouse_by_id",
		`SELECT w_name, w_tax, w_ytd
		 FROM warehouse
		 WHERE w_id = 42`,
		"warehouse",
	),
	newTPCCWorkload(
		"warehouse_id_range",
		`SELECT w_id, w_name
		 FROM warehouse
		 WHERE w_id BETWEEN 40 AND 45
		 ORDER BY w_id`,
		"warehouse",
	),
	newTPCCWorkload(
		"warehouse_id_list",
		`SELECT w_id, w_city, w_state
		 FROM warehouse
		 WHERE w_id IN (1, 25, 50, 75, 100)`,
		"warehouse",
	),
	newTPCCWorkload(
		"district_by_id",
		`SELECT d_name, d_tax, d_ytd, d_next_o_id
		 FROM district
		 WHERE d_w_id = 42 AND d_id = 3`,
		"district",
	),
	newTPCCWorkload(
		"districts_for_warehouse",
		`SELECT d_id, d_name, d_next_o_id
		 FROM district
		 WHERE d_w_id = 42
		 ORDER BY d_id`,
		"district",
	),
	newTPCCWorkload(
		"district_id_range",
		`SELECT d_id, d_tax
		 FROM district
		 WHERE d_w_id = 42 AND d_id BETWEEN 3 AND 7
		 ORDER BY d_id DESC`,
		"district",
	),
	newTPCCWorkload(
		"district_join_warehouse",
		`SELECT d.d_id, d.d_name, w.w_name
		 FROM district AS d
		 JOIN warehouse AS w ON w.w_id = d.d_w_id
		 WHERE d.d_w_id = 42 AND d.d_id = 3`,
		"d", "w",
	),
	newTPCCWorkload(
		"customer_by_id",
		`SELECT c_first, c_middle, c_last, c_balance
		 FROM customer
		 WHERE c_w_id = 42 AND c_d_id = 1 AND c_id = 707`,
		"customer",
	),
	newTPCCWorkload(
		"customer_id_range",
		`SELECT c_id, c_first, c_last
		 FROM customer
		 WHERE c_w_id = 42 AND c_d_id = 1 AND c_id BETWEEN 700 AND 710
		 ORDER BY c_id`,
		"customer",
	),
	newTPCCWorkload(
		"customer_last_count",
		`SELECT COUNT(c_id) AS customer_count
		 FROM customer
		 WHERE c_w_id = 42 AND c_d_id = 1 AND c_last = 'LAST007'`,
		"customer",
	),
	newTPCCWorkload(
		"customer_last_limit",
		`SELECT c_id, c_first, c_balance
		 FROM customer
		 WHERE c_w_id = 42 AND c_d_id = 1 AND c_last = 'LAST007'
		 ORDER BY c_first
		 LIMIT 5`,
		"customer",
	),
	newTPCCWorkload(
		"customer_credit_by_id",
		`SELECT c_credit, c_credit_lim, c_discount
		 FROM customer
		 WHERE c_w_id = 42 AND c_d_id = 1 AND c_id = 707`,
		"customer",
	),
	newTPCCWorkload(
		"customer_join_district",
		`SELECT c.c_id, c.c_last, d.d_name, d.d_tax
		 FROM customer AS c
		 JOIN district AS d ON d.d_w_id = c.c_w_id AND d.d_id = c.c_d_id
		 WHERE c.c_w_id = 42 AND c.c_d_id = 1 AND c.c_id = 707`,
		"c", "d",
	),
	newTPCCWorkload(
		"customer_join_warehouse",
		`SELECT c.c_id, c.c_balance, w.w_name, w.w_tax
		 FROM customer AS c
		 JOIN warehouse AS w ON w.w_id = c.c_w_id
		 WHERE c.c_w_id = 42 AND c.c_d_id = 1 AND c.c_id = 707`,
		"c", "w",
	),
	newTPCCWorkload(
		"order_by_id",
		`SELECT o_c_id, o_entry_d, o_carrier_id, o_ol_cnt
		 FROM orders
		 WHERE o_w_id = 42 AND o_d_id = 1 AND o_id = 707`,
		"orders",
	),
	newTPCCWorkload(
		"orders_customer_range",
		`SELECT o_id, o_entry_d, o_carrier_id
		 FROM orders
		 WHERE o_w_id = 42 AND o_d_id = 1 AND o_c_id = 707
		   AND o_id BETWEEN 700 AND 720
		 ORDER BY o_id`,
		"orders",
	),
	newTPCCWorkload(
		"orders_customer_count",
		`SELECT COUNT(*) AS order_count
		 FROM orders
		 WHERE o_w_id = 42 AND o_d_id = 1 AND o_c_id = 707`,
		"orders",
	),
	newTPCCWorkload(
		"orders_recent_range",
		`SELECT o_id, o_c_id, o_ol_cnt
		 FROM orders
		 WHERE o_w_id = 42 AND o_d_id = 1 AND o_id >= 990
		 ORDER BY o_id DESC`,
		"orders",
	),
	newTPCCWorkload(
		"order_join_customer",
		`SELECT o.o_id, o.o_entry_d, c.c_first, c.c_last
		 FROM orders AS o
		 JOIN customer AS c
		   ON c.c_w_id = o.o_w_id AND c.c_d_id = o.o_d_id AND c.c_id = o.o_c_id
		 WHERE o.o_w_id = 42 AND o.o_d_id = 1 AND o.o_id = 707`,
		"o", "c",
	),
	newTPCCWorkload(
		"new_order_by_id",
		`SELECT no_o_id, no_d_id, no_w_id
		 FROM new_order
		 WHERE no_w_id = 42 AND no_d_id = 1 AND no_o_id = 907`,
		"new_order",
	),
	newTPCCWorkload(
		"new_order_first",
		`SELECT no_o_id
		 FROM new_order
		 WHERE no_w_id = 42 AND no_d_id = 1
		 ORDER BY no_o_id
		 LIMIT 1`,
		"new_order",
	),
	newTPCCWorkload(
		"new_order_count",
		`SELECT COUNT(*) AS new_order_count
		 FROM new_order
		 WHERE no_w_id = 42 AND no_d_id = 1 AND no_o_id >= 900`,
		"new_order",
	),
	newTPCCWorkload(
		"order_line_by_id",
		`SELECT ol_i_id, ol_quantity, ol_amount
		 FROM order_line
		 WHERE ol_w_id = 42 AND ol_d_id = 1 AND ol_o_id = 907 AND ol_number = 3`,
		"order_line",
	),
	newTPCCWorkload(
		"order_line_order_sum",
		`SELECT SUM(ol_amount) AS order_amount
		 FROM order_line
		 WHERE ol_w_id = 42 AND ol_d_id = 1 AND ol_o_id = 907`,
		"order_line",
	),
	newTPCCWorkload(
		"order_line_order_range",
		`SELECT ol_o_id, ol_number, ol_i_id
		 FROM order_line
		 WHERE ol_w_id = 42 AND ol_d_id = 1
		   AND ol_o_id BETWEEN 900 AND 905 AND ol_number = 1
		 ORDER BY ol_o_id`,
		"order_line",
	),
	newTPCCWorkload(
		"order_line_join_orders",
		`SELECT ol.ol_number, ol.ol_i_id, o.o_c_id
		 FROM order_line AS ol
		 JOIN orders AS o
		   ON o.o_w_id = ol.ol_w_id AND o.o_d_id = ol.ol_d_id AND o.o_id = ol.ol_o_id
		 WHERE ol.ol_w_id = 42 AND ol.ol_d_id = 1 AND ol.ol_o_id = 907`,
		"ol", "o",
	),
	newTPCCWorkload(
		"item_by_id",
		`SELECT i_name, i_price, i_data
		 FROM item
		 WHERE i_id = 707`,
		"item",
	),
	newTPCCWorkload(
		"item_id_range",
		`SELECT i_id, i_name, i_price
		 FROM item
		 WHERE i_id BETWEEN 700 AND 710
		 ORDER BY i_id`,
		"item",
	),
	newTPCCWorkload(
		"stock_by_id",
		`SELECT s_quantity, s_ytd, s_order_cnt, s_remote_cnt
		 FROM stock
		 WHERE s_w_id = 42 AND s_i_id = 707`,
		"stock",
	),
	newTPCCWorkload(
		"stock_item_range",
		`SELECT s_i_id, s_quantity
		 FROM stock
		 WHERE s_w_id = 42 AND s_i_id BETWEEN 700 AND 710
		 ORDER BY s_i_id`,
		"stock",
	),
})

func assignRandomWeights(workloads []tpccWorkload) []tpccWorkload {
	const maxWeight = 100
	if len(workloads) > maxWeight {
		panic(fmt.Sprintf("cannot assign unique weights from 1 to %d to %d TPC-C workloads", maxWeight, len(workloads)))
	}
	weights := rand.Perm(maxWeight)
	for i := range workloads {
		workloads[i].weight = weights[i] + 1
	}
	return workloads
}

func newTPCCWorkload(name, statement string, tables ...string) tpccWorkload {
	hint := tpccBindingHintUseIndex
	if rand.IntN(2) == 1 {
		hint = tpccBindingHintTiFlash
	}
	return newTPCCWorkloadWithHint(name, statement, hint, tables...)
}

func newTPCCWorkloadWithHint(name, statement string, hint tpccBindingHint, tables ...string) tpccWorkload {
	statement = strings.TrimSpace(statement)
	const selectKeyword = "SELECT"
	if len(statement) < len(selectKeyword) || !strings.EqualFold(statement[:len(selectKeyword)], selectKeyword) {
		panic(fmt.Sprintf("TPC-C workload %q must start with SELECT", name))
	}
	if len(tables) == 0 {
		panic(fmt.Sprintf("TPC-C workload %q must reference at least one table", name))
	}
	marker := "optimizer_canary_tpcc:" + name
	tail := statement[len(selectKeyword):]
	var optimizerHint string
	switch hint {
	case tpccBindingHintUseIndex:
		hints := make([]string, 0, len(tables))
		for _, table := range tables {
			hints = append(hints, "USE_INDEX("+table+")")
		}
		optimizerHint = strings.Join(hints, " ")
	case tpccBindingHintTiFlash:
		optimizerHint = "READ_FROM_STORAGE(TIFLASH[" + strings.Join(tables, ", ") + "])"
	default:
		panic(fmt.Sprintf("TPC-C workload %q has unsupported binding hint %q", name, hint))
	}
	return tpccWorkload{
		name:        name,
		marker:      marker,
		sql:         selectKeyword + " /* " + marker + " */" + tail,
		boundSQL:    selectKeyword + " /*+ " + optimizerHint + " */ /* " + marker + " */" + tail,
		bindingHint: hint,
	}
}
