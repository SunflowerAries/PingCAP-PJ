两个 csv 文件，分别命名为 t1, t2，都由 a, b 两个整数列构成，每个文件的数据行数至少百万以上，请设计并实现一个算法快速计算出：select count(*) from t1 join t2 on t1.a = t2.a and t1.b > t2.b 要求：

1. 充分利用机器的 CPU/MEM 资源，计算速度越快越好
2. 分析并测试该算法对机器资源的使用情况
3. 请分析并测试各个表中 a 的 number of distinct values 对该算法性能的影响
4. 请思考并列举出所有能想到的做法，越多越好。描述他们的优劣点，适用的场景，什么场景下执行的快，什么场景下执行的慢

Hive: SQL->MapReduce

illustrate mpp backwards about stragglers https://laptrinhx.com/apache-hawq-next-step-in-massively-parallel-processing-3942821226/

MPP stragglers: If one node is constantly performing slower than the others, the whole engine performance is limited by the performance of this problematic node, regardless the cluster size.

**hash join**

smaller one as outer: litte refinement

hash table to smaller one(tradeoff: the effective block size of R is reduced)