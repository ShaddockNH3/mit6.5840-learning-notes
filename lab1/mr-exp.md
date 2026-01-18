# mr

这个东西到底要我们干啥？

配环境和跑通 demo 测试就不说了。

## 要做什么

在开始之前，你应该：

1. 看 src/main/mrsequential.go。这是答案的串行版本，看如何读取插件（Plugin）里的 Map 和 Reduce 函数的；如何读取输入文件内容；最关键是看如何进行排序的
2. 看 mrapps/wc.go。理解 Map 和 Reduce 函数的输入参数类型（KeyValue）和返回值类型。

在看完这两个之后，应该：

1. 定义定义沟通语言 RPC。打开 mr/rpc.go，定义 Args（请求参数）和 Reply（响应参数）结构体。思考 Worker 找 Coordinator 要任务时，需要发什么数据？Coordinator 回复任务时，需要给什么数据（文件名？任务类型？任务编号？nReduce的数量？）字段名首字母必须大写。
2. 打通“电话线”。修改 mr/worker.go：让它发一个 RPC 请求。修改 mr/coordinator.go：让它接收请求，并随便返回一个还没有开始的输入文件名（比如 pg-being_ernest.txt）。修改 mr/worker.go：收到文件名后，打印出来看看，证明连通了。
3. 实现 Map 任务的数据存储。在 Worker 里，拿到文件名后，调用 Map 函数（参考 mrsequential.go）。得到一堆 Key-Value 后，利用 ihash(key) 函数（原文提到在 worker.go 中）计算它属于哪个 Bucket。使用 json.NewEncoder 把结果写进 mr-X-Y 格式的文件里。
4. 实现 Reduce 任务逻辑。Worker 端：能够处理 Coordinator 发来的 "Reduce" 类型的任务。读取数据：Worker 需要去读取所有 Map 任务生成的对应这个 Reduce 编号的文件 (例如：如果是 Reduce 任务 0，就要读 mr-0-0, mr-1-0, mr-2-0...)。排序与聚合：把读到的数据排序（原文说 can steal some code from mrsequential.go for sorting），然后喂给插件里的 Reduce 函数。输出：把结果写到 mr-out-X。
5. Coordinator 的状态流转。Coordinator 数据结构来记录：Map 阶段结束了吗？只有当所有 Map 任务都由 "Idle" 变成 "Completed" 之后，如果有 Worker 来要任务，Coordinator 才能开始分发 Reduce 任务。（这里不需要复杂的容错，但必须要有“Map 做完才能做 Reduce”的逻辑判断）
6. 作业完成与退出。实现 Done() 方法：检查是否所有 Reduce 任务也都做完了？如果是，返回 true。

让 Worker 知道什么时候退场：原文建议如果连不上 Coordinator 就可以退出了，或者设计一个伪任务 "PleaseExit"。

以下内容暂时不需要纠结：

1. 崩溃恢复：这个逻辑比较复杂，建议等基本的 Map 和 Reduce 跑通了再加。
2. 原子重命名：原文提到的 ioutil.TempFile 和 os.Rename 技巧，是为了防止崩溃时留下烂文件。现在调试阶段，可以先直接写文件，功能跑通了再由这部分优化。
