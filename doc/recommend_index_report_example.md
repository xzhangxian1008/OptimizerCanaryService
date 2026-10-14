# Index Recommendation Overview

<table>
<thead><tr><th>SQL</th><th>Total Time</th><th>Count</th><th>Database</th><th>Table</th><th>Est Cost Change</th><th>Est Improvement</th><th>Advice</th></tr></thead>
<tbody>
<tr><td rowspan="1"><a href="#sql-78d2b15d217ab515af03ff2b589c29cf2a1f0380fae3ca44b102e80ca2298573">78d2b15d</a></td><td rowspan="1">0.11s</td><td rowspan="1">5</td><td>test</td><td>customer</td><td>6521776.71 -&gt; 1182.50</td><td>99.9819%</td><td><a href="#index-advice-78d2b15d217ab515af03ff2b589c29cf2a1f0380fae3ca44b102e80ca2298573-1">detail</a></td></tr>
<tr><td rowspan="1"><a href="#sql-e9572eabc1d81a0065a61aa08a38b91fa60045d08bd3758f1c69d6c0fc7c8a2b">e9572eab</a></td><td rowspan="1">0.10s</td><td rowspan="1">4</td><td>test</td><td>orders</td><td>52569425.39 -&gt; 21147.43</td><td>99.9598%</td><td><a href="#index-advice-e9572eabc1d81a0065a61aa08a38b91fa60045d08bd3758f1c69d6c0fc7c8a2b-1">detail</a></td></tr>
<tr><td rowspan="1"><a href="#sql-8b124dc6fbb9f51ba9799f2d6007d0f959d18f00ed15a45b93c0a0451a4a2e68">8b124dc6</a></td><td rowspan="1">0.08s</td><td rowspan="1">2</td><td>test</td><td>orders</td><td>46432561.23 -&gt; 3277.60</td><td>99.9929%</td><td><a href="#index-advice-8b124dc6fbb9f51ba9799f2d6007d0f959d18f00ed15a45b93c0a0451a4a2e68-1">detail</a></td></tr>
<tr><td rowspan="1"><a href="#sql-d2f47f5e68ed76c64f93f801d83d8ee4b855e2eacb00d0363a058ad5050aacca">d2f47f5e</a></td><td rowspan="1">0.03s</td><td rowspan="1">1</td><td>test</td><td>orders</td><td>67228764.21 -&gt; 34.89</td><td>99.9999%</td><td><a href="#index-advice-d2f47f5e68ed76c64f93f801d83d8ee4b855e2eacb00d0363a058ad5050aacca-1">detail</a></td></tr>
<tr><td rowspan="2"><a href="#sql-a3836a08683b6b039249fbd9dbdba3e0a48ce454539593db7944a50892510989">a3836a08</a></td><td rowspan="2">0.03s</td><td rowspan="2">4</td><td>test</td><td>t1</td><td>765424.20 -&gt; 730128.17</td><td>4.6113%</td><td><a href="#index-advice-a3836a08683b6b039249fbd9dbdba3e0a48ce454539593db7944a50892510989-1">detail</a></td></tr>
<tr><td>test</td><td>t2</td><td>765424.20 -&gt; 289706.55</td><td>62.1508%</td><td><a href="#index-advice-a3836a08683b6b039249fbd9dbdba3e0a48ce454539593db7944a50892510989-2">detail</a></td></tr>
<tr><td rowspan="1"><a href="#sql-2e3888683ade4cf13ccd94e3608c528f8fafea6102ca880fb829de2026eee6d3">2e388868</a></td><td rowspan="1">0.02s</td><td rowspan="1">1</td><td>test</td><td>orders</td><td>67228826.77 -&gt; 75.18</td><td>99.9999%</td><td><a href="#index-advice-2e3888683ade4cf13ccd94e3608c528f8fafea6102ca880fb829de2026eee6d3-1">detail</a></td></tr>
<tr><td rowspan="1"><a href="#sql-a93e251b62c1f1e624e67fa5163d1e3468663f2fa5cb38a239063ac009a0a371">a93e251b</a></td><td rowspan="1">0.02s</td><td rowspan="1">1</td><td>test</td><td>orders</td><td>67456764.23 -&gt; 58333155.05</td><td>13.5251%</td><td><a href="#index-advice-a93e251b62c1f1e624e67fa5163d1e3468663f2fa5cb38a239063ac009a0a371-1">detail</a></td></tr>
<tr><td rowspan="1"><a href="#sql-4d49791f77e4b9bc1298a0437c7ff2f9af38f84db4cea0cb3aa3dcbfb125b529">4d49791f</a></td><td rowspan="1">0.01s</td><td rowspan="1">1</td><td>test</td><td>orders</td><td>48831709.53 -&gt; 262.68</td><td>99.9995%</td><td><a href="#index-advice-4d49791f77e4b9bc1298a0437c7ff2f9af38f84db4cea0cb3aa3dcbfb125b529-1">detail</a></td></tr>
</tbody>
</table>

# Recommendation List

<a id="sql-78d2b15d217ab515af03ff2b589c29cf2a1f0380fae3ca44b102e80ca2298573"></a>

## SQL: 78d2b15d

<pre style="background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;">SQL Digest: 78d2b15d217ab515af03ff2b589c29cf2a1f0380fae3ca44b102e80ca2298573</pre>

### SQL Statement

```sql
select c_nationkey, c_name, c_custkey from customer where c_nationkey = 23 and c_acctbal = 1033597 limit 10
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
        └─Selection_26        1.00     97826397.58  mpp[tiflash]                  eq(test.customer.c_nationkey, 23)
          └─TableFullScan_25  1.81     97826393.23  mpp[tiflash]  table:customer  pushed down filter:eq(test.customer.c_acctbal, 1033597), keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
select /*+ hypo_index(customer, idx_c_acctbal_c_nationkey, c_acctbal,c_nationkey) */ c_nationkey, c_name, c_custkey from customer where c_nationkey = 23 and c_acctbal = 1033597 limit 10
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                              estRows  estCost  task       access object                                                            operator info
Projection_7                    1.00     1182.50  root                                                                                test.customer.c_nationkey, test.customer.c_name, test.customer.c_custkey
└─Projection_18                 1.00     1182.20  root                                                                                test.customer.c_custkey, test.customer.c_name, test.customer.c_nationkey
  └─IndexLookUp_17              1.00     1181.90  root                                                                                limit embedded(offset:0, count:10)
    ├─Limit_16(Build)           1.00     257.30   cop[tikv]                                                                           offset:0, count:10
    │ └─IndexRangeScan_14       1.00     257.30   cop[tikv]  table:customer, index:idx_c_acctbal_c_nationkey(c_acctbal, c_nationkey)  range:[1033597.00 23,1033597.00 23], keep order:false
    └─TableRowIDScan_15(Probe)  1.00     330.86   cop[tikv]  table:customer                                                           keep order:false</code></pre>

<a id="sql-e9572eabc1d81a0065a61aa08a38b91fa60045d08bd3758f1c69d6c0fc7c8a2b"></a>

## SQL: e9572eab

<pre style="background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;">SQL Digest: e9572eabc1d81a0065a61aa08a38b91fa60045d08bd3758f1c69d6c0fc7c8a2b</pre>

### SQL Statement

```sql
select count(*) from orders where o_clerk = 'Clerk#000003519'
```

<a id="index-advice-e9572eabc1d81a0065a61aa08a38b91fa60045d08bd3758f1c69d6c0fc7c8a2b-1"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_o_clerk ON orders(o_clerk);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                          estRows      estCost       task          access object  operator info
HashAgg_27                  1.00         52569425.39   root                         funcs:count(Column#13)-&gt;Column#11
└─TableReader_29            1.00         52567890.45   root                         MppVersion: 3, data:ExchangeSender_28
  └─ExchangeSender_28       1.00         788518339.12  mpp[tiflash]                 ExchangeType: PassThrough
    └─HashAgg_10            1.00         788518339.12  mpp[tiflash]                 funcs:count(1)-&gt;Column#13
      └─Selection_26        1494.15      788516832.17  mpp[tiflash]                 eq(test.orders.o_clerk, &#34;Clerk#000003519&#34;)
        └─TableFullScan_25  15000000.00  752516832.17  mpp[tiflash]  table:orders   keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
select /*+ hypo_index(orders, idx_o_clerk, o_clerk) */ count(*) from orders where o_clerk = 'Clerk#000003519'
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                       estRows  estCost    task       access object                             operator info
HashAgg_15               1.00     21147.43   root                                                 funcs:count(Column#12)-&gt;Column#11
└─IndexReader_16         1.00     19612.49   root                                                 index:HashAgg_7
  └─HashAgg_7            1.00     294155.73  cop[tikv]                                            funcs:count(1)-&gt;Column#12
    └─IndexRangeScan_14  1494.15  262825.14  cop[tikv]  table:orders, index:idx_o_clerk(o_clerk)  range:[&#34;Clerk#000003519&#34;,&#34;Clerk#000003519&#34;], keep order:false</code></pre>

<a id="sql-8b124dc6fbb9f51ba9799f2d6007d0f959d18f00ed15a45b93c0a0451a4a2e68"></a>

## SQL: 8b124dc6

<pre style="background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;">SQL Digest: 8b124dc6fbb9f51ba9799f2d6007d0f959d18f00ed15a45b93c0a0451a4a2e68</pre>

### SQL Statement

```sql
select o_clerk, o_comment from orders where o_custkey in (1051550, 22894, 775994, 32960, 446677) limit 532
```

<a id="index-advice-8b124dc6fbb9f51ba9799f2d6007d0f959d18f00ed15a45b93c0a0451a4a2e68-1"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_o_custkey_o_clerk_o_comment ON orders(o_custkey,o_clerk,o_comment);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                        estRows  estCost       task          access object  operator info
Limit_12                  74.28    46432561.23   root                         offset:0, count:532
└─TableReader_27          74.28    46432561.23   root                         MppVersion: 3, data:ExchangeSender_26
  └─ExchangeSender_26     74.28    696472239.80  mpp[tiflash]                 ExchangeType: PassThrough
    └─Limit_25            74.28    696472239.80  mpp[tiflash]                 offset:0, count:532
      └─TableFullScan_24  74.28    696472239.80  mpp[tiflash]  table:orders   pushed down filter:in(test.orders.o_custkey, 1051550, 22894, 775994, 32960, 446677), keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
select /*+ hypo_index(orders, idx_o_custkey_o_clerk_o_comment, o_custkey,o_clerk,o_comment) */ o_clerk, o_comment from orders where o_custkey in (1051550, 22894, 775994, 32960, 446677) limit 532
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                       estRows  estCost   task       access object                                                                       operator info
Limit_9                  74.28    3277.60   root                                                                                           offset:0, count:532
└─IndexReader_16         74.28    3277.60   root                                                                                           index:Limit_15
  └─Limit_15             74.28    20042.42  cop[tikv]                                                                                      offset:0, count:532
    └─IndexRangeScan_14  74.28    20042.42  cop[tikv]  table:orders, index:idx_o_custkey_o_clerk_o_comment(o_custkey, o_clerk, o_comment)  range:[22894,22894], [32960,32960], [446677,446677], [775994,775994], [1051550,1051550], keep order:false</code></pre>

<a id="sql-d2f47f5e68ed76c64f93f801d83d8ee4b855e2eacb00d0363a058ad5050aacca"></a>

## SQL: d2f47f5e

<pre style="background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;">SQL Digest: d2f47f5e68ed76c64f93f801d83d8ee4b855e2eacb00d0363a058ad5050aacca</pre>

### SQL Statement

```sql
select O_TOTALPRICE from orders where o_totalprice=219853.63 limit 30
```

<a id="index-advice-d2f47f5e68ed76c64f93f801d83d8ee4b855e2eacb00d0363a058ad5050aacca-1"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_o_totalprice ON orders(o_totalprice);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                          estRows      estCost        task          access object  operator info
Limit_12                    1.25         67228764.21    root                         offset:0, count:30
└─TableReader_29            1.25         67228764.21    root                         MppVersion: 3, data:ExchangeSender_28
  └─ExchangeSender_28       1.25         1008431330.78  mpp[tiflash]                 ExchangeType: PassThrough
    └─Limit_27              1.25         1008431330.78  mpp[tiflash]                 offset:0, count:30
      └─Selection_26        1.25         1008431330.78  mpp[tiflash]                 eq(test.orders.o_totalprice, 219853.63)
        └─TableFullScan_25  15000000.00  972431330.78   mpp[tiflash]  table:orders   keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
select /*+ hypo_index(orders, idx_o_totalprice, o_totalprice) */ O_TOTALPRICE from orders where o_totalprice=219853.63 limit 30
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                       estRows  estCost  task       access object                                       operator info
Limit_9                  1.25     34.89    root                                                           offset:0, count:30
└─IndexReader_16         1.25     34.89    root                                                           index:Limit_15
  └─Limit_15             1.25     284.98   cop[tikv]                                                      offset:0, count:30
    └─IndexRangeScan_14  1.25     284.98   cop[tikv]  table:orders, index:idx_o_totalprice(o_totalprice)  range:[219853.63,219853.63], keep order:false</code></pre>

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

<a id="sql-2e3888683ade4cf13ccd94e3608c528f8fafea6102ca880fb829de2026eee6d3"></a>

## SQL: 2e388868

<pre style="background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;">SQL Digest: 2e3888683ade4cf13ccd94e3608c528f8fafea6102ca880fb829de2026eee6d3</pre>

### SQL Statement

```sql
select count(*) from orders where o_totalprice=219853.63 limit 30
```

<a id="index-advice-2e3888683ade4cf13ccd94e3608c528f8fafea6102ca880fb829de2026eee6d3-1"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_o_totalprice ON orders(o_totalprice);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                          estRows      estCost        task          access object  operator info
Limit_12                    1.00         67228826.77    root                         offset:0, count:30
└─StreamAgg_19              1.00         67228826.77    root                         funcs:count(1)-&gt;Column#11
  └─TableReader_48          1.25         67228764.21    root                         MppVersion: 3, data:ExchangeSender_47
    └─ExchangeSender_47     1.25         1008431330.78  mpp[tiflash]                 ExchangeType: PassThrough
      └─Selection_46        1.25         1008431330.78  mpp[tiflash]                 eq(test.orders.o_totalprice, 219853.63)
        └─TableFullScan_45  15000000.00  972431330.78   mpp[tiflash]  table:orders   keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
select /*+ hypo_index(orders, idx_o_totalprice, o_totalprice) */ count(*) from orders where o_totalprice=219853.63 limit 30
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                         estRows  estCost  task       access object                                       operator info
Limit_12                   1.00     75.18    root                                                           offset:0, count:30
└─StreamAgg_37             1.00     75.18    root                                                           funcs:count(Column#14)-&gt;Column#11
  └─IndexReader_38         1.00     25.28    root                                                           index:StreamAgg_18
    └─StreamAgg_18         1.00     347.54   cop[tikv]                                                      funcs:count(1)-&gt;Column#14
      └─IndexRangeScan_36  1.25     284.98   cop[tikv]  table:orders, index:idx_o_totalprice(o_totalprice)  range:[219853.63,219853.63], keep order:false</code></pre>

<a id="sql-a93e251b62c1f1e624e67fa5163d1e3468663f2fa5cb38a239063ac009a0a371"></a>

## SQL: a93e251b

<pre style="background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;">SQL Digest: a93e251b62c1f1e624e67fa5163d1e3468663f2fa5cb38a239063ac009a0a371</pre>

### SQL Statement

```sql
select count(*) from orders where o_totalprice>=219853.63
```

<a id="index-advice-a93e251b62c1f1e624e67fa5163d1e3468663f2fa5cb38a239063ac009a0a371-1"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_o_totalprice ON orders(o_totalprice);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                          estRows      estCost        task          access object  operator info
HashAgg_27                  1.00         67456764.23    root                         funcs:count(Column#13)-&gt;Column#11
└─TableReader_29            1.00         67455229.29    root                         MppVersion: 3, data:ExchangeSender_28
  └─ExchangeSender_28       1.00         1011828421.72  mpp[tiflash]                 ExchangeType: PassThrough
    └─HashAgg_10            1.00         1011828421.72  mpp[tiflash]                 funcs:count(1)-&gt;Column#13
      └─Selection_26        3538560.81   1008431330.78  mpp[tiflash]                 ge(test.orders.o_totalprice, 219853.63)
        └─TableFullScan_25  15000000.00  972431330.78   mpp[tiflash]  table:orders   keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
select /*+ hypo_index(orders, idx_o_totalprice, o_totalprice) */ count(*) from orders where o_totalprice>=219853.63
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                       estRows     estCost       task       access object                                       operator info
HashAgg_15               1.00        58333155.05   root                                                           funcs:count(Column#12)-&gt;Column#11
└─IndexReader_16         1.00        58331620.11   root                                                           index:HashAgg_7
  └─HashAgg_7            1.00        874974269.96  cop[tikv]                                                      funcs:count(1)-&gt;Column#12
    └─IndexRangeScan_14  3538560.81  804343088.80  cop[tikv]  table:orders, index:idx_o_totalprice(o_totalprice)  range:[219853.63,+inf], keep order:false</code></pre>

<a id="sql-4d49791f77e4b9bc1298a0437c7ff2f9af38f84db4cea0cb3aa3dcbfb125b529"></a>

## SQL: 4d49791f

<pre style="background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;">SQL Digest: 4d49791f77e4b9bc1298a0437c7ff2f9af38f84db4cea0cb3aa3dcbfb125b529</pre>

### SQL Statement

```sql
select count(*) from orders where o_custkey = 861145
```

<a id="index-advice-4d49791f77e4b9bc1298a0437c7ff2f9af38f84db4cea0cb3aa3dcbfb125b529-1"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_o_custkey ON orders(o_custkey);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                        estRows      estCost       task          access object  operator info
StreamAgg_12              1.00         48831709.53   root                         funcs:count(1)-&gt;Column#11
└─TableReader_41          14.86        48830968.20   root                         MppVersion: 3, data:ExchangeSender_40
  └─ExchangeSender_40     14.86        732464000.00  mpp[tiflash]                 ExchangeType: PassThrough
    └─Selection_39        14.86        732464000.00  mpp[tiflash]                 eq(test.orders.o_custkey, 861145)
      └─TableFullScan_38  15000000.00  696464000.00  mpp[tiflash]  table:orders   keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
select /*+ hypo_index(orders, idx_o_custkey, o_custkey) */ count(*) from orders where o_custkey = 861145
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                       estRows  estCost  task       access object                                 operator info
StreamAgg_30             1.00     262.68   root                                                     funcs:count(Column#14)-&gt;Column#11
└─IndexReader_31         1.00     212.78   root                                                     index:StreamAgg_11
  └─StreamAgg_11         1.00     3159.96  cop[tikv]                                                funcs:count(1)-&gt;Column#14
    └─IndexRangeScan_29  14.86    2418.63  cop[tikv]  table:orders, index:idx_o_custkey(o_custkey)  range:[861145,861145], keep order:false</code></pre>
