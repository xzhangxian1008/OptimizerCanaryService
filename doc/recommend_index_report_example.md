# Index Recommendation Overview

<table>
<thead><tr><th>SQL</th><th>Total Time</th><th>Count</th><th>Database</th><th>Table</th><th>Est Cost Change</th><th>Est Improvement</th><th>Advice</th></tr></thead>
<tbody>
<tr><td rowspan="2"><a href="#sql-4f4c289e935f25297eae7382ead734c31ddefa109a21248129d622d9ec14a852">4f4c289e</a></td><td rowspan="2">0.06s</td><td rowspan="2">10</td><td>test</td><td>test_left</td><td>7092147.02 -&gt; 4577628.46</td><td>35.4550%</td><td><a href="#index-advice-4f4c289e935f25297eae7382ead734c31ddefa109a21248129d622d9ec14a852-1">detail</a></td></tr>
<tr><td>test</td><td>test_right</td><td>7092147.02 -&gt; 2591323.47</td><td>63.4621%</td><td><a href="#index-advice-4f4c289e935f25297eae7382ead734c31ddefa109a21248129d622d9ec14a852-2">detail</a></td></tr>
<tr><td rowspan="1"><a href="#sql-5c6384fbb4ac24893fc302e62303fc26ae6879537029d1819528e1ba6ef030cb">5c6384fb</a></td><td rowspan="1">0.01s</td><td rowspan="1">9</td><td>test</td><td>test_table</td><td>7338668.79 -&gt; 17.27</td><td>99.9998%</td><td><a href="#index-advice-5c6384fbb4ac24893fc302e62303fc26ae6879537029d1819528e1ba6ef030cb-1">detail</a></td></tr>
<tr><td rowspan="1"><a href="#sql-0101618ec5b3a438eb8a269bad586645551b7cf4befd1a7f7d0b77559b2658e4">0101618e</a></td><td rowspan="1">0.01s</td><td rowspan="1">10</td><td>test</td><td>test_table</td><td>6274552.34 -&gt; 64556.49</td><td>98.9711%</td><td><a href="#index-advice-0101618ec5b3a438eb8a269bad586645551b7cf4befd1a7f7d0b77559b2658e4-1">detail</a></td></tr>
<tr><td rowspan="1"><a href="#sql-092ee3850c12585fd64e0a4378a52ea99f6cd0d9e5a85d026f641b7660364986">092ee385</a></td><td rowspan="1">0.01s</td><td rowspan="1">8</td><td>test</td><td>test_table</td><td>7338911.54 -&gt; 1996.35</td><td>99.9728%</td><td><a href="#index-advice-092ee3850c12585fd64e0a4378a52ea99f6cd0d9e5a85d026f641b7660364986-1">detail</a></td></tr>
<tr><td rowspan="1"><a href="#sql-092f5f2b0ee9feac853046122952c7eef4afd018a6ae04bb927a74e0cf0542d9">092f5f2b</a></td><td rowspan="1">0.00s</td><td rowspan="1">5</td><td>test</td><td>test_table</td><td>6274818.07 -&gt; 3133.56</td><td>99.9501%</td><td><a href="#index-advice-092f5f2b0ee9feac853046122952c7eef4afd018a6ae04bb927a74e0cf0542d9-1">detail</a></td></tr>
<tr><td rowspan="1"><a href="#sql-dd87ef62f889ebb4c4504927daf72e3ef92794032dab6a10e2e6ac621d4ff23e">dd87ef62</a></td><td rowspan="1">0.00s</td><td rowspan="1">5</td><td>test</td><td>test_table</td><td>6305143.53 -&gt; 270582.74</td><td>95.7085%</td><td><a href="#index-advice-dd87ef62f889ebb4c4504927daf72e3ef92794032dab6a10e2e6ac621d4ff23e-1">detail</a></td></tr>
</tbody>
</table>

# Recommendation List

<a id="sql-4f4c289e935f25297eae7382ead734c31ddefa109a21248129d622d9ec14a852"></a>

## SQL: 4f4c289e

<pre style="background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;">SQL Digest: 4f4c289e935f25297eae7382ead734c31ddefa109a21248129d622d9ec14a852</pre>

### SQL Statement

```sql
SELECT count(*) FROM test_left JOIN test_right ON test_left.a = test_right.a WHERE test_left.b = 1
```

<a id="index-advice-4f4c289e935f25297eae7382ead734c31ddefa109a21248129d622d9ec14a852-1"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_b_a ON test_left(b,a);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                         estRows    estCost      task       access object     operator info
HashAgg_13                 1.00       7092147.02   root                         funcs:count(1)-&gt;Column#12
└─HashJoin_17              1000.00    7070680.04   root                         inner join, equal:[eq(test.test_left.a, test.test_right.a)]
  ├─TableReader_32(Build)  100.00     2516720.03   root                         data:Selection_31
  │ └─Selection_31         100.00     37738128.38  cop[tikv]                    eq(test.test_left.b, 1), not(isnull(test.test_left.a))
  │   └─TableFullScan_30   100000.00  27758128.38  cop[tikv]  table:test_left   keep order:false
  └─TableReader_35(Probe)  100000.00  2545843.02   root                         data:Selection_34
    └─Selection_34         100000.00  31851645.27  cop[tikv]                    not(isnull(test.test_right.a))
      └─TableFullScan_33   100000.00  26861645.27  cop[tikv]  table:test_right  keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
SELECT /*+ hypo_index(test_left, idx_b_a, b,a) */ count(*) FROM test_left JOIN test_right ON test_left.a = test_right.a WHERE test_left.b = 1
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                         estRows    estCost      task       access object                         operator info
HashAgg_13                 1.00       4577628.46   root                                             funcs:count(1)-&gt;Column#12
└─HashJoin_17              1000.00    4556161.48   root                                             inner join, equal:[eq(test.test_left.a, test.test_right.a)]
  ├─IndexReader_31(Build)  100.00     2201.47      root                                             index:IndexRangeScan_30
  │ └─IndexRangeScan_30    100.00     20350.00     cop[tikv]  table:test_left, index:idx_b_a(b, a)  range:[1 -inf,1 +inf], keep order:false
  └─TableReader_37(Probe)  100000.00  2545843.02   root                                             data:Selection_36
    └─Selection_36         100000.00  31851645.27  cop[tikv]                                        not(isnull(test.test_right.a))
      └─TableFullScan_35   100000.00  26861645.27  cop[tikv]  table:test_right                      keep order:false</code></pre>

<a id="index-advice-4f4c289e935f25297eae7382ead734c31ddefa109a21248129d622d9ec14a852-2"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_a ON test_right(a);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                         estRows    estCost      task       access object     operator info
HashAgg_13                 1.00       7092147.02   root                         funcs:count(1)-&gt;Column#12
└─HashJoin_17              1000.00    7070680.04   root                         inner join, equal:[eq(test.test_left.a, test.test_right.a)]
  ├─TableReader_32(Build)  100.00     2516720.03   root                         data:Selection_31
  │ └─Selection_31         100.00     37738128.38  cop[tikv]                    eq(test.test_left.b, 1), not(isnull(test.test_left.a))
  │   └─TableFullScan_30   100000.00  27758128.38  cop[tikv]  table:test_left   keep order:false
  └─TableReader_35(Probe)  100000.00  2545843.02   root                         data:Selection_34
    └─Selection_34         100000.00  31851645.27  cop[tikv]                    not(isnull(test.test_right.a))
      └─TableFullScan_33   100000.00  26861645.27  cop[tikv]  table:test_right  keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
SELECT /*+ hypo_index(test_right, idx_a, a) */ count(*) FROM test_left JOIN test_right ON test_left.a = test_right.a WHERE test_left.b = 1
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                         estRows    estCost      task       access object                     operator info
HashAgg_13                 1.00       2591323.47   root                                         funcs:count(1)-&gt;Column#12
└─IndexHashJoin_23         1000.00    2569856.49   root                                         inner join, inner:IndexReader_35, outer key:test.test_left.a, inner key:test.test_right.a, equal cond:eq(test.test_left.a, test.test_right.a)
  ├─TableReader_32(Build)  100.00     2516720.03   root                                         data:Selection_31
  │ └─Selection_31         100.00     37738128.38  cop[tikv]                                    eq(test.test_left.b, 1), not(isnull(test.test_left.a))
  │   └─TableFullScan_30   100000.00  27758128.38  cop[tikv]  table:test_left                   keep order:false
  └─IndexReader_35(Probe)  1000.00    184.04       root                                         index:Selection_34
    └─Selection_34         1000.00    2127.00      cop[tikv]                                    not(isnull(test.test_right.a))
      └─IndexRangeScan_33  1000.00    1628.00      cop[tikv]  table:test_right, index:idx_a(a)  range: decided by [eq(test.test_right.a, test.test_left.a)], keep order:false</code></pre>

<a id="sql-5c6384fbb4ac24893fc302e62303fc26ae6879537029d1819528e1ba6ef030cb"></a>

## SQL: 5c6384fb

<pre style="background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;">SQL Digest: 5c6384fbb4ac24893fc302e62303fc26ae6879537029d1819528e1ba6ef030cb</pre>

### SQL Statement

```sql
select c2 from test_table where c0 = 5 and c1 = 31387
```

<a id="index-advice-5c6384fbb4ac24893fc302e62303fc26ae6879537029d1819528e1ba6ef030cb-1"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_c1_c0_c2 ON test_table(c1,c0,c2);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                     estRows    estCost       task       access object     operator info
TableReader_11         1.00       7338668.79    root                         data:Projection_5
└─Projection_5         1.00       110080000.10  cop[tikv]                    test.test_table.c2
  └─Selection_10       1.00       110080000.00  cop[tikv]                    eq(test.test_table.c0, 5), eq(test.test_table.c1, 31387)
    └─TableFullScan_9  320000.00  78144000.00   cop[tikv]  table:test_table  keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
select /*+ hypo_index(test_table, idx_c1_c0_c2, c1,c0,c2) */ c2 from test_table where c0 = 5 and c1 = 31387
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                    estRows  estCost  task       access object                                     operator info
IndexReader_9         1.00     17.27    root                                                         index:Projection_5
└─Projection_5        1.00     227.41   cop[tikv]                                                    test.test_table.c2
  └─IndexRangeScan_8  1.00     227.31   cop[tikv]  table:test_table, index:idx_c1_c0_c2(c1, c0, c2)  range:[31387 5,31387 5], keep order:false</code></pre>

<a id="sql-0101618ec5b3a438eb8a269bad586645551b7cf4befd1a7f7d0b77559b2658e4"></a>

## SQL: 0101618e

<pre style="background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;">SQL Digest: 0101618ec5b3a438eb8a269bad586645551b7cf4befd1a7f7d0b77559b2658e4</pre>

### SQL Statement

```sql
select * from test_table where c2=77053
```

<a id="index-advice-0101618ec5b3a438eb8a269bad586645551b7cf4befd1a7f7d0b77559b2658e4-1"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_c2 ON test_table(c2);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                   estRows    estCost      task       access object     operator info
TableReader_7        33.07      6274552.34   root                         data:Selection_6
└─Selection_6        33.07      94112000.00  cop[tikv]                    eq(test.test_table.c2, 77053)
  └─TableFullScan_5  320000.00  78144000.00  cop[tikv]  table:test_table  keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
select /*+ hypo_index(test_table, idx_c2, c2) */ * from test_table where c2=77053
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                         estRows  estCost   task       access object                       operator info
IndexLookUp_7              33.07    64556.49  root                                           
├─IndexRangeScan_5(Build)  33.07    6728.84   cop[tikv]  table:test_table, index:idx_c2(c2)  range:[77053,77053], keep order:false
└─TableRowIDScan_6(Probe)  33.07    8074.61   cop[tikv]  table:test_table                    keep order:false</code></pre>

<a id="sql-092ee3850c12585fd64e0a4378a52ea99f6cd0d9e5a85d026f641b7660364986"></a>

## SQL: 092ee385

<pre style="background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;">SQL Digest: 092ee3850c12585fd64e0a4378a52ea99f6cd0d9e5a85d026f641b7660364986</pre>

### SQL Statement

```sql
select c1 from test_table where c0 = 5 and c2 > 6348
```

<a id="index-advice-092ee3850c12585fd64e0a4378a52ea99f6cd0d9e5a85d026f641b7660364986-1"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_c0_c2_c1 ON test_table(c0,c2,c1);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                     estRows    estCost       task       access object     operator info
TableReader_11         115.58     7338911.54    root                         data:Projection_5
└─Projection_5         115.58     110080011.53  cop[tikv]                    test.test_table.c1
  └─Selection_10       115.58     110080000.00  cop[tikv]                    eq(test.test_table.c0, 5), gt(test.test_table.c2, 6348)
    └─TableFullScan_9  320000.00  78144000.00   cop[tikv]  table:test_table  keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
select /*+ hypo_index(test_table, idx_c0_c2_c1, c0,c2,c1) */ c1 from test_table where c0 = 5 and c2 > 6348
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                    estRows  estCost   task       access object                                     operator info
IndexReader_9         115.58   1996.35   root                                                         index:Projection_5
└─Projection_5        115.58   26283.72  cop[tikv]                                                    test.test_table.c1
  └─IndexRangeScan_8  115.58   26272.19  cop[tikv]  table:test_table, index:idx_c0_c2_c1(c0, c2, c1)  range:(5 6348,5 +inf], keep order:false</code></pre>

<a id="sql-092f5f2b0ee9feac853046122952c7eef4afd018a6ae04bb927a74e0cf0542d9"></a>

## SQL: 092f5f2b

<pre style="background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;">SQL Digest: 092f5f2b0ee9feac853046122952c7eef4afd018a6ae04bb927a74e0cf0542d9</pre>

### SQL Statement

```sql
select c0, c2 from test_table where c1 in (31387, 44843, 26964, 46643, 33079)
```

<a id="index-advice-092f5f2b0ee9feac853046122952c7eef4afd018a6ae04bb927a74e0cf0542d9-1"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_c1_c0_c2 ON test_table(c1,c0,c2);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                     estRows    estCost      task       access object     operator info
TableReader_11         161.60     6274818.07   root                         data:Projection_5
└─Projection_5         161.60     94112032.25  cop[tikv]                    test.test_table.c0, test.test_table.c2
  └─Selection_10       161.60     94112000.00  cop[tikv]                    in(test.test_table.c1, 31387, 44843, 26964, 46643, 33079)
    └─TableFullScan_9  320000.00  78144000.00  cop[tikv]  table:test_table  keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
select /*+ hypo_index(test_table, idx_c1_c0_c2, c1,c0,c2) */ c0, c2 from test_table where c1 in (31387, 44843, 26964, 46643, 33079)
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                    estRows  estCost   task       access object                                     operator info
IndexReader_9         161.60   3133.56   root                                                         index:Projection_5
└─Projection_5        161.60   36764.62  cop[tikv]                                                    test.test_table.c0, test.test_table.c2
  └─IndexRangeScan_8  161.60   36732.36  cop[tikv]  table:test_table, index:idx_c1_c0_c2(c1, c0, c2)  range:[26964,26964], [31387,31387], [33079,33079], [44843,44843], [46643,46643], keep order:false</code></pre>

<a id="sql-dd87ef62f889ebb4c4504927daf72e3ef92794032dab6a10e2e6ac621d4ff23e"></a>

## SQL: dd87ef62

<pre style="background: transparent; border: none; box-shadow: none; padding: 0; white-space: pre; overflow-x: auto; font-family: monospace;">SQL Digest: dd87ef62f889ebb4c4504927daf72e3ef92794032dab6a10e2e6ac621d4ff23e</pre>

### SQL Statement

```sql
select count(*) from test_table where c1>=77641
```

<a id="index-advice-dd87ef62f889ebb4c4504927daf72e3ef92794032dab6a10e2e6ac621d4ff23e-1"></a>

### Advice

#### Create Index Statement

```sql
CREATE INDEX idx_c1 ON test_table(c1);
```

#### Old Plan

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                        estRows    estCost      task       access object     operator info
HashAgg_13                1.00       6305143.53   root                         funcs:count(Column#7)-&gt;Column#6
└─TableReader_14          1.00       6303608.59   root                         data:HashAgg_6
  └─HashAgg_6             1.00       94554097.11  cop[tikv]                    funcs:count(1)-&gt;Column#7
    └─Selection_12        22073.64   94112000.00  cop[tikv]                    ge(test.test_table.c1, 77641)
      └─TableFullScan_11  320000.00  78144000.00  cop[tikv]  table:test_table  keep order:false</code></pre>

#### New Plan

```sql
-- Sql with hypo_index hint
select /*+ hypo_index(test_table, idx_c1, c1) */ count(*) from test_table where c1>=77641
```

<pre style="white-space: pre; overflow-x: auto; max-width: 100%;"><code style="white-space: pre; overflow-wrap: normal; word-break: normal;">id                       estRows   estCost     task       access object                       operator info
HashAgg_12               1.00      270582.74   root                                           funcs:count(Column#7)-&gt;Column#6
└─IndexReader_13         1.00      269047.80   root                                           index:HashAgg_6
  └─HashAgg_6            1.00      4035685.34  cop[tikv]                                      funcs:count(1)-&gt;Column#7
    └─IndexRangeScan_11  22073.64  3593588.23  cop[tikv]  table:test_table, index:idx_c1(c1)  range:[77641,+inf], keep order:false</code></pre>
