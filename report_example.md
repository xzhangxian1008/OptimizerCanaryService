# SQL Plan Comparison

| SQL Digest | ExecCount | Total ExecTime | Current Plan | New Plan | Plan Change | Binding of the Current Plan |
| --- | ---: | ---: | --- | --- | --- | --- |
| [5a59161f](#sql-5a59161f) | 1244 | 13.34s | [25da9ba3](#current-plan-5a59161f-25da9ba3) | [3599a539](#new-plan-5a59161f-3599a539) | Join Order Change | [binding stmt](#binding-stmt-5a59161f-25da9ba3) |
| [508af7b7](#sql-508af7b7) | 1285 | 13.25s | [81239958](#current-plan-508af7b7-81239958) | [81b8fd42](#new-plan-508af7b7-81b8fd42) | TiKV->TiFlash | [binding stmt](#binding-stmt-508af7b7-81239958) |
| [7ae3853d](#sql-7ae3853d) | 1051 | 11.45s | [7a3d4314](#current-plan-7ae3853d-7a3d4314) | [95cca885](#new-plan-7ae3853d-95cca885) | Index Change | [binding stmt](#binding-stmt-7ae3853d-7a3d4314) |
| [b595cdf2](#sql-b595cdf2) | 1290 | 10.78s | [aa021341](#current-plan-b595cdf2-aa021341) | [dd6eb16c](#new-plan-b595cdf2-dd6eb16c) | TiFlash->TiKV | [binding stmt](#binding-stmt-b595cdf2-aa021341) |
| [71aa781a](#sql-71aa781a) | 1049 | 9.40s | [3408b6ab](#current-plan-71aa781a-3408b6ab) | [78828f61](#new-plan-71aa781a-78828f61) | Others | [binding stmt](#binding-stmt-71aa781a-3408b6ab) |
| [5f700e0f](#sql-5f700e0f) | 863 | 8.80s | [5684c867](#current-plan-5f700e0f-5684c867) | [1ca7919d](#new-plan-5f700e0f-1ca7919d) | TiKV->TiFlash | [binding stmt](#binding-stmt-5f700e0f-5684c867) |
| [435ba825](#sql-435ba825) | 710 | 6.05s | [7d89739d](#current-plan-435ba825-7d89739d) | [02e1284d](#new-plan-435ba825-02e1284d) | Index Change | [binding stmt](#binding-stmt-435ba825-7d89739d) |
| [169b16a2](#sql-169b16a2) | 429 | 4.71s | [72b00515](#current-plan-169b16a2-72b00515) | [d99b28d7](#new-plan-169b16a2-d99b28d7) | Others | [binding stmt](#binding-stmt-169b16a2-72b00515) |
| [c874ca57](#sql-c874ca57) | 649 | 4.64s | [1e80b8c4](#current-plan-c874ca57-1e80b8c4) | [04539a74](#new-plan-c874ca57-04539a74) | TiFlash->TiKV | [binding stmt](#binding-stmt-c874ca57-1e80b8c4) |
| [c39285d1](#sql-c39285d1) | 568 | 4.51s | [4c1ca90b](#current-plan-c39285d1-4c1ca90b) | [f166f97a](#new-plan-c39285d1-f166f97a) | Join Order Change | [binding stmt](#binding-stmt-c39285d1-4c1ca90b) |
| [48a60462](#sql-48a60462) | 662 | 4.47s | [90abfd62](#current-plan-48a60462-90abfd62) | [dde0625c](#new-plan-48a60462-dde0625c) | TiKV->TiFlash | [binding stmt](#binding-stmt-48a60462-90abfd62) |
| [b2193e67](#sql-b2193e67) | 529 | 3.77s | [f7a22f73](#current-plan-b2193e67-f7a22f73) | [d6deb9ad](#new-plan-b2193e67-d6deb9ad) | Others | [binding stmt](#binding-stmt-b2193e67-f7a22f73) |
| [1e8079bc](#sql-1e8079bc) | 332 | 3.03s | [7e2639e9](#current-plan-1e8079bc-7e2639e9) | [54ed9ed5](#new-plan-1e8079bc-54ed9ed5) | Index Change | [binding stmt](#binding-stmt-1e8079bc-7e2639e9) |
| [bc78a1ac](#sql-bc78a1ac) | 1387 | 1.95s | [8eae982c](#current-plan-bc78a1ac-8eae982c) | [8eae982c](#new-plan-bc78a1ac-8eae982c) | Join Order Change | [binding stmt](#binding-stmt-bc78a1ac-8eae982c) |
| [c76e06dd](#sql-c76e06dd) | 260 | 1.80s | [7d6aacd0](#current-plan-c76e06dd-7d6aacd0) | [e97e832b](#new-plan-c76e06dd-e97e832b) | TiFlash->TiKV | [binding stmt](#binding-stmt-c76e06dd-7d6aacd0) |
| [49fc955d](#sql-49fc955d) | 1271 | 1.69s | [b612e9ab](#current-plan-49fc955d-b612e9ab) | [b612e9ab](#new-plan-49fc955d-b612e9ab) | Others | [binding stmt](#binding-stmt-49fc955d-b612e9ab) |
| [52772a82](#sql-52772a82) | 1256 | 1.63s | [5b116e13](#current-plan-52772a82-5b116e13) | [5b116e13](#new-plan-52772a82-5b116e13) | Join Order Change | [binding stmt](#binding-stmt-52772a82-5b116e13) |
| [3053011a](#sql-3053011a) | 163 | 1.54s | [ff7a44a1](#current-plan-3053011a-ff7a44a1) | [72d6501b](#new-plan-3053011a-72d6501b) | Index Change | [binding stmt](#binding-stmt-3053011a-ff7a44a1) |
| [44ab0ff9](#sql-44ab0ff9) | 1180 | 1.49s | [1e6e8634](#current-plan-44ab0ff9-1e6e8634) | [2e2d3a08](#new-plan-44ab0ff9-2e2d3a08) | TiKV->TiFlash | [binding stmt](#binding-stmt-44ab0ff9-1e6e8634) |
| [2be26e27](#sql-2be26e27) | 781 | 1.29s | [769bb3f9](#current-plan-2be26e27-769bb3f9) | [769bb3f9](#new-plan-2be26e27-769bb3f9) | Others | [binding stmt](#binding-stmt-2be26e27-769bb3f9) |
| [14d8bcdf](#sql-14d8bcdf) | 613 | 0.94s | [99616e47](#current-plan-14d8bcdf-99616e47) | [08478e33](#new-plan-14d8bcdf-08478e33) | TiFlash->TiKV | [binding stmt](#binding-stmt-14d8bcdf-99616e47) |
| [39601b01](#sql-39601b01) | 562 | 0.74s | [b965db72](#current-plan-39601b01-b965db72) | [b965db72](#new-plan-39601b01-b965db72) | Join Order Change | [binding stmt](#binding-stmt-39601b01-b965db72) |
| [1427f837](#sql-1427f837) | 148 | 0.64s | [425fc44d](#current-plan-1427f837-425fc44d) | [7c016980](#new-plan-1427f837-7c016980) | Index Change | [binding stmt](#binding-stmt-1427f837-425fc44d) |
| [10818e2a](#sql-10818e2a) | 299 | 0.61s | [737a0989](#current-plan-10818e2a-737a0989) | [1f11d658](#new-plan-10818e2a-1f11d658) | TiKV->TiFlash | [binding stmt](#binding-stmt-10818e2a-737a0989) |
| [747de7b8](#sql-747de7b8) | 92 | 0.44s | [fe087e4e](#current-plan-747de7b8-fe087e4e) | [9bf0da55](#new-plan-747de7b8-9bf0da55) | TiFlash->TiKV | [binding stmt](#binding-stmt-747de7b8-fe087e4e) |

<a id="sql-5a59161f"></a>

## SQL: 5a59161f

Schema: optimizer\_canary\_prepare  
SQL Digest: 5a59161f19de34a396a60725222d9c672ee46534847701cd4b6a2b91137e6731

```text
SELECT /* optimizer_canary_tpcc:order_line_join_orders */ ol.ol_number, ol.ol_i_id, o.o_c_id
		 FROM order_line AS ol
		 JOIN orders AS o
		   ON o.o_w_id = ol.ol_w_id AND o.o_d_id = ol.ol_d_id AND o.o_id = ol.ol_o_id
		 WHERE ol.ol_w_id = 42 AND ol.ol_d_id = 1 AND ol.ol_o_id = 907
```

<a id="current-plan-5a59161f-25da9ba3"></a>

### Current Plan: 25da9ba3

Schema: optimizer\_canary\_prepare  
SQL Digest: 5a59161f19de34a396a60725222d9c672ee46534847701cd4b6a2b91137e6731  
Plan Digest: 25da9ba3fdaa8b9036c55c8d734a2a51968f30d509fd4d4345a23e653305461f

```text
	id                       	task        	estRows	operator info                                                                                                                                                                                                                                                                            
	Projection_12            	root        	3.80   	optimizer_canary_prepare.order_line.ol_number, optimizer_canary_prepare.order_line.ol_i_id, optimizer_canary_prepare.orders.o_c_id                                                                                                                                                       
	└─HashJoin_16            	root        	3.80   	inner join, equal:[eq(optimizer_canary_prepare.orders.o_w_id, optimizer_canary_prepare.order_line.ol_w_id) eq(optimizer_canary_prepare.orders.o_d_id, optimizer_canary_prepare.order_line.ol_d_id) eq(optimizer_canary_prepare.orders.o_id, optimizer_canary_prepare.order_line.ol_o_id)]
	  ├─Point_Get_41(Build)  	root        	1      	table:orders, clustered index:PRIMARY(o_w_id, o_d_id, o_id)                                                                                                                                                                                                                              
	  └─TableReader_47(Probe)	root        	38.26  	MppVersion: 3, data:ExchangeSender_46                                                                                                                                                                                                                                                    
	    └─ExchangeSender_46  	cop[tiflash]	38.26  	ExchangeType: PassThrough                                                                                                                                                                                                                                                                
	      └─TableRangeScan_45	cop[tiflash]	38.26  	table:ol, range:[42 1 907,42 1 907], keep order:false                                                                                                                                                                                                                                    
```

<a id="new-plan-5a59161f-3599a539"></a>

### New Plan: 3599a539

Schema: optimizer\_canary\_prepare  
SQL Digest: 5a59161f19de34a396a60725222d9c672ee46534847701cd4b6a2b91137e6731  
Plan Digest: 3599a5399acde2e0799ec1e50c84c7602dc5d27fd261ba932db8821d67b3863d

```text
id	estRows	task	access object	operator info
Projection_12	3.80	root		optimizer_canary_prepare.order_line.ol_number, optimizer_canary_prepare.order_line.ol_i_id, optimizer_canary_prepare.orders.o_c_id
└─MergeJoin_17	3.80	root		inner join, left key:optimizer_canary_prepare.orders.o_w_id, optimizer_canary_prepare.orders.o_d_id, optimizer_canary_prepare.orders.o_id, right key:optimizer_canary_prepare.order_line.ol_w_id, optimizer_canary_prepare.order_line.ol_d_id, optimizer_canary_prepare.order_line.ol_o_id
  ├─TableReader_32(Build)	38.26	root		data:TableRangeScan_31
  │ └─TableRangeScan_31	38.26	cop[tikv]	table:ol	range:[42 1 907,42 1 907], keep order:true
  └─Point_Get_30(Probe)	1.00	root	table:orders, clustered index:PRIMARY(o_w_id, o_d_id, o_id)	
```

<a id="binding-stmt-5a59161f-25da9ba3"></a>

### Binding Stmt: 5a59161f_25da9ba3

Schema: optimizer\_canary\_prepare  
SQL Digest: 5a59161f19de34a396a60725222d9c672ee46534847701cd4b6a2b91137e6731  
Plan Digest: 25da9ba3fdaa8b9036c55c8d734a2a51968f30d509fd4d4345a23e653305461f

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:order_line_join_orders */ ol.ol_number, ol.ol_i_id, o.o_c_id
		 FROM order_line AS ol
		 JOIN orders AS o
		   ON o.o_w_id = ol.ol_w_id AND o.o_d_id = ol.ol_d_id AND o.o_id = ol.ol_o_id
		 WHERE ol.ol_w_id = 42 AND ol.ol_d_id = 1 AND ol.ol_o_id = 907 USING SELECT /*+ hash_join_probe(`optimizer_canary_prepare`.`ol`), read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`ol`]), read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`ol`, `optimizer_canary_prepare`.`o`]) */ /* optimizer_canary_tpcc:order_line_join_orders */ ol.ol_number, ol.ol_i_id, o.o_c_id
		 FROM order_line AS ol
		 JOIN orders AS o
		   ON o.o_w_id = ol.ol_w_id AND o.o_d_id = ol.ol_d_id AND o.o_id = ol.ol_o_id
		 WHERE ol.ol_w_id = 42 AND ol.ol_d_id = 1 AND ol.ol_o_id = 907;
```

<a id="sql-508af7b7"></a>

## SQL: 508af7b7

Schema: optimizer\_canary\_prepare  
SQL Digest: 508af7b7a21479af7d9b2bb6d03fa43cda9a4b86a3acac8e7e260d407041ec56

```text
SELECT /* optimizer_canary_tpcc:customer_last_limit */ c_id, c_first, c_balance
		 FROM customer
		 WHERE c_w_id = 42 AND c_d_id = 1 AND c_last = 'LAST007'
		 ORDER BY c_first
		 LIMIT 5
```

<a id="current-plan-508af7b7-81239958"></a>

### Current Plan: 81239958

Schema: optimizer\_canary\_prepare  
SQL Digest: 508af7b7a21479af7d9b2bb6d03fa43cda9a4b86a3acac8e7e260d407041ec56  
Plan Digest: 81239958b20033948c9e3d0a40c50e9be93c6927c006d4bdb1c2e1697741d6fe

```text
	id                         	task        	estRows	operator info                                               
	TopN_13                    	root        	5      	optimizer_canary_prepare.customer.c_first, offset:0, count:5
	└─TableReader_26           	root        	5      	MppVersion: 3, data:ExchangeSender_25                       
	  └─ExchangeSender_25      	cop[tiflash]	5      	ExchangeType: PassThrough                                   
	    └─TopN_24              	cop[tiflash]	5      	optimizer_canary_prepare.customer.c_first, offset:0, count:5
	      └─Selection_23       	cop[tiflash]	100    	eq(optimizer_canary_prepare.customer.c_last, "LAST007")     
	        └─TableRangeScan_22	cop[tiflash]	1000   	table:customer, range:[42 1,42 1], keep order:false         
```

<a id="new-plan-508af7b7-81b8fd42"></a>

### New Plan: 81b8fd42

Schema: optimizer\_canary\_prepare  
SQL Digest: 508af7b7a21479af7d9b2bb6d03fa43cda9a4b86a3acac8e7e260d407041ec56  
Plan Digest: 81b8fd42b93137f77840bb22c487871aeeb5b79609d60d6b325567fe536ce278

```text
id	estRows	task	access object	operator info
Projection_49	5.00	root		optimizer_canary_prepare.customer.c_id, optimizer_canary_prepare.customer.c_first, optimizer_canary_prepare.customer.c_balance
└─IndexLookUp_48	5.00	root		limit embedded(offset:0, count:5)
  ├─Limit_47(Build)	5.00	cop[tikv]		offset:0, count:5
  │ └─IndexRangeScan_45	5.00	cop[tikv]	table:customer, index:idx_customer(c_w_id, c_d_id, c_last, c_first)	range:[42 1 "LAST007",42 1 "LAST007"], keep order:true
  └─TableRowIDScan_46(Probe)	5.00	cop[tikv]	table:customer	keep order:false
```

<a id="binding-stmt-508af7b7-81239958"></a>

### Binding Stmt: 508af7b7_81239958

Schema: optimizer\_canary\_prepare  
SQL Digest: 508af7b7a21479af7d9b2bb6d03fa43cda9a4b86a3acac8e7e260d407041ec56  
Plan Digest: 81239958b20033948c9e3d0a40c50e9be93c6927c006d4bdb1c2e1697741d6fe

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:customer_last_limit */ c_id, c_first, c_balance
		 FROM customer
		 WHERE c_w_id = 42 AND c_d_id = 1 AND c_last = 'LAST007'
		 ORDER BY c_first
		 LIMIT 5 USING SELECT /*+ read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`customer`]) */ /* optimizer_canary_tpcc:customer_last_limit */ c_id, c_first, c_balance
		 FROM customer
		 WHERE c_w_id = 42 AND c_d_id = 1 AND c_last = 'LAST007'
		 ORDER BY c_first
		 LIMIT 5;
```

<a id="sql-7ae3853d"></a>

## SQL: 7ae3853d

Schema: optimizer\_canary\_prepare  
SQL Digest: 7ae3853d0af69d54413e6faa27f5be144bb33eeeec77dbfb6633d064f2358865

```text
SELECT /* optimizer_canary_tpcc:order_line_order_sum */ SUM(ol_amount) AS order_amount
		 FROM order_line
		 WHERE ol_w_id = 42 AND ol_d_id = 1 AND ol_o_id = 907
```

<a id="current-plan-7ae3853d-7a3d4314"></a>

### Current Plan: 7a3d4314

Schema: optimizer\_canary\_prepare  
SQL Digest: 7ae3853d0af69d54413e6faa27f5be144bb33eeeec77dbfb6633d064f2358865  
Plan Digest: 7a3d4314172b962a1b1877902fe71f4893c9efd790d4f436045041e9589f3d13

```text
	id                       	task        	estRows	operator info                                                      
	HashAgg_18               	root        	1      	funcs:sum(Column#13)->Column#12                                    
	└─TableReader_20         	root        	1      	MppVersion: 3, data:ExchangeSender_19                              
	  └─ExchangeSender_19    	cop[tiflash]	1      	ExchangeType: PassThrough                                          
	    └─HashAgg_10         	cop[tiflash]	1      	funcs:sum(optimizer_canary_prepare.order_line.ol_amount)->Column#13
	      └─TableRangeScan_17	cop[tiflash]	38.26  	table:order_line, range:[42 1 907,42 1 907], keep order:false      
```

<a id="new-plan-7ae3853d-95cca885"></a>

### New Plan: 95cca885

Schema: optimizer\_canary\_prepare  
SQL Digest: 7ae3853d0af69d54413e6faa27f5be144bb33eeeec77dbfb6633d064f2358865  
Plan Digest: 95cca885436b11e8c1242e613a7fdd2a54f734f40bc8326b8c562f389ccc1d9c

```text
id	estRows	task	access object	operator info
StreamAgg_28	1.00	root		funcs:sum(Column#15)->Column#12
└─TableReader_29	1.00	root		data:StreamAgg_11
  └─StreamAgg_11	1.00	cop[tikv]		funcs:sum(optimizer_canary_prepare.order_line.ol_amount)->Column#15
    └─TableRangeScan_26	38.26	cop[tikv]	table:order_line	range:[42 1 907,42 1 907], keep order:false
```

<a id="binding-stmt-7ae3853d-7a3d4314"></a>

### Binding Stmt: 7ae3853d_7a3d4314

Schema: optimizer\_canary\_prepare  
SQL Digest: 7ae3853d0af69d54413e6faa27f5be144bb33eeeec77dbfb6633d064f2358865  
Plan Digest: 7a3d4314172b962a1b1877902fe71f4893c9efd790d4f436045041e9589f3d13

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:order_line_order_sum */ SUM(ol_amount) AS order_amount
		 FROM order_line
		 WHERE ol_w_id = 42 AND ol_d_id = 1 AND ol_o_id = 907 USING SELECT /*+ hash_agg(@`sel_1`), read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`order_line`]) */ /* optimizer_canary_tpcc:order_line_order_sum */ SUM(ol_amount) AS order_amount
		 FROM order_line
		 WHERE ol_w_id = 42 AND ol_d_id = 1 AND ol_o_id = 907;
```

<a id="sql-b595cdf2"></a>

## SQL: b595cdf2

Schema: optimizer\_canary\_prepare  
SQL Digest: b595cdf241ce7e7a2a250059a2afba622655ca931d25f2a855766be2f29adfda

```text
SELECT /* optimizer_canary_tpcc:stock_item_range */ s_i_id, s_quantity
		 FROM stock
		 WHERE s_w_id = 42 AND s_i_id BETWEEN 700 AND 710
		 ORDER BY s_i_id
```

<a id="current-plan-b595cdf2-aa021341"></a>

### Current Plan: aa021341

Schema: optimizer\_canary\_prepare  
SQL Digest: b595cdf241ce7e7a2a250059a2afba622655ca931d25f2a855766be2f29adfda  
Plan Digest: aa02134140d706b8d28c7b476feadba340ab34d49cc1c9e17f225435d4085412

```text
	id                       	task        	estRows	operator info                                                                   
	Sort_5                   	root        	11.00  	optimizer_canary_prepare.stock.s_i_id                                           
	└─TableReader_15         	root        	11.00  	MppVersion: 3, data:ExchangeSender_14                                           
	  └─ExchangeSender_14    	cop[tiflash]	11.00  	ExchangeType: PassThrough                                                       
	    └─Projection_8       	cop[tiflash]	11.00  	optimizer_canary_prepare.stock.s_i_id, optimizer_canary_prepare.stock.s_quantity
	      └─TableRangeScan_13	cop[tiflash]	104.88 	table:stock, range:[42 700,42 710], keep order:false                            
```

<a id="new-plan-b595cdf2-dd6eb16c"></a>

### New Plan: dd6eb16c

Schema: optimizer\_canary\_prepare  
SQL Digest: b595cdf241ce7e7a2a250059a2afba622655ca931d25f2a855766be2f29adfda  
Plan Digest: dd6eb16c06c02d21bf67fac050183425b29615c07a589563e0122293d8409e04

```text
id	estRows	task	access object	operator info
TableReader_30	11.00	root		data:Projection_23
└─Projection_23	11.00	cop[tikv]		optimizer_canary_prepare.stock.s_i_id, optimizer_canary_prepare.stock.s_quantity
  └─TableRangeScan_28	104.88	cop[tikv]	table:stock	range:[42 700,42 710], keep order:true
```

<a id="binding-stmt-b595cdf2-aa021341"></a>

### Binding Stmt: b595cdf2_aa021341

Schema: optimizer\_canary\_prepare  
SQL Digest: b595cdf241ce7e7a2a250059a2afba622655ca931d25f2a855766be2f29adfda  
Plan Digest: aa02134140d706b8d28c7b476feadba340ab34d49cc1c9e17f225435d4085412

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:stock_item_range */ s_i_id, s_quantity
		 FROM stock
		 WHERE s_w_id = 42 AND s_i_id BETWEEN 700 AND 710
		 ORDER BY s_i_id USING SELECT /*+ read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`stock`]) */ /* optimizer_canary_tpcc:stock_item_range */ s_i_id, s_quantity
		 FROM stock
		 WHERE s_w_id = 42 AND s_i_id BETWEEN 700 AND 710
		 ORDER BY s_i_id;
```

<a id="sql-71aa781a"></a>

## SQL: 71aa781a

Schema: optimizer\_canary\_prepare  
SQL Digest: 71aa781a5fc2a59f7e1d1dfc15966ab797a1c4acf89ebaaab45fa7fb365f0e8f

```text
SELECT /* optimizer_canary_tpcc:orders_recent_range */ o_id, o_c_id, o_ol_cnt
		 FROM orders
		 WHERE o_w_id = 42 AND o_d_id = 1 AND o_id >= 990
		 ORDER BY o_id DESC
```

<a id="current-plan-71aa781a-3408b6ab"></a>

### Current Plan: 3408b6ab

Schema: optimizer\_canary\_prepare  
SQL Digest: 71aa781a5fc2a59f7e1d1dfc15966ab797a1c4acf89ebaaab45fa7fb365f0e8f  
Plan Digest: 3408b6abcf974f8edafc7abfe297c85b05c31400dee834296cb33774331da954

```text
	id                       	task        	estRows	operator info                                                                                                         
	Sort_5                   	root        	12     	optimizer_canary_prepare.orders.o_id:desc                                                                             
	└─TableReader_15         	root        	12     	MppVersion: 3, data:ExchangeSender_14                                                                                 
	  └─ExchangeSender_14    	cop[tiflash]	12     	ExchangeType: PassThrough                                                                                             
	    └─Projection_8       	cop[tiflash]	12     	optimizer_canary_prepare.orders.o_id, optimizer_canary_prepare.orders.o_c_id, optimizer_canary_prepare.orders.o_ol_cnt
	      └─TableRangeScan_13	cop[tiflash]	109.54 	table:orders, range:[42 1 990,42 1 +inf], keep order:false                                                            
```

<a id="new-plan-71aa781a-78828f61"></a>

### New Plan: 78828f61

Schema: optimizer\_canary\_prepare  
SQL Digest: 71aa781a5fc2a59f7e1d1dfc15966ab797a1c4acf89ebaaab45fa7fb365f0e8f  
Plan Digest: 78828f61a42993a1267d205ef8d22f123dfd7ef89a801e0d6a2f1c0c88c157dc

```text
id	estRows	task	access object	operator info
TableReader_30	12.00	root		data:Projection_23
└─Projection_23	12.00	cop[tikv]		optimizer_canary_prepare.orders.o_id, optimizer_canary_prepare.orders.o_c_id, optimizer_canary_prepare.orders.o_ol_cnt
  └─TableRangeScan_28	109.54	cop[tikv]	table:orders	range:[42 1 990,42 1 +inf], keep order:true, desc
```

<a id="binding-stmt-71aa781a-3408b6ab"></a>

### Binding Stmt: 71aa781a_3408b6ab

Schema: optimizer\_canary\_prepare  
SQL Digest: 71aa781a5fc2a59f7e1d1dfc15966ab797a1c4acf89ebaaab45fa7fb365f0e8f  
Plan Digest: 3408b6abcf974f8edafc7abfe297c85b05c31400dee834296cb33774331da954

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:orders_recent_range */ o_id, o_c_id, o_ol_cnt
		 FROM orders
		 WHERE o_w_id = 42 AND o_d_id = 1 AND o_id >= 990
		 ORDER BY o_id DESC USING SELECT /*+ read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`orders`]) */ /* optimizer_canary_tpcc:orders_recent_range */ o_id, o_c_id, o_ol_cnt
		 FROM orders
		 WHERE o_w_id = 42 AND o_d_id = 1 AND o_id >= 990
		 ORDER BY o_id DESC;
```

<a id="sql-5f700e0f"></a>

## SQL: 5f700e0f

Schema: optimizer\_canary\_prepare  
SQL Digest: 5f700e0fb4c4c2114f85654677712400b623e43f85cbe0ab4843152ced65d901

```text
SELECT /* optimizer_canary_tpcc:customer_id_range */ c_id, c_first, c_last
		 FROM customer
		 WHERE c_w_id = 42 AND c_d_id = 1 AND c_id BETWEEN 700 AND 710
		 ORDER BY c_id
```

<a id="current-plan-5f700e0f-5684c867"></a>

### Current Plan: 5684c867

Schema: optimizer\_canary\_prepare  
SQL Digest: 5f700e0fb4c4c2114f85654677712400b623e43f85cbe0ab4843152ced65d901  
Plan Digest: 5684c86795e2c13d87ac2529229b267049cecaf31c681a5073c4eb56ecbe4572

```text
	id                       	task        	estRows	operator info                                                                                                              
	Sort_5                   	root        	11.00  	optimizer_canary_prepare.customer.c_id                                                                                     
	└─TableReader_15         	root        	11.00  	MppVersion: 3, data:ExchangeSender_14                                                                                      
	  └─ExchangeSender_14    	cop[tiflash]	11.00  	ExchangeType: PassThrough                                                                                                  
	    └─Projection_8       	cop[tiflash]	11.00  	optimizer_canary_prepare.customer.c_id, optimizer_canary_prepare.customer.c_first, optimizer_canary_prepare.customer.c_last
	      └─TableRangeScan_13	cop[tiflash]	104.88 	table:customer, range:[42 1 700,42 1 710], keep order:false                                                                
```

<a id="new-plan-5f700e0f-1ca7919d"></a>

### New Plan: 1ca7919d

Schema: optimizer\_canary\_prepare  
SQL Digest: 5f700e0fb4c4c2114f85654677712400b623e43f85cbe0ab4843152ced65d901  
Plan Digest: 1ca7919db8bc8e0a06729cf7d726ca03b422539e517f308b982d2a4ce901a3ba

```text
id	estRows	task	access object	operator info
TableReader_30	11.00	root		data:Projection_23
└─Projection_23	11.00	cop[tikv]		optimizer_canary_prepare.customer.c_id, optimizer_canary_prepare.customer.c_first, optimizer_canary_prepare.customer.c_last
  └─TableRangeScan_28	104.88	cop[tikv]	table:customer	range:[42 1 700,42 1 710], keep order:true
```

<a id="binding-stmt-5f700e0f-5684c867"></a>

### Binding Stmt: 5f700e0f_5684c867

Schema: optimizer\_canary\_prepare  
SQL Digest: 5f700e0fb4c4c2114f85654677712400b623e43f85cbe0ab4843152ced65d901  
Plan Digest: 5684c86795e2c13d87ac2529229b267049cecaf31c681a5073c4eb56ecbe4572

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:customer_id_range */ c_id, c_first, c_last
		 FROM customer
		 WHERE c_w_id = 42 AND c_d_id = 1 AND c_id BETWEEN 700 AND 710
		 ORDER BY c_id USING SELECT /*+ read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`customer`]) */ /* optimizer_canary_tpcc:customer_id_range */ c_id, c_first, c_last
		 FROM customer
		 WHERE c_w_id = 42 AND c_d_id = 1 AND c_id BETWEEN 700 AND 710
		 ORDER BY c_id;
```

<a id="sql-435ba825"></a>

## SQL: 435ba825

Schema: optimizer\_canary\_prepare  
SQL Digest: 435ba825a2950da2dbfca05e4b1a265a4df2bce2e22941a493594cf2b2d56678

```text
SELECT /* optimizer_canary_tpcc:new_order_count */ COUNT(*) AS new_order_count
		 FROM new_order
		 WHERE no_w_id = 42 AND no_d_id = 1 AND no_o_id >= 900
```

<a id="current-plan-435ba825-7d89739d"></a>

### Current Plan: 7d89739d

Schema: optimizer\_canary\_prepare  
SQL Digest: 435ba825a2950da2dbfca05e4b1a265a4df2bce2e22941a493594cf2b2d56678  
Plan Digest: 7d89739d8612006df803e5dbbef614a64b5e6ee6fa477072a322098cd659ef58

```text
	id                       	task        	estRows	operator info                                                
	HashAgg_18               	root        	1      	funcs:count(Column#6)->Column#5                              
	└─TableReader_20         	root        	1      	MppVersion: 3, data:ExchangeSender_19                        
	  └─ExchangeSender_19    	cop[tiflash]	1      	ExchangeType: PassThrough                                    
	    └─HashAgg_10         	cop[tiflash]	1      	funcs:count(1)->Column#6                                     
	      └─TableRangeScan_17	cop[tiflash]	174.93 	table:new_order, range:[42 1 900,42 1 +inf], keep order:false
```

<a id="new-plan-435ba825-02e1284d"></a>

### New Plan: 02e1284d

Schema: optimizer\_canary\_prepare  
SQL Digest: 435ba825a2950da2dbfca05e4b1a265a4df2bce2e22941a493594cf2b2d56678  
Plan Digest: 02e1284d0251d8abfadf76b1da116d194dcb4a6fcae69cab82bac1d6ca50d4b2

```text
id	estRows	task	access object	operator info
StreamAgg_28	1.00	root		funcs:count(Column#8)->Column#5
└─TableReader_29	1.00	root		data:StreamAgg_11
  └─StreamAgg_11	1.00	cop[tikv]		funcs:count(1)->Column#8
    └─TableRangeScan_26	174.93	cop[tikv]	table:new_order	range:[42 1 900,42 1 +inf], keep order:false
```

<a id="binding-stmt-435ba825-7d89739d"></a>

### Binding Stmt: 435ba825_7d89739d

Schema: optimizer\_canary\_prepare  
SQL Digest: 435ba825a2950da2dbfca05e4b1a265a4df2bce2e22941a493594cf2b2d56678  
Plan Digest: 7d89739d8612006df803e5dbbef614a64b5e6ee6fa477072a322098cd659ef58

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:new_order_count */ COUNT(*) AS new_order_count
		 FROM new_order
		 WHERE no_w_id = 42 AND no_d_id = 1 AND no_o_id >= 900 USING SELECT /*+ hash_agg(@`sel_1`), read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`new_order`]) */ /* optimizer_canary_tpcc:new_order_count */ COUNT(*) AS new_order_count
		 FROM new_order
		 WHERE no_w_id = 42 AND no_d_id = 1 AND no_o_id >= 900;
```

<a id="sql-169b16a2"></a>

## SQL: 169b16a2

Schema: optimizer\_canary\_prepare  
SQL Digest: 169b16a2070436601c1b2dba54a93e762ad2317430896eb1c2bf58e41d8d8808

```text
SELECT /* optimizer_canary_tpcc:order_line_order_range */ ol_o_id, ol_number, ol_i_id
		 FROM order_line
		 WHERE ol_w_id = 42 AND ol_d_id = 1
		   AND ol_o_id BETWEEN 900 AND 905 AND ol_number = 1
		 ORDER BY ol_o_id
```

<a id="current-plan-169b16a2-72b00515"></a>

### Current Plan: 72b00515

Schema: optimizer\_canary\_prepare  
SQL Digest: 169b16a2070436601c1b2dba54a93e762ad2317430896eb1c2bf58e41d8d8808  
Plan Digest: 72b00515e3924043c1cc67a6ef0a6e6c88ceddc8b4b70afbfed3a1e7a7778e51

```text
	id                         	task        	estRows	operator info                                                                                                                          
	Sort_5                     	root        	5.85   	optimizer_canary_prepare.order_line.ol_o_id                                                                                            
	└─TableReader_17           	root        	5.85   	MppVersion: 3, data:ExchangeSender_16                                                                                                  
	  └─ExchangeSender_16      	cop[tiflash]	5.85   	ExchangeType: PassThrough                                                                                                              
	    └─Projection_8         	cop[tiflash]	5.85   	optimizer_canary_prepare.order_line.ol_o_id, optimizer_canary_prepare.order_line.ol_number, optimizer_canary_prepare.order_line.ol_i_id
	      └─Selection_15       	cop[tiflash]	5.85   	eq(optimizer_canary_prepare.order_line.ol_number, 1)                                                                                   
	        └─TableRangeScan_14	cop[tiflash]	294.35 	table:order_line, range:[42 1 900,42 1 905], keep order:false                                                                          
```

<a id="new-plan-169b16a2-d99b28d7"></a>

### New Plan: d99b28d7

Schema: optimizer\_canary\_prepare  
SQL Digest: 169b16a2070436601c1b2dba54a93e762ad2317430896eb1c2bf58e41d8d8808  
Plan Digest: d99b28d7ca4cd6a41a7a5b886ac32718c1ffd0b7861b65ce2a56401a2cadcd5f

```text
id	estRows	task	access object	operator info
TableReader_36	5.85	root		data:Projection_27
└─Projection_27	5.85	cop[tikv]		optimizer_canary_prepare.order_line.ol_o_id, optimizer_canary_prepare.order_line.ol_number, optimizer_canary_prepare.order_line.ol_i_id
  └─Selection_34	5.85	cop[tikv]		eq(optimizer_canary_prepare.order_line.ol_number, 1)
    └─TableRangeScan_33	294.35	cop[tikv]	table:order_line	range:[42 1 900,42 1 905], keep order:true
```

<a id="binding-stmt-169b16a2-72b00515"></a>

### Binding Stmt: 169b16a2_72b00515

Schema: optimizer\_canary\_prepare  
SQL Digest: 169b16a2070436601c1b2dba54a93e762ad2317430896eb1c2bf58e41d8d8808  
Plan Digest: 72b00515e3924043c1cc67a6ef0a6e6c88ceddc8b4b70afbfed3a1e7a7778e51

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:order_line_order_range */ ol_o_id, ol_number, ol_i_id
		 FROM order_line
		 WHERE ol_w_id = 42 AND ol_d_id = 1
		   AND ol_o_id BETWEEN 900 AND 905 AND ol_number = 1
		 ORDER BY ol_o_id USING SELECT /*+ read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`order_line`]) */ /* optimizer_canary_tpcc:order_line_order_range */ ol_o_id, ol_number, ol_i_id
		 FROM order_line
		 WHERE ol_w_id = 42 AND ol_d_id = 1
		   AND ol_o_id BETWEEN 900 AND 905 AND ol_number = 1
		 ORDER BY ol_o_id;
```

<a id="sql-c874ca57"></a>

## SQL: c874ca57

Schema: optimizer\_canary\_prepare  
SQL Digest: c874ca5737aa0a0152a5c89f517ca94d6802a6c1e0627b6941716054f608e971

```text
SELECT /* optimizer_canary_tpcc:order_status_order_lines */ ol_i_id, ol_supply_w_id, ol_quantity, ol_amount, ol_delivery_d
		 FROM order_line
		 WHERE ol_w_id = 1 AND ol_d_id = 1 AND ol_o_id = 907
```

<a id="current-plan-c874ca57-1e80b8c4"></a>

### Current Plan: 1e80b8c4

Schema: optimizer\_canary\_prepare  
SQL Digest: c874ca5737aa0a0152a5c89f517ca94d6802a6c1e0627b6941716054f608e971  
Plan Digest: 1e80b8c419a29e6a5949851e9a094690142cfe2e5c2aaf89b0029ad8a9a8af62

```text
	id                     	task        	estRows	operator info                                                                                                                                                                                                                                     
	TableReader_12         	root        	3.81   	MppVersion: 3, data:ExchangeSender_11                                                                                                                                                                                                             
	└─ExchangeSender_11    	cop[tiflash]	3.81   	ExchangeType: PassThrough                                                                                                                                                                                                                         
	  └─Projection_5       	cop[tiflash]	3.81   	optimizer_canary_prepare.order_line.ol_i_id, optimizer_canary_prepare.order_line.ol_supply_w_id, optimizer_canary_prepare.order_line.ol_quantity, optimizer_canary_prepare.order_line.ol_amount, optimizer_canary_prepare.order_line.ol_delivery_d
	    └─TableRangeScan_10	cop[tiflash]	38.28  	table:order_line, range:[1 1 907,1 1 907], keep order:false                                                                                                                                                                                       
```

<a id="new-plan-c874ca57-04539a74"></a>

### New Plan: 04539a74

Schema: optimizer\_canary\_prepare  
SQL Digest: c874ca5737aa0a0152a5c89f517ca94d6802a6c1e0627b6941716054f608e971  
Plan Digest: 04539a74c29ecda6fc4c0b170e7ffc7bba71e9821990c0293831d09ddb67dfff

```text
id	estRows	task	access object	operator info
TableReader_17	3.81	root		data:Projection_6
└─Projection_6	3.81	cop[tikv]		optimizer_canary_prepare.order_line.ol_i_id, optimizer_canary_prepare.order_line.ol_supply_w_id, optimizer_canary_prepare.order_line.ol_quantity, optimizer_canary_prepare.order_line.ol_amount, optimizer_canary_prepare.order_line.ol_delivery_d
  └─TableRangeScan_15	38.28	cop[tikv]	table:order_line	range:[1 1 907,1 1 907], keep order:false
```

<a id="binding-stmt-c874ca57-1e80b8c4"></a>

### Binding Stmt: c874ca57_1e80b8c4

Schema: optimizer\_canary\_prepare  
SQL Digest: c874ca5737aa0a0152a5c89f517ca94d6802a6c1e0627b6941716054f608e971  
Plan Digest: 1e80b8c419a29e6a5949851e9a094690142cfe2e5c2aaf89b0029ad8a9a8af62

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:order_status_order_lines */ ol_i_id, ol_supply_w_id, ol_quantity, ol_amount, ol_delivery_d
		 FROM order_line
		 WHERE ol_w_id = 1 AND ol_d_id = 1 AND ol_o_id = 907 USING SELECT /*+ read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`order_line`]) */ /* optimizer_canary_tpcc:order_status_order_lines */ ol_i_id, ol_supply_w_id, ol_quantity, ol_amount, ol_delivery_d
		 FROM order_line
		 WHERE ol_w_id = 1 AND ol_d_id = 1 AND ol_o_id = 907;
```

<a id="sql-c39285d1"></a>

## SQL: c39285d1

Schema: optimizer\_canary\_prepare  
SQL Digest: c39285d1e3ef1c00af6873649253e8882288e878c745613033f84d79e0322c2d

```text
SELECT /* optimizer_canary_tpcc:new_order_first */ no_o_id
		 FROM new_order
		 WHERE no_w_id = 42 AND no_d_id = 1
		 ORDER BY no_o_id
		 LIMIT 1
```

<a id="current-plan-c39285d1-4c1ca90b"></a>

### Current Plan: 4c1ca90b

Schema: optimizer\_canary\_prepare  
SQL Digest: c39285d1e3ef1c00af6873649253e8882288e878c745613033f84d79e0322c2d  
Plan Digest: 4c1ca90be602c152dd2cdc725e0e800d2f04fa07b78965f3febd518385f93335

```text
	id                       	task        	estRows	operator info                                                
	TopN_13                  	root        	1      	optimizer_canary_prepare.new_order.no_o_id, offset:0, count:1
	└─TableReader_24         	root        	1      	MppVersion: 3, data:ExchangeSender_23                        
	  └─ExchangeSender_23    	cop[tiflash]	1      	ExchangeType: PassThrough                                    
	    └─TopN_22            	cop[tiflash]	1      	optimizer_canary_prepare.new_order.no_o_id, offset:0, count:1
	      └─TableRangeScan_21	cop[tiflash]	300    	table:new_order, range:[42 1,42 1], keep order:false         
```

<a id="new-plan-c39285d1-f166f97a"></a>

### New Plan: f166f97a

Schema: optimizer\_canary\_prepare  
SQL Digest: c39285d1e3ef1c00af6873649253e8882288e878c745613033f84d79e0322c2d  
Plan Digest: f166f97a9f4312fa300b1eff175f624d9a94b733373c4db9f13c6c00f79d5c8a

```text
id	estRows	task	access object	operator info
Limit_14	1.00	root		offset:0, count:1
└─TableReader_33	1.00	root		data:Limit_32
  └─Limit_32	1.00	cop[tikv]		offset:0, count:1
    └─TableRangeScan_30	3.99	cop[tikv]	table:new_order	range:[42 1,42 1], keep order:true
```

<a id="binding-stmt-c39285d1-4c1ca90b"></a>

### Binding Stmt: c39285d1_4c1ca90b

Schema: optimizer\_canary\_prepare  
SQL Digest: c39285d1e3ef1c00af6873649253e8882288e878c745613033f84d79e0322c2d  
Plan Digest: 4c1ca90be602c152dd2cdc725e0e800d2f04fa07b78965f3febd518385f93335

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:new_order_first */ no_o_id
		 FROM new_order
		 WHERE no_w_id = 42 AND no_d_id = 1
		 ORDER BY no_o_id
		 LIMIT 1 USING SELECT /*+ read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`new_order`]) */ /* optimizer_canary_tpcc:new_order_first */ no_o_id
		 FROM new_order
		 WHERE no_w_id = 42 AND no_d_id = 1
		 ORDER BY no_o_id
		 LIMIT 1;
```

<a id="sql-48a60462"></a>

## SQL: 48a60462

Schema: optimizer\_canary\_prepare  
SQL Digest: 48a60462829f33d45d865d105a8bca1b36763105f8049fec0604dd71b0a6dcd7

```text
SELECT /* optimizer_canary_tpcc:districts_for_warehouse */ d_id, d_name, d_next_o_id
		 FROM district
		 WHERE d_w_id = 42
		 ORDER BY d_id
```

<a id="current-plan-48a60462-90abfd62"></a>

### Current Plan: 90abfd62

Schema: optimizer\_canary\_prepare  
SQL Digest: 48a60462829f33d45d865d105a8bca1b36763105f8049fec0604dd71b0a6dcd7  
Plan Digest: 90abfd62b2768a384c1d5b14b9ee9ef4a6f00fbce301ea66e3034c1cbb3bb26c

```text
	id                       	task        	estRows	operator info                                                                                                                  
	Sort_5                   	root        	10     	optimizer_canary_prepare.district.d_id                                                                                         
	└─TableReader_15         	root        	10     	MppVersion: 3, data:ExchangeSender_14                                                                                          
	  └─ExchangeSender_14    	cop[tiflash]	10     	ExchangeType: PassThrough                                                                                                      
	    └─Projection_8       	cop[tiflash]	10     	optimizer_canary_prepare.district.d_id, optimizer_canary_prepare.district.d_name, optimizer_canary_prepare.district.d_next_o_id
	      └─TableRangeScan_13	cop[tiflash]	10     	table:district, range:[42,42], keep order:false                                                                                
```

<a id="new-plan-48a60462-dde0625c"></a>

### New Plan: dde0625c

Schema: optimizer\_canary\_prepare  
SQL Digest: 48a60462829f33d45d865d105a8bca1b36763105f8049fec0604dd71b0a6dcd7  
Plan Digest: dde0625c47aedab6efd71bb02c7048216629acb3dbbdc317e86bed9f4719b56d

```text
id	estRows	task	access object	operator info
TableReader_30	10.00	root		data:Projection_23
└─Projection_23	10.00	cop[tikv]		optimizer_canary_prepare.district.d_id, optimizer_canary_prepare.district.d_name, optimizer_canary_prepare.district.d_next_o_id
  └─TableRangeScan_28	10.00	cop[tikv]	table:district	range:[42,42], keep order:true
```

<a id="binding-stmt-48a60462-90abfd62"></a>

### Binding Stmt: 48a60462_90abfd62

Schema: optimizer\_canary\_prepare  
SQL Digest: 48a60462829f33d45d865d105a8bca1b36763105f8049fec0604dd71b0a6dcd7  
Plan Digest: 90abfd62b2768a384c1d5b14b9ee9ef4a6f00fbce301ea66e3034c1cbb3bb26c

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:districts_for_warehouse */ d_id, d_name, d_next_o_id
		 FROM district
		 WHERE d_w_id = 42
		 ORDER BY d_id USING SELECT /*+ read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`district`]) */ /* optimizer_canary_tpcc:districts_for_warehouse */ d_id, d_name, d_next_o_id
		 FROM district
		 WHERE d_w_id = 42
		 ORDER BY d_id;
```

<a id="sql-b2193e67"></a>

## SQL: b2193e67

Schema: optimizer\_canary\_prepare  
SQL Digest: b2193e672c7f58fdc9ce3900c7ac8cd5fefc58b2e56320ae9f62e09ad73e0e36

```text
SELECT /* optimizer_canary_tpcc:order_status_customer_by_last */ c_balance, c_first, c_middle, c_id
		 FROM customer
		 WHERE c_w_id = 1 AND c_d_id = 1 AND c_last = 'LAST007'
		 ORDER BY c_first
```

<a id="current-plan-b2193e67-f7a22f73"></a>

### Current Plan: f7a22f73

Schema: optimizer\_canary\_prepare  
SQL Digest: b2193e672c7f58fdc9ce3900c7ac8cd5fefc58b2e56320ae9f62e09ad73e0e36  
Plan Digest: f7a22f73b58ed092c434faa121696cd7a7c010bb4aab13973c8ecf2edd30491e

```text
	id                         	task        	estRows	operator info                                                                                                                                                             
	Sort_5                     	root        	100    	optimizer_canary_prepare.customer.c_first                                                                                                                                 
	└─TableReader_17           	root        	100    	MppVersion: 3, data:ExchangeSender_16                                                                                                                                     
	  └─ExchangeSender_16      	cop[tiflash]	100    	ExchangeType: PassThrough                                                                                                                                                 
	    └─Projection_8         	cop[tiflash]	100    	optimizer_canary_prepare.customer.c_balance, optimizer_canary_prepare.customer.c_first, optimizer_canary_prepare.customer.c_middle, optimizer_canary_prepare.customer.c_id
	      └─Selection_15       	cop[tiflash]	100    	eq(optimizer_canary_prepare.customer.c_last, "LAST007")                                                                                                                   
	        └─TableRangeScan_14	cop[tiflash]	1000   	table:customer, range:[1 1,1 1], keep order:false                                                                                                                         
```

<a id="new-plan-b2193e67-d6deb9ad"></a>

### New Plan: d6deb9ad

Schema: optimizer\_canary\_prepare  
SQL Digest: b2193e672c7f58fdc9ce3900c7ac8cd5fefc58b2e56320ae9f62e09ad73e0e36  
Plan Digest: d6deb9ad0f3ef13c952d75ec174fa00c40ef4a655f787b01559f8c7cb0ae8fda

```text
id	estRows	task	access object	operator info
Sort_5	100.00	root		optimizer_canary_prepare.customer.c_first
└─TableReader_27	100.00	root		data:Projection_9
  └─Projection_9	100.00	cop[tikv]		optimizer_canary_prepare.customer.c_balance, optimizer_canary_prepare.customer.c_first, optimizer_canary_prepare.customer.c_middle, optimizer_canary_prepare.customer.c_id
    └─Selection_25	100.00	cop[tikv]		eq(optimizer_canary_prepare.customer.c_last, "LAST007")
      └─TableRangeScan_24	1000.00	cop[tikv]	table:customer	range:[1 1,1 1], keep order:false
```

<a id="binding-stmt-b2193e67-f7a22f73"></a>

### Binding Stmt: b2193e67_f7a22f73

Schema: optimizer\_canary\_prepare  
SQL Digest: b2193e672c7f58fdc9ce3900c7ac8cd5fefc58b2e56320ae9f62e09ad73e0e36  
Plan Digest: f7a22f73b58ed092c434faa121696cd7a7c010bb4aab13973c8ecf2edd30491e

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:order_status_customer_by_last */ c_balance, c_first, c_middle, c_id
		 FROM customer
		 WHERE c_w_id = 1 AND c_d_id = 1 AND c_last = 'LAST007'
		 ORDER BY c_first USING SELECT /*+ read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`customer`]) */ /* optimizer_canary_tpcc:order_status_customer_by_last */ c_balance, c_first, c_middle, c_id
		 FROM customer
		 WHERE c_w_id = 1 AND c_d_id = 1 AND c_last = 'LAST007'
		 ORDER BY c_first;
```

<a id="sql-1e8079bc"></a>

## SQL: 1e8079bc

Schema: optimizer\_canary\_prepare  
SQL Digest: 1e8079bca3ff212059cbd8002c5ed5b8b389a5bffee2d81ee71ca2f95bfa37ff

```text
SELECT /* optimizer_canary_tpcc:orders_customer_count */ COUNT(*) AS order_count
		 FROM orders
		 WHERE o_w_id = 42 AND o_d_id = 1 AND o_c_id = 707
```

<a id="current-plan-1e8079bc-7e2639e9"></a>

### Current Plan: 7e2639e9

Schema: optimizer\_canary\_prepare  
SQL Digest: 1e8079bca3ff212059cbd8002c5ed5b8b389a5bffee2d81ee71ca2f95bfa37ff  
Plan Digest: 7e2639e9b5e8c8ab5c195062251f3377ab4e3f902c21588bd55ec282ab207260

```text
	id                       	task        	estRows	operator info                                    
	StreamAgg_12             	root        	1      	funcs:count(1)->Column#10                        
	└─TableReader_27         	root        	10     	MppVersion: 3, data:ExchangeSender_26            
	  └─ExchangeSender_26    	cop[tiflash]	10     	ExchangeType: PassThrough                        
	    └─Selection_25       	cop[tiflash]	10     	eq(optimizer_canary_prepare.orders.o_c_id, 707)  
	      └─TableRangeScan_24	cop[tiflash]	1000   	table:orders, range:[42 1,42 1], keep order:false
```

<a id="new-plan-1e8079bc-54ed9ed5"></a>

### New Plan: 54ed9ed5

Schema: optimizer\_canary\_prepare  
SQL Digest: 1e8079bca3ff212059cbd8002c5ed5b8b389a5bffee2d81ee71ca2f95bfa37ff  
Plan Digest: 54ed9ed5b424168bd7a66f36e5964691bc6af5f2d9c0dfd9a66d40697300a7f0

```text
id	estRows	task	access object	operator info
StreamAgg_30	1.00	root		funcs:count(Column#16)->Column#10
└─IndexReader_31	1.00	root		index:StreamAgg_11
  └─StreamAgg_11	1.00	cop[tikv]		funcs:count(1)->Column#16
    └─IndexRangeScan_29	10.00	cop[tikv]	table:orders, index:idx_order(o_w_id, o_d_id, o_c_id, o_id)	range:[42 1 707,42 1 707], keep order:false
```

<a id="binding-stmt-1e8079bc-7e2639e9"></a>

### Binding Stmt: 1e8079bc_7e2639e9

Schema: optimizer\_canary\_prepare  
SQL Digest: 1e8079bca3ff212059cbd8002c5ed5b8b389a5bffee2d81ee71ca2f95bfa37ff  
Plan Digest: 7e2639e9b5e8c8ab5c195062251f3377ab4e3f902c21588bd55ec282ab207260

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:orders_customer_count */ COUNT(*) AS order_count
		 FROM orders
		 WHERE o_w_id = 42 AND o_d_id = 1 AND o_c_id = 707 USING SELECT /*+ stream_agg(@`sel_1`), read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`orders`]) */ /* optimizer_canary_tpcc:orders_customer_count */ COUNT(*) AS order_count
		 FROM orders
		 WHERE o_w_id = 42 AND o_d_id = 1 AND o_c_id = 707;
```

<a id="sql-bc78a1ac"></a>

## SQL: bc78a1ac

Schema: optimizer\_canary\_prepare  
SQL Digest: bc78a1ac00af1e512619306025a7f9f690d0cf44a8ee735f4c85f1a8c144e6a1

```text
SELECT /* optimizer_canary_tpcc:customer_join_district */ c.c_id, c.c_last, d.d_name, d.d_tax
		 FROM customer AS c
		 JOIN district AS d ON d.d_w_id = c.c_w_id AND d.d_id = c.c_d_id
		 WHERE c.c_w_id = 42 AND c.c_d_id = 1 AND c.c_id = 707
```

<a id="current-plan-bc78a1ac-8eae982c"></a>

### Current Plan: 8eae982c

Schema: optimizer\_canary\_prepare  
SQL Digest: bc78a1ac00af1e512619306025a7f9f690d0cf44a8ee735f4c85f1a8c144e6a1  
Plan Digest: 8eae982cc1437c6ed94c86294a145bc4da3900b216f322a018cce3330e4fda9f

```text
	id                   	task	estRows	operator info                                                                                                                                                                                      
	MergeJoin_14         	root	1      	inner join, left key:optimizer_canary_prepare.customer.c_w_id, optimizer_canary_prepare.customer.c_d_id, right key:optimizer_canary_prepare.district.d_w_id, optimizer_canary_prepare.district.d_id
	├─Point_Get_28(Build)	root	1      	table:district, clustered index:PRIMARY(d_w_id, d_id)                                                                                                                                              
	└─Point_Get_27(Probe)	root	1      	table:customer, clustered index:PRIMARY(c_w_id, c_d_id, c_id)                                                                                                                                      
```

<a id="new-plan-bc78a1ac-8eae982c"></a>

### New Plan: 8eae982c

Schema: optimizer\_canary\_prepare  
SQL Digest: bc78a1ac00af1e512619306025a7f9f690d0cf44a8ee735f4c85f1a8c144e6a1  
Plan Digest: 8eae982cc1437c6ed94c86294a145bc4da3900b216f322a018cce3330e4fda9f

```text
id	estRows	task	access object	operator info
MergeJoin_14	1.00	root		inner join, left key:optimizer_canary_prepare.customer.c_w_id, optimizer_canary_prepare.customer.c_d_id, right key:optimizer_canary_prepare.district.d_w_id, optimizer_canary_prepare.district.d_id
├─Point_Get_29(Build)	1.00	root	table:district, clustered index:PRIMARY(d_w_id, d_id)	
└─Point_Get_28(Probe)	1.00	root	table:customer, clustered index:PRIMARY(c_w_id, c_d_id, c_id)	
```

<a id="binding-stmt-bc78a1ac-8eae982c"></a>

### Binding Stmt: bc78a1ac_8eae982c

Schema: optimizer\_canary\_prepare  
SQL Digest: bc78a1ac00af1e512619306025a7f9f690d0cf44a8ee735f4c85f1a8c144e6a1  
Plan Digest: 8eae982cc1437c6ed94c86294a145bc4da3900b216f322a018cce3330e4fda9f

```text
No PLAN_HINT available.
```

<a id="sql-c76e06dd"></a>

## SQL: c76e06dd

Schema: optimizer\_canary\_prepare  
SQL Digest: c76e06ddb23dfd3a1f2be4df70761fa4d7d1dcbe7ceae2485510bd28293ca712

```text
SELECT /* optimizer_canary_tpcc:district_id_range */ d_id, d_tax
		 FROM district
		 WHERE d_w_id = 42 AND d_id BETWEEN 3 AND 7
		 ORDER BY d_id DESC
```

<a id="current-plan-c76e06dd-7d6aacd0"></a>

### Current Plan: 7d6aacd0

Schema: optimizer\_canary\_prepare  
SQL Digest: c76e06ddb23dfd3a1f2be4df70761fa4d7d1dcbe7ceae2485510bd28293ca712  
Plan Digest: 7d6aacd065ec51d4de813171a8132ec986bb691bc3bddfd96cc3f13e254d47d5

```text
	id                       	task        	estRows	operator info                                                                  
	Sort_5                   	root        	5      	optimizer_canary_prepare.district.d_id:desc                                    
	└─TableReader_15         	root        	5      	MppVersion: 3, data:ExchangeSender_14                                          
	  └─ExchangeSender_14    	cop[tiflash]	5      	ExchangeType: PassThrough                                                      
	    └─Projection_8       	cop[tiflash]	5      	optimizer_canary_prepare.district.d_id, optimizer_canary_prepare.district.d_tax
	      └─TableRangeScan_13	cop[tiflash]	7.07   	table:district, range:[42 3,42 7], keep order:false                            
```

<a id="new-plan-c76e06dd-e97e832b"></a>

### New Plan: e97e832b

Schema: optimizer\_canary\_prepare  
SQL Digest: c76e06ddb23dfd3a1f2be4df70761fa4d7d1dcbe7ceae2485510bd28293ca712  
Plan Digest: e97e832bda126163c1dcce54ace6fcc4dbf73463ad89fdfcf0d8f31c1409b60c

```text
id	estRows	task	access object	operator info
TableReader_30	5.00	root		data:Projection_23
└─Projection_23	5.00	cop[tikv]		optimizer_canary_prepare.district.d_id, optimizer_canary_prepare.district.d_tax
  └─TableRangeScan_28	7.07	cop[tikv]	table:district	range:[42 3,42 7], keep order:true, desc
```

<a id="binding-stmt-c76e06dd-7d6aacd0"></a>

### Binding Stmt: c76e06dd_7d6aacd0

Schema: optimizer\_canary\_prepare  
SQL Digest: c76e06ddb23dfd3a1f2be4df70761fa4d7d1dcbe7ceae2485510bd28293ca712  
Plan Digest: 7d6aacd065ec51d4de813171a8132ec986bb691bc3bddfd96cc3f13e254d47d5

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:district_id_range */ d_id, d_tax
		 FROM district
		 WHERE d_w_id = 42 AND d_id BETWEEN 3 AND 7
		 ORDER BY d_id DESC USING SELECT /*+ read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`district`]) */ /* optimizer_canary_tpcc:district_id_range */ d_id, d_tax
		 FROM district
		 WHERE d_w_id = 42 AND d_id BETWEEN 3 AND 7
		 ORDER BY d_id DESC;
```

<a id="sql-49fc955d"></a>

## SQL: 49fc955d

Schema: optimizer\_canary\_prepare  
SQL Digest: 49fc955d2d16d4ac050f081e31e044cdfe608a14ad2bd3711f57789c5dbbded5

```text
SELECT /* optimizer_canary_tpcc:customer_join_warehouse */ c.c_id, c.c_balance, w.w_name, w.w_tax
		 FROM customer AS c
		 JOIN warehouse AS w ON w.w_id = c.c_w_id
		 WHERE c.c_w_id = 42 AND c.c_d_id = 1 AND c.c_id = 707
```

<a id="current-plan-49fc955d-b612e9ab"></a>

### Current Plan: b612e9ab

Schema: optimizer\_canary\_prepare  
SQL Digest: 49fc955d2d16d4ac050f081e31e044cdfe608a14ad2bd3711f57789c5dbbded5  
Plan Digest: b612e9abcff63019ef6afdafa9c91a717451501f05455a63a153e2a98e53f7d3

```text
	id                   	task	estRows	operator info                                                                                                   
	MergeJoin_14         	root	1      	inner join, left key:optimizer_canary_prepare.customer.c_w_id, right key:optimizer_canary_prepare.warehouse.w_id
	├─Point_Get_37(Build)	root	1      	table:warehouse, handle:42                                                                                      
	└─Point_Get_35(Probe)	root	1      	table:customer, clustered index:PRIMARY(c_w_id, c_d_id, c_id)                                                   
```

<a id="new-plan-49fc955d-b612e9ab"></a>

### New Plan: b612e9ab

Schema: optimizer\_canary\_prepare  
SQL Digest: 49fc955d2d16d4ac050f081e31e044cdfe608a14ad2bd3711f57789c5dbbded5  
Plan Digest: b612e9abcff63019ef6afdafa9c91a717451501f05455a63a153e2a98e53f7d3

```text
id	estRows	task	access object	operator info
MergeJoin_14	1.00	root		inner join, left key:optimizer_canary_prepare.customer.c_w_id, right key:optimizer_canary_prepare.warehouse.w_id
├─Point_Get_28(Build)	1.00	root	table:warehouse	handle:42
└─Point_Get_27(Probe)	1.00	root	table:customer, clustered index:PRIMARY(c_w_id, c_d_id, c_id)	
```

<a id="binding-stmt-49fc955d-b612e9ab"></a>

### Binding Stmt: 49fc955d_b612e9ab

Schema: optimizer\_canary\_prepare  
SQL Digest: 49fc955d2d16d4ac050f081e31e044cdfe608a14ad2bd3711f57789c5dbbded5  
Plan Digest: b612e9abcff63019ef6afdafa9c91a717451501f05455a63a153e2a98e53f7d3

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:customer_join_warehouse */ c.c_id, c.c_balance, w.w_name, w.w_tax
		 FROM customer AS c
		 JOIN warehouse AS w ON w.w_id = c.c_w_id
		 WHERE c.c_w_id = 42 AND c.c_d_id = 1 AND c.c_id = 707 USING SELECT /*+ read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`c`, `optimizer_canary_prepare`.`w`]) */ /* optimizer_canary_tpcc:customer_join_warehouse */ c.c_id, c.c_balance, w.w_name, w.w_tax
		 FROM customer AS c
		 JOIN warehouse AS w ON w.w_id = c.c_w_id
		 WHERE c.c_w_id = 42 AND c.c_d_id = 1 AND c.c_id = 707;
```

<a id="sql-52772a82"></a>

## SQL: 52772a82

Schema: optimizer\_canary\_prepare  
SQL Digest: 52772a82e7401df4b46018712cd37de14a4851ce105e4fca217f825e6e604fdd

```text
SELECT /* optimizer_canary_tpcc:district_join_warehouse */ d.d_id, d.d_name, w.w_name
		 FROM district AS d
		 JOIN warehouse AS w ON w.w_id = d.d_w_id
		 WHERE d.d_w_id = 42 AND d.d_id = 3
```

<a id="current-plan-52772a82-5b116e13"></a>

### Current Plan: 5b116e13

Schema: optimizer\_canary\_prepare  
SQL Digest: 52772a82e7401df4b46018712cd37de14a4851ce105e4fca217f825e6e604fdd  
Plan Digest: 5b116e13409bf1102426da0a1c470773ddf64c620d8a5ff8005134a5d1676cf9

```text
	id                   	task	estRows	operator info                                                                                                   
	MergeJoin_14         	root	1      	inner join, left key:optimizer_canary_prepare.district.d_w_id, right key:optimizer_canary_prepare.warehouse.w_id
	├─Point_Get_37(Build)	root	1      	table:warehouse, handle:42                                                                                      
	└─Point_Get_35(Probe)	root	1      	table:district, clustered index:PRIMARY(d_w_id, d_id)                                                           
```

<a id="new-plan-52772a82-5b116e13"></a>

### New Plan: 5b116e13

Schema: optimizer\_canary\_prepare  
SQL Digest: 52772a82e7401df4b46018712cd37de14a4851ce105e4fca217f825e6e604fdd  
Plan Digest: 5b116e13409bf1102426da0a1c470773ddf64c620d8a5ff8005134a5d1676cf9

```text
id	estRows	task	access object	operator info
MergeJoin_14	1.00	root		inner join, left key:optimizer_canary_prepare.district.d_w_id, right key:optimizer_canary_prepare.warehouse.w_id
├─Point_Get_28(Build)	1.00	root	table:warehouse	handle:42
└─Point_Get_27(Probe)	1.00	root	table:district, clustered index:PRIMARY(d_w_id, d_id)	
```

<a id="binding-stmt-52772a82-5b116e13"></a>

### Binding Stmt: 52772a82_5b116e13

Schema: optimizer\_canary\_prepare  
SQL Digest: 52772a82e7401df4b46018712cd37de14a4851ce105e4fca217f825e6e604fdd  
Plan Digest: 5b116e13409bf1102426da0a1c470773ddf64c620d8a5ff8005134a5d1676cf9

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:district_join_warehouse */ d.d_id, d.d_name, w.w_name
		 FROM district AS d
		 JOIN warehouse AS w ON w.w_id = d.d_w_id
		 WHERE d.d_w_id = 42 AND d.d_id = 3 USING SELECT /*+ read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`d`, `optimizer_canary_prepare`.`w`]) */ /* optimizer_canary_tpcc:district_join_warehouse */ d.d_id, d.d_name, w.w_name
		 FROM district AS d
		 JOIN warehouse AS w ON w.w_id = d.d_w_id
		 WHERE d.d_w_id = 42 AND d.d_id = 3;
```

<a id="sql-3053011a"></a>

## SQL: 3053011a

Schema: optimizer\_canary\_prepare  
SQL Digest: 3053011a790dc4f62a1feba5e7df69309e76252a6b9f22a08917858c504609a5

```text
SELECT /* optimizer_canary_tpcc:orders_customer_range */ o_id, o_entry_d, o_carrier_id
		 FROM orders
		 WHERE o_w_id = 42 AND o_d_id = 1 AND o_c_id = 707
		   AND o_id BETWEEN 700 AND 720
		 ORDER BY o_id
```

<a id="current-plan-3053011a-ff7a44a1"></a>

### Current Plan: ff7a44a1

Schema: optimizer\_canary\_prepare  
SQL Digest: 3053011a790dc4f62a1feba5e7df69309e76252a6b9f22a08917858c504609a5  
Plan Digest: ff7a44a1fc7cfd997a7c26af9552448dce2e2d07a07492138dad4131b20f2939

```text
	id                         	task        	estRows	operator info                                                                                                                
	Sort_5                     	root        	3.81   	optimizer_canary_prepare.orders.o_id                                                                                         
	└─TableReader_17           	root        	3.81   	MppVersion: 3, data:ExchangeSender_16                                                                                        
	  └─ExchangeSender_16      	cop[tiflash]	3.81   	ExchangeType: PassThrough                                                                                                    
	    └─Projection_8         	cop[tiflash]	3.81   	optimizer_canary_prepare.orders.o_id, optimizer_canary_prepare.orders.o_entry_d, optimizer_canary_prepare.orders.o_carrier_id
	      └─Selection_15       	cop[tiflash]	3.81   	eq(optimizer_canary_prepare.orders.o_c_id, 707)                                                                              
	        └─TableRangeScan_14	cop[tiflash]	144.91 	table:orders, range:[42 1 700,42 1 720], keep order:false                                                                    
```

<a id="new-plan-3053011a-72d6501b"></a>

### New Plan: 72d6501b

Schema: optimizer\_canary\_prepare  
SQL Digest: 3053011a790dc4f62a1feba5e7df69309e76252a6b9f22a08917858c504609a5  
Plan Digest: 72d6501b1906e22a98fe9e506d552434a349c60b28b52254376a0043a2f6c2f1

```text
id	estRows	task	access object	operator info
TableReader_42	3.81	root		data:Projection_30
└─Projection_30	3.81	cop[tikv]		optimizer_canary_prepare.orders.o_id, optimizer_canary_prepare.orders.o_entry_d, optimizer_canary_prepare.orders.o_carrier_id
  └─Selection_40	3.81	cop[tikv]		eq(optimizer_canary_prepare.orders.o_c_id, 707)
    └─TableRangeScan_39	144.91	cop[tikv]	table:orders	range:[42 1 700,42 1 720], keep order:true
```

<a id="binding-stmt-3053011a-ff7a44a1"></a>

### Binding Stmt: 3053011a_ff7a44a1

Schema: optimizer\_canary\_prepare  
SQL Digest: 3053011a790dc4f62a1feba5e7df69309e76252a6b9f22a08917858c504609a5  
Plan Digest: ff7a44a1fc7cfd997a7c26af9552448dce2e2d07a07492138dad4131b20f2939

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:orders_customer_range */ o_id, o_entry_d, o_carrier_id
		 FROM orders
		 WHERE o_w_id = 42 AND o_d_id = 1 AND o_c_id = 707
		   AND o_id BETWEEN 700 AND 720
		 ORDER BY o_id USING SELECT /*+ read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`orders`]) */ /* optimizer_canary_tpcc:orders_customer_range */ o_id, o_entry_d, o_carrier_id
		 FROM orders
		 WHERE o_w_id = 42 AND o_d_id = 1 AND o_c_id = 707
		   AND o_id BETWEEN 700 AND 720
		 ORDER BY o_id;
```

<a id="sql-44ab0ff9"></a>

## SQL: 44ab0ff9

Schema: optimizer\_canary\_prepare  
SQL Digest: 44ab0ff9b4a746a154545c1a697ae7db1672e9d1698901648f7e4f40dbc55ef7

```text
SELECT /* optimizer_canary_tpcc:order_status_latest_order */ o_id, o_carrier_id, o_entry_d
		 FROM orders
		 WHERE o_w_id = 1 AND o_d_id = 1 AND o_c_id = 7
		 ORDER BY o_id DESC LIMIT 1
```

<a id="current-plan-44ab0ff9-1e6e8634"></a>

### Current Plan: 1e6e8634

Schema: optimizer\_canary\_prepare  
SQL Digest: 44ab0ff9b4a746a154545c1a697ae7db1672e9d1698901648f7e4f40dbc55ef7  
Plan Digest: 1e6e8634d769139bfd65e2b802dbd6e1210418e680c0091f52f29c31c06b42cf

```text
	id                         	task     	estRows	operator info                                                                                                                
	Projection_7               	root     	1      	optimizer_canary_prepare.orders.o_id, optimizer_canary_prepare.orders.o_carrier_id, optimizer_canary_prepare.orders.o_entry_d
	└─Limit_14                 	root     	1      	offset:0, count:1                                                                                                            
	  └─TableReader_27         	root     	1      	data:Limit_26                                                                                                                
	    └─Limit_26             	cop[tikv]	1      	offset:0, count:1                                                                                                            
	      └─Selection_25       	cop[tikv]	1      	eq(optimizer_canary_prepare.orders.o_c_id, 7)                                                                                
	        └─TableRangeScan_24	cop[tikv]	110.10 	table:orders, range:[1 1,1 1], keep order:true, desc                                                                         
```

<a id="new-plan-44ab0ff9-2e2d3a08"></a>

### New Plan: 2e2d3a08

Schema: optimizer\_canary\_prepare  
SQL Digest: 44ab0ff9b4a746a154545c1a697ae7db1672e9d1698901648f7e4f40dbc55ef7  
Plan Digest: 2e2d3a0834be56617cc33b86ddafa2e8d1fc0d9af9eaa751c252aaac5587f5c7

```text
id	estRows	task	access object	operator info
Projection_7	1.00	root		optimizer_canary_prepare.orders.o_id, optimizer_canary_prepare.orders.o_carrier_id, optimizer_canary_prepare.orders.o_entry_d
└─Projection_54	1.00	root		optimizer_canary_prepare.orders.o_id, optimizer_canary_prepare.orders.o_entry_d, optimizer_canary_prepare.orders.o_carrier_id
  └─IndexLookUp_53	1.00	root		limit embedded(offset:0, count:1)
    ├─Limit_52(Build)	1.00	cop[tikv]		offset:0, count:1
    │ └─IndexRangeScan_50	1.00	cop[tikv]	table:orders, index:idx_order(o_w_id, o_d_id, o_c_id, o_id)	range:[1 1 7,1 1 7], keep order:true, desc
    └─TableRowIDScan_51(Probe)	1.00	cop[tikv]	table:orders	keep order:false
```

<a id="binding-stmt-44ab0ff9-1e6e8634"></a>

### Binding Stmt: 44ab0ff9_1e6e8634

Schema: optimizer\_canary\_prepare  
SQL Digest: 44ab0ff9b4a746a154545c1a697ae7db1672e9d1698901648f7e4f40dbc55ef7  
Plan Digest: 1e6e8634d769139bfd65e2b802dbd6e1210418e680c0091f52f29c31c06b42cf

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:order_status_latest_order */ o_id, o_carrier_id, o_entry_d
		 FROM orders
		 WHERE o_w_id = 1 AND o_d_id = 1 AND o_c_id = 7
		 ORDER BY o_id DESC LIMIT 1 USING SELECT /*+ use_index(@`sel_1` `optimizer_canary_prepare`.`orders` ), order_index(@`sel_1` `optimizer_canary_prepare`.`orders` `primary`), limit_to_cop(@`sel_1`) */ /* optimizer_canary_tpcc:order_status_latest_order */ o_id, o_carrier_id, o_entry_d
		 FROM orders
		 WHERE o_w_id = 1 AND o_d_id = 1 AND o_c_id = 7
		 ORDER BY o_id DESC LIMIT 1;
```

<a id="sql-2be26e27"></a>

## SQL: 2be26e27

Schema: optimizer\_canary\_prepare  
SQL Digest: 2be26e2714f7c174f2c469eed75f77ca4ebafadb42f196f5220bb4ba577324c4

```text
SELECT /* optimizer_canary_tpcc:order_join_customer */ o.o_id, o.o_entry_d, c.c_first, c.c_last
		 FROM orders AS o
		 JOIN customer AS c
		   ON c.c_w_id = o.o_w_id AND c.c_d_id = o.o_d_id AND c.c_id = o.o_c_id
		 WHERE o.o_w_id = 42 AND o.o_d_id = 1 AND o.o_id = 707
```

<a id="current-plan-2be26e27-769bb3f9"></a>

### Current Plan: 769bb3f9

Schema: optimizer\_canary\_prepare  
SQL Digest: 2be26e2714f7c174f2c469eed75f77ca4ebafadb42f196f5220bb4ba577324c4  
Plan Digest: 769bb3f96ee00bbe2835f3788d0d981907bca5ac8209c65ff2e634c62a7bf5a0

```text
	id                     	task     	estRows	operator info                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                      
	IndexJoin_14           	root     	1      	inner join, inner:TableReader_30, outer key:optimizer_canary_prepare.orders.o_c_id, optimizer_canary_prepare.orders.o_w_id, optimizer_canary_prepare.orders.o_d_id, inner key:optimizer_canary_prepare.customer.c_id, optimizer_canary_prepare.customer.c_w_id, optimizer_canary_prepare.customer.c_d_id, equal cond:eq(optimizer_canary_prepare.orders.o_c_id, optimizer_canary_prepare.customer.c_id), eq(optimizer_canary_prepare.orders.o_d_id, optimizer_canary_prepare.customer.c_d_id), eq(optimizer_canary_prepare.orders.o_w_id, optimizer_canary_prepare.customer.c_w_id)
	├─Selection_27(Build)  	root     	1      	not(isnull(optimizer_canary_prepare.orders.o_c_id))                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                
	│ └─Point_Get_26       	root     	1      	table:orders, clustered index:PRIMARY(o_w_id, o_d_id, o_id)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                        
	└─TableReader_30(Probe)	root     	0.01   	data:Selection_29                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  
	  └─Selection_29       	cop[tikv]	0.01   	eq(optimizer_canary_prepare.customer.c_d_id, 1), eq(optimizer_canary_prepare.customer.c_w_id, 42)                                                                                                                                                                                                                                                                                                                                                                                                                                                                                  
	    └─TableRangeScan_28	cop[tikv]	1      	table:c, range: decided by [eq(optimizer_canary_prepare.customer.c_w_id, optimizer_canary_prepare.orders.o_w_id) eq(optimizer_canary_prepare.customer.c_d_id, optimizer_canary_prepare.orders.o_d_id) eq(optimizer_canary_prepare.customer.c_id, optimizer_canary_prepare.orders.o_c_id)], keep order:false                                                                                                                                                                                                                                                                        
```

<a id="new-plan-2be26e27-769bb3f9"></a>

### New Plan: 769bb3f9

Schema: optimizer\_canary\_prepare  
SQL Digest: 2be26e2714f7c174f2c469eed75f77ca4ebafadb42f196f5220bb4ba577324c4  
Plan Digest: 769bb3f96ee00bbe2835f3788d0d981907bca5ac8209c65ff2e634c62a7bf5a0

```text
id	estRows	task	access object	operator info
IndexJoin_15	1.00	root		inner join, inner:TableReader_31, outer key:optimizer_canary_prepare.orders.o_c_id, optimizer_canary_prepare.orders.o_w_id, optimizer_canary_prepare.orders.o_d_id, inner key:optimizer_canary_prepare.customer.c_id, optimizer_canary_prepare.customer.c_w_id, optimizer_canary_prepare.customer.c_d_id, equal cond:eq(optimizer_canary_prepare.orders.o_c_id, optimizer_canary_prepare.customer.c_id), eq(optimizer_canary_prepare.orders.o_d_id, optimizer_canary_prepare.customer.c_d_id), eq(optimizer_canary_prepare.orders.o_w_id, optimizer_canary_prepare.customer.c_w_id)
├─Selection_28(Build)	1.00	root		not(isnull(optimizer_canary_prepare.orders.o_c_id))
│ └─Point_Get_27	1.00	root	table:orders, clustered index:PRIMARY(o_w_id, o_d_id, o_id)	
└─TableReader_31(Probe)	0.01	root		data:Selection_30
  └─Selection_30	0.01	cop[tikv]		eq(optimizer_canary_prepare.customer.c_d_id, 1), eq(optimizer_canary_prepare.customer.c_w_id, 42)
    └─TableRangeScan_29	1.00	cop[tikv]	table:c	range: decided by [eq(optimizer_canary_prepare.customer.c_w_id, optimizer_canary_prepare.orders.o_w_id) eq(optimizer_canary_prepare.customer.c_d_id, optimizer_canary_prepare.orders.o_d_id) eq(optimizer_canary_prepare.customer.c_id, optimizer_canary_prepare.orders.o_c_id)], keep order:false
```

<a id="binding-stmt-2be26e27-769bb3f9"></a>

### Binding Stmt: 2be26e27_769bb3f9

Schema: optimizer\_canary\_prepare  
SQL Digest: 2be26e2714f7c174f2c469eed75f77ca4ebafadb42f196f5220bb4ba577324c4  
Plan Digest: 769bb3f96ee00bbe2835f3788d0d981907bca5ac8209c65ff2e634c62a7bf5a0

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:order_join_customer */ o.o_id, o.o_entry_d, c.c_first, c.c_last
		 FROM orders AS o
		 JOIN customer AS c
		   ON c.c_w_id = o.o_w_id AND c.c_d_id = o.o_d_id AND c.c_id = o.o_c_id
		 WHERE o.o_w_id = 42 AND o.o_d_id = 1 AND o.o_id = 707 USING SELECT /*+ inl_join(`optimizer_canary_prepare`.`c`), use_index(@`sel_1` `optimizer_canary_prepare`.`c` ), no_order_index(@`sel_1` `optimizer_canary_prepare`.`c` `primary`) */ /* optimizer_canary_tpcc:order_join_customer */ o.o_id, o.o_entry_d, c.c_first, c.c_last
		 FROM orders AS o
		 JOIN customer AS c
		   ON c.c_w_id = o.o_w_id AND c.c_d_id = o.o_d_id AND c.c_id = o.o_c_id
		 WHERE o.o_w_id = 42 AND o.o_d_id = 1 AND o.o_id = 707;
```

<a id="sql-14d8bcdf"></a>

## SQL: 14d8bcdf

Schema: optimizer\_canary\_prepare  
SQL Digest: 14d8bcdf812bfee135785bd6204d32d4e3d9403932877657a37d1ec97bccc3ad

```text
SELECT /* optimizer_canary_tpcc:customer_last_count */ COUNT(c_id) AS customer_count
		 FROM customer
		 WHERE c_w_id = 42 AND c_d_id = 1 AND c_last = 'LAST007'
```

<a id="current-plan-14d8bcdf-99616e47"></a>

### Current Plan: 99616e47

Schema: optimizer\_canary\_prepare  
SQL Digest: 14d8bcdf812bfee135785bd6204d32d4e3d9403932877657a37d1ec97bccc3ad  
Plan Digest: 99616e4779344f4d2fc3aba9ec2df6d3335a627461959d559207a9a0bd9abdfc

```text
	id                       	task     	estRows	operator info                                                 
	StreamAgg_22             	root     	1      	funcs:count(Column#25)->Column#23                             
	└─TableReader_23         	root     	1      	data:StreamAgg_11                                             
	  └─StreamAgg_11         	cop[tikv]	1      	funcs:count(optimizer_canary_prepare.customer.c_id)->Column#25
	    └─Selection_21       	cop[tikv]	100    	eq(optimizer_canary_prepare.customer.c_last, "LAST007")       
	      └─TableRangeScan_20	cop[tikv]	1000   	table:customer, range:[42 1,42 1], keep order:false           
```

<a id="new-plan-14d8bcdf-08478e33"></a>

### New Plan: 08478e33

Schema: optimizer\_canary\_prepare  
SQL Digest: 14d8bcdf812bfee135785bd6204d32d4e3d9403932877657a37d1ec97bccc3ad  
Plan Digest: 08478e3372c0666b5ce83fa4ac7b062c69acf0ddd21781e4487b84dbfd6e8489

```text
id	estRows	task	access object	operator info
StreamAgg_30	1.00	root		funcs:count(Column#29)->Column#23
└─IndexReader_31	1.00	root		index:StreamAgg_11
  └─StreamAgg_11	1.00	cop[tikv]		funcs:count(optimizer_canary_prepare.customer.c_id)->Column#29
    └─IndexRangeScan_29	100.00	cop[tikv]	table:customer, index:idx_customer(c_w_id, c_d_id, c_last, c_first)	range:[42 1 "LAST007",42 1 "LAST007"], keep order:false
```

<a id="binding-stmt-14d8bcdf-99616e47"></a>

### Binding Stmt: 14d8bcdf_99616e47

Schema: optimizer\_canary\_prepare  
SQL Digest: 14d8bcdf812bfee135785bd6204d32d4e3d9403932877657a37d1ec97bccc3ad  
Plan Digest: 99616e4779344f4d2fc3aba9ec2df6d3335a627461959d559207a9a0bd9abdfc

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:customer_last_count */ COUNT(c_id) AS customer_count
		 FROM customer
		 WHERE c_w_id = 42 AND c_d_id = 1 AND c_last = 'LAST007' USING SELECT /*+ stream_agg(@`sel_1`), use_index(@`sel_1` `optimizer_canary_prepare`.`customer` ), no_order_index(@`sel_1` `optimizer_canary_prepare`.`customer` `primary`), agg_to_cop(@`sel_1`) */ /* optimizer_canary_tpcc:customer_last_count */ COUNT(c_id) AS customer_count
		 FROM customer
		 WHERE c_w_id = 42 AND c_d_id = 1 AND c_last = 'LAST007';
```

<a id="sql-39601b01"></a>

## SQL: 39601b01

Schema: optimizer\_canary\_prepare  
SQL Digest: 39601b01e7dd65eca522e2c269a675159880ec0925e552ede4658fa9c077838d

```text
SELECT /* optimizer_canary_tpcc:new_order_customer */ c_discount, c_last, c_credit, w_tax
		 FROM customer, warehouse
		 WHERE w_id = 1 AND c_w_id = w_id AND c_d_id = 1 AND c_id = 7
```

<a id="current-plan-39601b01-b965db72"></a>

### Current Plan: b965db72

Schema: optimizer\_canary\_prepare  
SQL Digest: 39601b01e7dd65eca522e2c269a675159880ec0925e552ede4658fa9c077838d  
Plan Digest: b965db72efee9df7f4d9048f9be5621bd2c3eb65b38b3c7481e95b8f67af6996

```text
	id                     	task	estRows	operator info                                                                                                                                                               
	Projection_8           	root	1      	optimizer_canary_prepare.customer.c_discount, optimizer_canary_prepare.customer.c_last, optimizer_canary_prepare.customer.c_credit, optimizer_canary_prepare.warehouse.w_tax
	└─MergeJoin_13         	root	1      	inner join, left key:optimizer_canary_prepare.customer.c_w_id, right key:optimizer_canary_prepare.warehouse.w_id                                                            
	  ├─Point_Get_36(Build)	root	1      	table:warehouse, handle:1                                                                                                                                                   
	  └─Point_Get_34(Probe)	root	1      	table:customer, clustered index:PRIMARY(c_w_id, c_d_id, c_id)                                                                                                               
```

<a id="new-plan-39601b01-b965db72"></a>

### New Plan: b965db72

Schema: optimizer\_canary\_prepare  
SQL Digest: 39601b01e7dd65eca522e2c269a675159880ec0925e552ede4658fa9c077838d  
Plan Digest: b965db72efee9df7f4d9048f9be5621bd2c3eb65b38b3c7481e95b8f67af6996

```text
id	estRows	task	access object	operator info
Projection_8	1.00	root		optimizer_canary_prepare.customer.c_discount, optimizer_canary_prepare.customer.c_last, optimizer_canary_prepare.customer.c_credit, optimizer_canary_prepare.warehouse.w_tax
└─MergeJoin_13	1.00	root		inner join, left key:optimizer_canary_prepare.customer.c_w_id, right key:optimizer_canary_prepare.warehouse.w_id
  ├─Point_Get_27(Build)	1.00	root	table:warehouse	handle:1
  └─Point_Get_26(Probe)	1.00	root	table:customer, clustered index:PRIMARY(c_w_id, c_d_id, c_id)	
```

<a id="binding-stmt-39601b01-b965db72"></a>

### Binding Stmt: 39601b01_b965db72

Schema: optimizer\_canary\_prepare  
SQL Digest: 39601b01e7dd65eca522e2c269a675159880ec0925e552ede4658fa9c077838d  
Plan Digest: b965db72efee9df7f4d9048f9be5621bd2c3eb65b38b3c7481e95b8f67af6996

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:new_order_customer */ c_discount, c_last, c_credit, w_tax
		 FROM customer, warehouse
		 WHERE w_id = 1 AND c_w_id = w_id AND c_d_id = 1 AND c_id = 7 USING SELECT /*+ read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`customer`, `optimizer_canary_prepare`.`warehouse`]) */ /* optimizer_canary_tpcc:new_order_customer */ c_discount, c_last, c_credit, w_tax
		 FROM customer, warehouse
		 WHERE w_id = 1 AND c_w_id = w_id AND c_d_id = 1 AND c_id = 7;
```

<a id="sql-1427f837"></a>

## SQL: 1427f837

Schema: optimizer\_canary\_prepare  
SQL Digest: 1427f8377de924b2b6c4a6adc5aa0c6e01c9bef0dba52ecfe6fb2f0f780a3571

```text
SELECT /* optimizer_canary_tpcc:item_id_range */ i_id, i_name, i_price
		 FROM item
		 WHERE i_id BETWEEN 700 AND 710
		 ORDER BY i_id
```

<a id="current-plan-1427f837-425fc44d"></a>

### Current Plan: 425fc44d

Schema: optimizer\_canary\_prepare  
SQL Digest: 1427f8377de924b2b6c4a6adc5aa0c6e01c9bef0dba52ecfe6fb2f0f780a3571  
Plan Digest: 425fc44d007c6d757c675ba671d751847e3718eab1113f48a83c2dfae8fab980

```text
	id                     	task        	estRows	operator info                                
	Sort_5                 	root        	11     	optimizer_canary_prepare.item.i_id           
	└─TableReader_14       	root        	11     	MppVersion: 3, data:ExchangeSender_13        
	  └─ExchangeSender_13  	cop[tiflash]	11     	ExchangeType: PassThrough                    
	    └─TableRangeScan_12	cop[tiflash]	11     	table:item, range:[700,710], keep order:false
```

<a id="new-plan-1427f837-7c016980"></a>

### New Plan: 7c016980

Schema: optimizer\_canary\_prepare  
SQL Digest: 1427f8377de924b2b6c4a6adc5aa0c6e01c9bef0dba52ecfe6fb2f0f780a3571  
Plan Digest: 7c0169809b413af00dbf47e0b746ba5d738d101e3be5f79dd8da80abebea1d89

```text
id	estRows	task	access object	operator info
TableReader_20	11.00	root		data:TableRangeScan_19
└─TableRangeScan_19	11.00	cop[tikv]	table:item	range:[700,710], keep order:true
```

<a id="binding-stmt-1427f837-425fc44d"></a>

### Binding Stmt: 1427f837_425fc44d

Schema: optimizer\_canary\_prepare  
SQL Digest: 1427f8377de924b2b6c4a6adc5aa0c6e01c9bef0dba52ecfe6fb2f0f780a3571  
Plan Digest: 425fc44d007c6d757c675ba671d751847e3718eab1113f48a83c2dfae8fab980

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:item_id_range */ i_id, i_name, i_price
		 FROM item
		 WHERE i_id BETWEEN 700 AND 710
		 ORDER BY i_id USING SELECT /*+ read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`item`]) */ /* optimizer_canary_tpcc:item_id_range */ i_id, i_name, i_price
		 FROM item
		 WHERE i_id BETWEEN 700 AND 710
		 ORDER BY i_id;
```

<a id="sql-10818e2a"></a>

## SQL: 10818e2a

Schema: optimizer\_canary\_prepare  
SQL Digest: 10818e2af6261c09d1c7738edbfa84de3370c88f001db20cda95d87f650f2465

```text
SELECT /* optimizer_canary_tpcc:stock_level */ COUNT(DISTINCT s_i_id) AS stock_count
		 FROM order_line, stock
		 WHERE ol_w_id = 1 AND ol_d_id = 1
		   AND ol_o_id < 1001 AND ol_o_id >= 981
		   AND s_w_id = 1 AND s_i_id = ol_i_id AND s_quantity < 25
```

<a id="current-plan-10818e2a-737a0989"></a>

### Current Plan: 737a0989

Schema: optimizer\_canary\_prepare  
SQL Digest: 10818e2af6261c09d1c7738edbfa84de3370c88f001db20cda95d87f650f2465  
Plan Digest: 737a0989f9d988f81e31a06028cfad9b6fa5e05de2736b9959299afbfec83c57

```text
	id                       	task     	estRows	operator info                                                                                             
	HashAgg_13               	root     	1      	funcs:count(distinct optimizer_canary_prepare.stock.s_i_id)->Column#30                                    
	└─HashJoin_18            	root     	9998.55	inner join, equal:[eq(optimizer_canary_prepare.order_line.ol_i_id, optimizer_canary_prepare.stock.s_i_id)]
	  ├─TableReader_39(Build)	root     	500    	data:Selection_38                                                                                         
	  │ └─Selection_38       	cop[tikv]	500    	lt(optimizer_canary_prepare.stock.s_quantity, 25)                                                         
	  │   └─TableRangeScan_37	cop[tikv]	1000   	table:stock, range:[1,1], keep order:false                                                                
	  └─TableReader_33(Probe)	root     	702.98 	data:TableRangeScan_32                                                                                    
	    └─TableRangeScan_32  	cop[tikv]	702.98 	table:order_line, range:[1 1 981,1 1 1001), keep order:false                                              
```

<a id="new-plan-10818e2a-1f11d658"></a>

### New Plan: 1f11d658

Schema: optimizer\_canary\_prepare  
SQL Digest: 10818e2af6261c09d1c7738edbfa84de3370c88f001db20cda95d87f650f2465  
Plan Digest: 1f11d658a948473734439c8ef84f36cf5e91dab18d6991b41553debfb353bfd7

```text
id	estRows	task	access object	operator info
TableReader_91	1.00	root		MppVersion: 3, data:ExchangeSender_90
└─ExchangeSender_90	1.00	mpp[tiflash]		ExchangeType: PassThrough
  └─Projection_84	1.00	mpp[tiflash]		Column#30
    └─HashAgg_85	1.00	mpp[tiflash]		funcs:sum(Column#32)->Column#30
      └─ExchangeReceiver_89	1.00	mpp[tiflash]		
        └─ExchangeSender_88	1.00	mpp[tiflash]		ExchangeType: PassThrough, Compression: FAST
          └─HashAgg_85	1.00	mpp[tiflash]		funcs:count(distinct optimizer_canary_prepare.stock.s_i_id)->Column#32, stream_count: 8
            └─ExchangeReceiver_87	1.00	mpp[tiflash]		stream_count: 8
              └─ExchangeSender_86	1.00	mpp[tiflash]		ExchangeType: HashPartition, Compression: FAST, Hash Cols: [name: optimizer_canary_prepare.stock.s_i_id, collate: binary], stream_count: 8
                └─HashAgg_83	1.00	mpp[tiflash]		group by:optimizer_canary_prepare.stock.s_i_id, 
                  └─Projection_56	9998.55	mpp[tiflash]		optimizer_canary_prepare.stock.s_i_id
                    └─HashJoin_55	9998.55	mpp[tiflash]		inner join, equal:[eq(optimizer_canary_prepare.order_line.ol_i_id, optimizer_canary_prepare.stock.s_i_id)]
                      ├─ExchangeReceiver_34(Build)	702.98	mpp[tiflash]		
                      │ └─ExchangeSender_33	702.98	mpp[tiflash]		ExchangeType: Broadcast, Compression: FAST
                      │   └─TableRangeScan_32	702.98	mpp[tiflash]	table:order_line	range:[1 1 981,1 1 1001), keep order:false
                      └─Selection_36(Probe)	500.00	mpp[tiflash]		lt(optimizer_canary_prepare.stock.s_quantity, 25)
                        └─TableRangeScan_35	1000.00	mpp[tiflash]	table:stock	range:[1,1], keep order:false
```

<a id="binding-stmt-10818e2a-737a0989"></a>

### Binding Stmt: 10818e2a_737a0989

Schema: optimizer\_canary\_prepare  
SQL Digest: 10818e2af6261c09d1c7738edbfa84de3370c88f001db20cda95d87f650f2465  
Plan Digest: 737a0989f9d988f81e31a06028cfad9b6fa5e05de2736b9959299afbfec83c57

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:stock_level */ COUNT(DISTINCT s_i_id) AS stock_count
		 FROM order_line, stock
		 WHERE ol_w_id = 1 AND ol_d_id = 1
		   AND ol_o_id < 1001 AND ol_o_id >= 981
		   AND s_w_id = 1 AND s_i_id = ol_i_id AND s_quantity < 25 USING SELECT /*+ hash_agg(@`sel_1`), hash_join_build(`optimizer_canary_prepare`.`stock`), use_index(@`sel_1` `optimizer_canary_prepare`.`order_line` ), no_order_index(@`sel_1` `optimizer_canary_prepare`.`order_line` `primary`), use_index(@`sel_1` `optimizer_canary_prepare`.`stock` ), no_order_index(@`sel_1` `optimizer_canary_prepare`.`stock` `primary`) */ /* optimizer_canary_tpcc:stock_level */ COUNT(DISTINCT s_i_id) AS stock_count
		 FROM order_line, stock
		 WHERE ol_w_id = 1 AND ol_d_id = 1
		   AND ol_o_id < 1001 AND ol_o_id >= 981
		   AND s_w_id = 1 AND s_i_id = ol_i_id AND s_quantity < 25;
```

<a id="sql-747de7b8"></a>

## SQL: 747de7b8

Schema: optimizer\_canary\_prepare  
SQL Digest: 747de7b8640b16ee35d4cf5ff7b7ed5a6f34833ffdb2a9e9e208bf471bb5fbed

```text
SELECT /* optimizer_canary_tpcc:warehouse_id_range */ w_id, w_name
		 FROM warehouse
		 WHERE w_id BETWEEN 40 AND 45
		 ORDER BY w_id
```

<a id="current-plan-747de7b8-fe087e4e"></a>

### Current Plan: fe087e4e

Schema: optimizer\_canary\_prepare  
SQL Digest: 747de7b8640b16ee35d4cf5ff7b7ed5a6f34833ffdb2a9e9e208bf471bb5fbed  
Plan Digest: fe087e4ecd5c64f14e0792f5994769a84a2547cc54f535e028a2cf408e4df0ba

```text
	id                     	task        	estRows	operator info                                   
	Sort_5                 	root        	6      	optimizer_canary_prepare.warehouse.w_id         
	└─TableReader_14       	root        	6      	MppVersion: 3, data:ExchangeSender_13           
	  └─ExchangeSender_13  	cop[tiflash]	6      	ExchangeType: PassThrough                       
	    └─TableRangeScan_12	cop[tiflash]	6      	table:warehouse, range:[40,45], keep order:false
```

<a id="new-plan-747de7b8-9bf0da55"></a>

### New Plan: 9bf0da55

Schema: optimizer\_canary\_prepare  
SQL Digest: 747de7b8640b16ee35d4cf5ff7b7ed5a6f34833ffdb2a9e9e208bf471bb5fbed  
Plan Digest: 9bf0da556f2db59522d84bf543af968840b8d9f1793471d2babb88a99e2ff5b7

```text
id	estRows	task	access object	operator info
TableReader_20	6.00	root		data:TableRangeScan_19
└─TableRangeScan_19	6.00	cop[tikv]	table:warehouse	range:[40,45], keep order:true
```

<a id="binding-stmt-747de7b8-fe087e4e"></a>

### Binding Stmt: 747de7b8_fe087e4e

Schema: optimizer\_canary\_prepare  
SQL Digest: 747de7b8640b16ee35d4cf5ff7b7ed5a6f34833ffdb2a9e9e208bf471bb5fbed  
Plan Digest: fe087e4ecd5c64f14e0792f5994769a84a2547cc54f535e028a2cf408e4df0ba

```text
CREATE GLOBAL BINDING FOR SELECT /* optimizer_canary_tpcc:warehouse_id_range */ w_id, w_name
		 FROM warehouse
		 WHERE w_id BETWEEN 40 AND 45
		 ORDER BY w_id USING SELECT /*+ read_from_storage(@`sel_1` tiflash[`optimizer_canary_prepare`.`warehouse`]) */ /* optimizer_canary_tpcc:warehouse_id_range */ w_id, w_name
		 FROM warehouse
		 WHERE w_id BETWEEN 40 AND 45
		 ORDER BY w_id;
```
