# Index Recommendation Overview

<table>
<thead><tr><th>SQL</th><th>Total Time</th><th>Count</th><th>Database</th><th>Table</th><th>Est Cost Change</th><th>Est Improvement</th><th>Advice</th></tr></thead>
<tbody>
<tr><td rowspan="1"><a href="#sql-cc596423402ee0ede86d44e6d5a81c7871191ceacb128cdcd99d10be85cc49b9">cc596423</a></td><td rowspan="1">2.36s</td><td rowspan="1">6</td><td>test</td><td>orders</td><td>46430949.50 -&gt; 1172.88</td><td>99.9975%</td><td><a href="#index-advice-cc596423402ee0ede86d44e6d5a81c7871191ceacb128cdcd99d10be85cc49b9-1">detail</a></td></tr>
<tr><td rowspan="1"><a href="#sql-63a05b9892b62c5a876a284fbd70829d06ba8202c25eb5c100965c0b7bbfb1b9">63a05b98</a></td><td rowspan="1">1.26s</td><td rowspan="1">5</td><td>test</td><td>orders</td><td>46431128.54 -&gt; 17416.61</td><td>99.9625%</td><td><a href="#index-advice-63a05b9892b62c5a876a284fbd70829d06ba8202c25eb5c100965c0b7bbfb1b9-1">detail</a></td></tr>
<tr><td rowspan="1"><a href="#sql-940173e4c71643da0175cdcd9016707189f1a04b30d156334424f778d673268c">940173e4</a></td><td rowspan="1">0.95s</td><td rowspan="1">7</td><td>test</td><td>customer</td><td>4670958.02 -&gt; 1174.09</td><td>99.9749%</td><td><a href="#index-advice-940173e4c71643da0175cdcd9016707189f1a04b30d156334424f778d673268c-1">detail</a></td></tr>
<tr><td rowspan="1"><a href="#sql-78d2b15d217ab515af03ff2b589c29cf2a1f0380fae3ca44b102e80ca2298573">78d2b15d</a></td><td rowspan="1">0.89s</td><td rowspan="1">4</td><td>test</td><td>customer</td><td>6521776.71 -&gt; 1182.50</td><td>99.9819%</td><td><a href="#index-advice-78d2b15d217ab515af03ff2b589c29cf2a1f0380fae3ca44b102e80ca2298573-1">detail</a></td></tr>
<tr><td rowspan="1"><a href="#sql-a65ee4eb49299740c484dd2667ba6fa9964559481efe3f836441aaf35e770e19">a65ee4eb</a></td><td rowspan="1">0.17s</td><td rowspan="1">7</td><td>test</td><td>lineitem</td><td>61884742.83 -&gt; 347.68</td><td>99.9994%</td><td><a href="#index-advice-a65ee4eb49299740c484dd2667ba6fa9964559481efe3f836441aaf35e770e19-1">detail</a></td></tr>
<tr><td rowspan="2"><a href="#sql-a3836a08683b6b039249fbd9dbdba3e0a48ce454539593db7944a50892510989">a3836a08</a></td><td rowspan="2">0.06s</td><td rowspan="2">13</td><td>test</td><td>t1</td><td>765424.20 -&gt; 730128.17</td><td>4.6113%</td><td><a href="#index-advice-a3836a08683b6b039249fbd9dbdba3e0a48ce454539593db7944a50892510989-1">detail</a></td></tr>
<tr><td>test</td><td>t2</td><td>765424.20 -&gt; 289706.55</td><td>62.1508%</td><td><a href="#index-advice-a3836a08683b6b039249fbd9dbdba3e0a48ce454539593db7944a50892510989-2">detail</a></td></tr>
</tbody>
</table>

# Recommendation List

<a id="sql-cc596423402ee0ede86d44e6d5a81c7871191ceacb128cdcd99d10be85cc49b9"></a>

## SQL: cc596423

<pre style="background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;">SQL Digest: cc596423402ee0ede86d44e6d5a81c7871191ceacb128cdcd99d10be85cc49b9</pre>

### SQL Statement

```sql
select o_orderdate, o_custkey, o_orderkey from orders where o_shippriority = 100 limit 100
```

<a id="index-advice-cc596423402ee0ede86d44e6d5a81c7871191ceacb128cdcd99d10be85cc49b9-1"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_o_shippriority ON orders(o_shippriority);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                          estRows  estCost       task          access object  operator info
Projection_7                1.00     46430949.50   root                         test.orders.o_orderdate, test.orders.o_custkey, test.orders.o_orderkey
└─Limit_12                  1.00     46430949.20   root                         offset:0, count:100
  └─TableReader_27          1.00     46430949.20   root                         MppVersion: 3, data:ExchangeSender_26
    └─ExchangeSender_26     1.00     696464097.18  mpp[tiflash]                 ExchangeType: PassThrough
      └─Limit_25            1.00     696464097.18  mpp[tiflash]                 offset:0, count:100
        └─TableFullScan_24  1.00     696464097.18  mpp[tiflash]  table:orders   pushed down filter:eq(test.orders.o_shippriority, 100), keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
select /*+ hypo_index(orders, idx_o_shippriority, o_shippriority) */ o_orderdate, o_custkey, o_orderkey from orders where o_shippriority = 100 limit 100
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                              estRows  estCost  task       access object                                           operator info
Projection_7                    1.00     1172.88  root                                                               test.orders.o_orderdate, test.orders.o_custkey, test.orders.o_orderkey
└─Projection_18                 1.00     1172.59  root                                                               test.orders.o_orderkey, test.orders.o_custkey, test.orders.o_orderdate
  └─IndexLookUp_17              1.00     1172.29  root                                                               limit embedded(offset:0, count:100)
    ├─Limit_16(Build)           1.00     203.50   cop[tikv]                                                          offset:0, count:100
    │ └─IndexRangeScan_14       1.00     203.50   cop[tikv]  table:orders, index:idx_o_shippriority(o_shippriority)  range:[100,100], keep order:false
    └─TableRowIDScan_15(Probe)  1.00     319.31   cop[tikv]  table:orders                                            keep order:false</code></pre>

<a id="sql-63a05b9892b62c5a876a284fbd70829d06ba8202c25eb5c100965c0b7bbfb1b9"></a>

## SQL: 63a05b98

<pre style="background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;">SQL Digest: 63a05b9892b62c5a876a284fbd70829d06ba8202c25eb5c100965c0b7bbfb1b9</pre>

### SQL Statement

```sql
select o_orderdate, o_custkey, o_orderkey from orders where o_custkey = 100 limit 100
```

<a id="index-advice-63a05b9892b62c5a876a284fbd70829d06ba8202c25eb5c100965c0b7bbfb1b9-1"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_o_custkey ON orders(o_custkey);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                          estRows  estCost       task          access object  operator info
Projection_7                14.86    46431128.54   root                         test.orders.o_orderdate, test.orders.o_custkey, test.orders.o_orderkey
└─Limit_12                  14.86    46431124.09   root                         offset:0, count:100
  └─TableReader_27          14.86    46431124.09   root                         MppVersion: 3, data:ExchangeSender_26
    └─ExchangeSender_26     14.86    696465292.51  mpp[tiflash]                 ExchangeType: PassThrough
      └─Limit_25            14.86    696465292.51  mpp[tiflash]                 offset:0, count:100
        └─TableFullScan_24  14.86    696465292.51  mpp[tiflash]  table:orders   pushed down filter:eq(test.orders.o_custkey, 100), keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
select /*+ hypo_index(orders, idx_o_custkey, o_custkey) */ o_orderdate, o_custkey, o_orderkey from orders where o_custkey = 100 limit 100
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                            estRows  estCost   task       access object                                 operator info
Projection_7                  14.86    17416.61  root                                                     test.orders.o_orderdate, test.orders.o_custkey, test.orders.o_orderkey
└─IndexLookUp_17              14.86    17412.16  root                                                     limit embedded(offset:0, count:100)
  ├─Limit_16(Build)           14.86    3023.28   cop[tikv]                                                offset:0, count:100
  │ └─IndexRangeScan_14       14.86    3023.28   cop[tikv]  table:orders, index:idx_o_custkey(o_custkey)  range:[100,100], keep order:false
  └─TableRowIDScan_15(Probe)  14.86    4743.83   cop[tikv]  table:orders                                  keep order:false</code></pre>

<a id="sql-940173e4c71643da0175cdcd9016707189f1a04b30d156334424f778d673268c"></a>

## SQL: 940173e4

<pre style="background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;">SQL Digest: 940173e4c71643da0175cdcd9016707189f1a04b30d156334424f778d673268c</pre>

### SQL Statement

```sql
select c_nationkey, c_name, c_custkey from customer where c_nationkey = 100 and c_acctbal < 3 limit 10
```

<a id="index-advice-940173e4c71643da0175cdcd9016707189f1a04b30d156334424f778d673268c-1"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_c_nationkey ON customer(c_nationkey);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                            estRows  estCost      task          access object   operator info
Projection_7                  1.00     4670958.02   root                          test.customer.c_nationkey, test.customer.c_name, test.customer.c_custkey
└─Limit_12                    1.00     4670957.72   root                          offset:0, count:10
  └─TableReader_29            1.00     4670957.72   root                          MppVersion: 3, data:ExchangeSender_28
    └─ExchangeSender_28       1.00     70064117.24  mpp[tiflash]                  ExchangeType: PassThrough
      └─Limit_27              1.00     70064117.24  mpp[tiflash]                  offset:0, count:10
        └─Selection_26        1.00     70064117.24  mpp[tiflash]                  lt(test.customer.c_acctbal, 3)
          └─TableFullScan_25  1.00     70064114.84  mpp[tiflash]  table:customer  pushed down filter:eq(test.customer.c_nationkey, 100), keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
select /*+ hypo_index(customer, idx_c_nationkey, c_nationkey) */ c_nationkey, c_name, c_custkey from customer where c_nationkey = 100 and c_acctbal < 3 limit 10
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                              estRows  estCost  task       access object                                       operator info
Projection_7                    1.00     1174.09  root                                                           test.customer.c_nationkey, test.customer.c_name, test.customer.c_custkey
└─Limit_10                      1.00     1173.79  root                                                           offset:0, count:10
  └─IndexLookUp_18              1.00     1173.79  root                                                           
    ├─IndexRangeScan_14(Build)  1.00     203.50   cop[tikv]  table:customer, index:idx_c_nationkey(c_nationkey)  range:[100,100], keep order:false
    └─Limit_17(Probe)           1.00     380.76   cop[tikv]                                                      offset:0, count:10
      └─Selection_16            1.00     380.76   cop[tikv]                                                      lt(test.customer.c_acctbal, 3)
        └─TableRowIDScan_15     1.00     330.86   cop[tikv]  table:customer                                      keep order:false</code></pre>

<a id="sql-78d2b15d217ab515af03ff2b589c29cf2a1f0380fae3ca44b102e80ca2298573"></a>

## SQL: 78d2b15d

<pre style="background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;">SQL Digest: 78d2b15d217ab515af03ff2b589c29cf2a1f0380fae3ca44b102e80ca2298573</pre>

### SQL Statement

```sql
select c_nationkey, c_name, c_custkey from customer where c_nationkey = 10 and c_acctbal = 10 limit 10
```

<a id="index-advice-78d2b15d217ab515af03ff2b589c29cf2a1f0380fae3ca44b102e80ca2298573-1"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_c_acctbal_c_nationkey ON customer(c_acctbal,c_nationkey);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                            estRows  estCost      task          access object   operator info
Projection_7                  1.00     6521776.71   root                          test.customer.c_nationkey, test.customer.c_name, test.customer.c_custkey
└─Limit_12                    1.00     6521776.41   root                          offset:0, count:10
  └─TableReader_29            1.00     6521776.41   root                          MppVersion: 3, data:ExchangeSender_28
    └─ExchangeSender_28       1.00     97826397.58  mpp[tiflash]                  ExchangeType: PassThrough
      └─Limit_27              1.00     97826397.58  mpp[tiflash]                  offset:0, count:10
        └─Selection_26        1.00     97826397.58  mpp[tiflash]                  eq(test.customer.c_nationkey, 10)
          └─TableFullScan_25  1.81     97826393.23  mpp[tiflash]  table:customer  pushed down filter:eq(test.customer.c_acctbal, 10), keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
select /*+ hypo_index(customer, idx_c_acctbal_c_nationkey, c_acctbal,c_nationkey) */ c_nationkey, c_name, c_custkey from customer where c_nationkey = 10 and c_acctbal = 10 limit 10
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                              estRows  estCost  task       access object                                                            operator info
Projection_7                    1.00     1182.50  root                                                                                test.customer.c_nationkey, test.customer.c_name, test.customer.c_custkey
└─Projection_18                 1.00     1182.20  root                                                                                test.customer.c_custkey, test.customer.c_name, test.customer.c_nationkey
  └─IndexLookUp_17              1.00     1181.90  root                                                                                limit embedded(offset:0, count:10)
    ├─Limit_16(Build)           1.00     257.30   cop[tikv]                                                                           offset:0, count:10
    │ └─IndexRangeScan_14       1.00     257.30   cop[tikv]  table:customer, index:idx_c_acctbal_c_nationkey(c_acctbal, c_nationkey)  range:[10.00 10,10.00 10], keep order:false
    └─TableRowIDScan_15(Probe)  1.00     330.86   cop[tikv]  table:customer                                                           keep order:false</code></pre>

<a id="sql-a65ee4eb49299740c484dd2667ba6fa9964559481efe3f836441aaf35e770e19"></a>

## SQL: a65ee4eb

<pre style="background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;">SQL Digest: a65ee4eb49299740c484dd2667ba6fa9964559481efe3f836441aaf35e770e19</pre>

### SQL Statement

```sql
select L_PARTKEY, l_comment from lineitem where L_PARTKEY = 100 limit 10
```

<a id="index-advice-a65ee4eb49299740c484dd2667ba6fa9964559481efe3f836441aaf35e770e19-1"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_l_partkey_L_PARTKEY_l_comment ON lineitem(l_partkey,L_PARTKEY,l_comment);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                        estRows  estCost       task          access object   operator info
Limit_12                  10.00    61884742.83   root                          offset:0, count:10
└─TableReader_27          10.00    61884742.83   root                          MppVersion: 3, data:ExchangeSender_26
  └─ExchangeSender_26     10.00    928269778.39  mpp[tiflash]                  ExchangeType: PassThrough
    └─Limit_25            10.00    928269778.39  mpp[tiflash]                  offset:0, count:10
      └─TableFullScan_24  10.03    928269778.39  mpp[tiflash]  table:lineitem  pushed down filter:eq(test.lineitem.l_partkey, 100), keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
select /*+ hypo_index(lineitem, idx_l_partkey_L_PARTKEY_l_comment, l_partkey,L_PARTKEY,l_comment) */ L_PARTKEY, l_comment from lineitem where L_PARTKEY = 100 limit 10
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                       estRows  estCost  task       access object                                                                             operator info
Limit_9                  10.00    347.68   root                                                                                                 offset:0, count:10
└─IndexReader_16         10.00    347.68   root                                                                                                 index:Limit_15
  └─Limit_15             10.00    2760.01  cop[tikv]                                                                                            offset:0, count:10
    └─IndexRangeScan_14  10.00    2760.01  cop[tikv]  table:lineitem, index:idx_l_partkey_L_PARTKEY_l_comment(l_partkey, L_PARTKEY, l_comment)  range:[100,100], keep order:false</code></pre>

<a id="sql-a3836a08683b6b039249fbd9dbdba3e0a48ce454539593db7944a50892510989"></a>

## SQL: a3836a08

<pre style="background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;">SQL Digest: a3836a08683b6b039249fbd9dbdba3e0a48ce454539593db7944a50892510989</pre>

### SQL Statement

```sql
SELECT *
FROM t1
JOIN t2 ON t1.a = t2.a
WHERE t1.b = 1
```

<a id="index-advice-a3836a08683b6b039249fbd9dbdba3e0a48ce454539593db7944a50892510989-1"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_b ON t1(b);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                       estRows   estCost     task       access object  operator info
HashJoin_11              100.00    765424.20   root                      inner join, equal:[eq(test.t1.a, test.t2.a)]
├─TableReader_26(Build)  100.00    230600.53   root                      data:Selection_25
│ └─Selection_25         100.00    3440000.00  cop[tikv]                 eq(test.t1.b, 1), not(isnull(test.t1.a))
│   └─TableFullScan_24   10000.00  2442000.00  cop[tikv]  table:t1       keep order:false
└─TableReader_29(Probe)  10000.00  322786.67   root                      data:Selection_28
  └─Selection_28         10000.00  2941000.00  cop[tikv]                 not(isnull(test.t2.a))
    └─TableFullScan_27   10000.00  2442000.00  cop[tikv]  table:t2       keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
SELECT /*+ hypo_index(t1, idx_b, b) */ *
FROM t1
JOIN t2 ON t1.a = t2.a
WHERE t1.b = 1
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                            estRows   estCost     task       access object             operator info
HashJoin_11                   100.00    730128.17   root                                 inner join, equal:[eq(test.t1.a, test.t2.a)]
├─IndexLookUp_27(Build)       100.00    195304.50   root                                 
│ ├─IndexRangeScan_24(Build)  100.00    20350.00    cop[tikv]  table:t1, index:idx_b(b)  range:[1,1], keep order:false
│ └─Selection_26(Probe)       100.00    29410.00    cop[tikv]                            not(isnull(test.t1.a))
│   └─TableRowIDScan_25       100.00    24420.00    cop[tikv]  table:t1                  keep order:false
└─TableReader_30(Probe)       10000.00  322786.67   root                                 data:Selection_29
  └─Selection_29              10000.00  2941000.00  cop[tikv]                            not(isnull(test.t2.a))
    └─TableFullScan_28        10000.00  2442000.00  cop[tikv]  table:t2                  keep order:false</code></pre>

<a id="index-advice-a3836a08683b6b039249fbd9dbdba3e0a48ce454539593db7944a50892510989-2"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_a ON t2(a);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                       estRows   estCost     task       access object  operator info
HashJoin_11              100.00    765424.20   root                      inner join, equal:[eq(test.t1.a, test.t2.a)]
├─TableReader_26(Build)  100.00    230600.53   root                      data:Selection_25
│ └─Selection_25         100.00    3440000.00  cop[tikv]                 eq(test.t1.b, 1), not(isnull(test.t1.a))
│   └─TableFullScan_24   10000.00  2442000.00  cop[tikv]  table:t1       keep order:false
└─TableReader_29(Probe)  10000.00  322786.67   root                      data:Selection_28
  └─Selection_28         10000.00  2941000.00  cop[tikv]                 not(isnull(test.t2.a))
    └─TableFullScan_27   10000.00  2442000.00  cop[tikv]  table:t2       keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
SELECT /*+ hypo_index(t2, idx_a, a) */ *
FROM t1
JOIN t2 ON t1.a = t2.a
WHERE t1.b = 1
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                            estRows   estCost     task       access object             operator info
IndexJoin_13                  100.00    289706.55   root                                 inner join, inner:IndexLookUp_30, outer key:test.t1.a, inner key:test.t2.a, equal cond:eq(test.t1.a, test.t2.a)
├─TableReader_26(Build)       100.00    230600.53   root                                 data:Selection_25
│ └─Selection_25              100.00    3440000.00  cop[tikv]                            eq(test.t1.b, 1), not(isnull(test.t1.a))
│   └─TableFullScan_24        10000.00  2442000.00  cop[tikv]  table:t1                  keep order:false
└─IndexLookUp_30(Probe)       100.00    1955.71     root                                 
  ├─Selection_29(Build)       100.00    253.40      cop[tikv]                            not(isnull(test.t2.a))
  │ └─IndexRangeScan_27       100.00    203.50      cop[tikv]  table:t2, index:idx_a(a)  range: decided by [eq(test.t2.a, test.t1.a)], keep order:false
  └─TableRowIDScan_28(Probe)  100.00    244.20      cop[tikv]  table:t2                  keep order:false</code></pre>
