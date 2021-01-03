## Question

两个 csv 文件，分别命名为 t1, t2，都由 a, b 两个整数列构成，每个文件的数据行数至少百万以上，请设计并实现一个算法快速计算出：select count(*) from t1 join t2 on t1.a = t2.a and t1.b > t2.b 要求：

1. 充分利用机器的 CPU/MEM 资源，计算速度越快越好
2. 分析并测试该算法对机器资源的使用情况
3. 请分析并测试各个表中 a 的 number of distinct values 对该算法性能的影响
4. 请思考并列举出所有能想到的做法，越多越好。描述他们的优劣点，适用的场景，什么场景下执行的快，什么场景下执行的慢

## Directory

文件内容：

- data_gen.py 可通过 `python3 data_gen.py --t1-size xxx --t1-conflict-ratio yyy -t2-size zzz ` 在当前目录下生成测试数据集：创建 `data` 文件夹并生成 t1-x.csv 和 t2-x.csv 文件（每个文件中有 125, 000 条数据），`--t1(2)-size` 为 optional，指示 t1/t2-x.csv 文件共有多少行，默认 1,000,000 行，`--t1-conflict-ratio` 也为 optional，指示 t1 中 a 列 distinct values(t1_size / t1_conflict_ratio) 的数目，默认为 100。
- smpworker.go 为采用 SMP 架构的算法实现：从 t1 和 t2 文件中选出小的文件来通过取模（modulo 为 101）运算创建 hash table，大的文件同时从文件中读出内容，创建完成后，将大的文件中读出的行与创建好的 hash table 进行匹配。
- smpworker_test.go 为测试文件。

## Answer

- 分析并测试该算法对机器资源的使用情况

  目前内存中需要维护小的 csv 文件的 hash table，hash table 建立好后大的 csv 文件即可开始匹配。对于大的 csv 文件，因为给每一个 chunk（tx-y.csv）分配了一个 goroutine，memory 中只需要能提供一个 chunk 的 hash table 所需空间即可正常执行。通过 `go test -bench SMP` 可以查看当前算法对机器资源的使用情况：当 t1 与 t2 各有一百万行，且 a 的 number of distinct values 大概为 100 时，需要 3.6 秒（8 核）的运行时间与 250 MB 左右的 memory。

  ```
  goos: linux
  goarch: amd64
  BenchmarkSMP-8   	       1	3600461422 ns/op	252734112 B/op	 4020558 allocs/op
  ```

- 请分析并测试各个表中 a 的 number of distinct values 对该算法性能的影响

  假设 t2 的数据规模 <= t1，则当前算法会**选择 t2 建立 hash table**，当 t2 中的 number of distinct values 的数目 >> hashLimit（modulo）时，hashLimit 足够大，则可以保证 t2 在 hash table 的每个 slot 中分布足够均匀，此时 t1 number of distinct values 的变化对算法性能影响不大；

  当 t2 中的 number of distinct values 的数目（设为 $d_2$） $\approx$ hashLimit（modulo）时，假设 $y_i$ 为 t1 在第 i 个 slot 中出现的频率，可有 $y_1+...y_{\mbox{modulo}} = 1$，则 t1 与 t2 匹配需要 $\frac{y_1+...+y_{d_2}}{d_2}t_1t_2$ 次，假设 t1 均匀分布，则可推出需要 $\frac{t_1t_2}{\mbox{modulo}}$，**因此 t2 number of distinct values 的变化对于算法影响不大。**

  当 t1 中的 number of distinct values 的数目 > hashLimit（modulo）时，hashLimit 足够大，则可以保证 t2 在 hash table 中分布足够均匀，对算法性能应该影响不大

  当 t1 与 t2 均有 5,000,000 行时，且 t2 number of distinct values  为 5000 时
  
  | t1 number of distinct values | time to finish |
  | :--------------------------: | :------------: |
  |              20              |    93.240s     |
  |              50              |    90.379s     |
  |             100              |    90.834s     |
  |             500              |    94.608s     |
  |             1000             |    96.929s     |
  |             5000             |    98.780s     |
  

当 t1 与 t2 均有 5,000,000 行时，且 t2 number of distinct values  为 500 时

| t1 number of distinct values | time to finish |
| :--------------------------: | :------------: |
|              20              |    161.580s    |
|              50              |    162.514s    |
|             100              |    176.292s    |
|             500              |    181.825s    |
|             1000             |    123.348s    |
|             5000             |    89.618s     |

当 t1 与 t2 均有 5,000,000 行时，且 t2 number of distinct values  为 50 时

| t1 number of distinct values | time to finish |
| :--------------------------: | :------------: |
|              20              |    350.046s    |
|              50              |    351.503s    |
|             100              |    175.755s    |
|             500              |    101.284s    |
|             1000             |    92.684s     |
|             5000             |    83.845s     |

- 请思考并列举出所有能想到的做法，越多越好。描述他们的优劣点，适用的场景，什么场景下执行的快，什么场景下执行的慢

  |                             做法                             |                             优点                             |                             缺点                             |                             场景                             |
  | :----------------------------------------------------------: | :----------------------------------------------------------: | :----------------------------------------------------------: | :----------------------------------------------------------: |
  |                       SMP（当前算法）                        | 模型简单。executor 间共享 memory，disk 资源，访问另一个 executor 资源的速度最快 |               只利用一台机器的性能，伸缩性最差               | 数据量小时（~GB）适用，因为模型简单，可快速实现。当要求高并发时会因为共享资源出现大量的 contention 导致性能严重降低。 |
  | MPP：将 t1 和 t2 分别在一台机器上运行，当数据少的一方先完成 hash table 的建立时，另一方可以直接将当前处理的数据 on the fly 传送到已经处理完的机器上进行匹配，并将以前已经处理的数据 batch 传送到另一台机器上。如果数据规模实在太大，则需要对 t1，t2 任务本身进行切分 | 将计算扩展到整个集群，每个 executor 都有自己的 CPU，memory，disk 资源，可以通过 network message passing 或通过高速 interconnects 连接 disk 资源来实现 executor 之间的交互。执行效率（单节点吞吐率）、并发能力高 | 因为每个 executor 一般都只会根据自己的本地数据进行计算，会因为 stragglers 降低整个集群性能 | 数据量较大（GB~TB）时适用，根据 HAWQ 的博客[2]，传统的 MPP 架构在集群规模上存在一定限制的，10-18 个 executor 并发执行可得到最高性能 |
  | Batch Processing：在 MPP 的基础上加上全局 scheduler 以实现更高的 throughput | 有 scheduler 可以通过类似于 MapReduce 的 backup tasks 来缓解 stragglers，因而相较 MPP 伸缩性有非常大的变化。将计算和存储分离（虽然尽量保证 locality，但是也可以来自其他存储节点）。 |                  执行效率（单节点吞吐率）低                  |                   数据量极大（TB~）时适用                    |

## Reference

[1] SMP and MPP https://cloudblogs.microsoft.com/sqlserver/2014/07/30/transitioning-from-smp-to-mpp-the-why-and-the-how/

[2] HAWQ, MPP and Batch Processing https://laptrinhx.com/apache-hawq-next-step-in-massively-parallel-processing-3942821226/

[3] Golang test https://books.studygolang.com/The-Golang-Standard-Library-by-Example/chapter09/09.2.html



