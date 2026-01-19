# mr-arr

## 第一次尝试

最后的结果是：[test1](./test1.md)

### 拿到题目

拿到题目之后，应该直接去看官方文档：[mr](./mr.md)，然后翻译了一下题目，并且理解了一下 题目的意思：[mr-exp](./mr-exp.md)。

在此之后，去理解官方的串行版本实现：[mr-seq](./code/mrsequential.go)以及 key-value 的接口定义：[mr-kv](./code/wc.go)。

然后就开始写代码。

### 设计 RPC

事实上我觉得设计 RPC 是最难的地方，你设计明白了 RPC，就等于把整个系统的设计都想明白了。

首先阅读官网实现：

```go
type ExampleArgs struct {
    X int
}

type ExampleReply struct {
    Y int
}
```

仿照着上面的写真正的 Args 和 Reply：

第一个版本的 Args 带有 WorkerId，State，Category，Reply 带有 TaskNumber，TaskType，FileName，BlockIndex（处理块索引）以及 IntermediateFile（中间文件列表）

首先是针对 Reply 的改进，事实上，如果是 reduce 操作，获取的应该是排好序的中间文件列表，然后进行归并操作。但是其实，中间文件名是有规律的，所以只需要告诉 reduce 任务编号就行了。

然后让 reduce 任务自己去找对应的中间文件，这样的话，就不需要传递大量的文件名字符串了。官网给的中间名称格式是 mr-X-Y，其中

- X 的意思是 map 任务编号
- Y 的意思是 reduce 任务编号

而显然，reduce 任务编号就是 BlockIndex，那么 map 任务编号就是从 0 到 n-1 的整数，其中 n 是 map 任务的总数。

并且，map 操作需要 R 来计算哈希值，所以 Reply 还需要带有 R 的值。

这样，Reply 变成了 TaskNumber，TaskType，FileName，BlockIndex，Nmap，NReduce。

然后细想之下，TaskNumber 是任务编号，同时也可以代表处理的块编号，所以其实可以把 BlockIndex 去掉。

最后，Reply 变成了 TaskNumber，TaskType，FileName，Nmap，NReduce。

其中，TaskNumber 是任务编号，TaskType 是任务类型（Map/Reduce/Wait/AllDone），FileName 是处理的文件名，Nmap 是 map 任务总数，NReduce 是 reduce 任务总数。

然后是 Args 的改进，一开始我是认为 State 和 Category 的意思是说，论文里不是有个要周期性 ping 来确认是否活着吗，然后就保留了。

但是想了一下，感觉作为请求参数有点多余了，这里只是发，都来请求了，所以肯定是待命状态。

所以就把 State 和 Category 去掉了。

然后继续思考 WorkerId 的作用，发现其实根本没用。因为 master 根本不需要知道是谁在请求任务，它只需要给出一个任务，然后知道整个任务做完就行了，所以删掉。

整个时候 Args 变成了空结构体，这样肯定是不对的。

所以重新思考了一下，分配的任务，那自然要有任务编号和任务类型，所以 TaskNumber 和 TaskType 是必须的。

然后我设计了一个 PrevTaskState，代表上一个任务的状态，可能的值有两种，可能的值有 Success，Faild，Timeout 以及我当时想的 FirstTime。

因为第一次请求，肯定得有个状态代表。

但是后来想了一下，FirstTime 其实没有意义，而且字符串匹配并非很高效，所以就改为了只有四个状态。

-1 代表是第一次请求任务，0代表是上次任务失败，1代表是上次任务成功，2代表是上次任务超时。

这个时候就突然意识到了，其实根本不需要四个状态来表达，因为 master 根本不需要关心你是怎么失败的，它只需要知道你上次任务成功了没有就行了。

所以只需要在第一次请求的时候，TaskNumber 设为 -1 就行了。

只需要知道成功与否就行了，不需要关心失败的原因。

因为只要失败了，就会重新分配任务，只要成功了，就不会重新分配任务。

失败的原因不重要，反正处理的结果都是重新分配任务。

然后设计了 TaskState，上次任务的状态，初值为 false

最后，Args 也设计出来了，是 TaskNumber，TaskType，TaskState。

### 跑通与测试 RPC

跑通 RPC，事实上只需要把 worker.go 里 CallExample()的注释删掉即可。

然后在 main 文件夹下，运行：

```bash
go build -buildmode=plugin ../mrapps/wc.go
```

然后运行：

```bash
go run mrcoordinator.go pg-*.txt
```

然后开一个新终端运行：

```bash
go run mrworker.go wc.so
```

如果收到了 100 那就是成功了。

接下来，仿照 CallExample() 研究了它的代码，然后设计了自己的 CallMyExample()，随便赋值跑了一次，跑通即可。

### 写 map/ reduce

接下来就是在 worker.go 的 Worker 函数里写 map 任务。

首先先去 mrworker 里看看逻辑，知道 mapf 和reducef 到底是什么。

首先先初始化参数，最开始应该是初始化一个 args，把 TaskNumber 设为 -1，TaskType 设为空，TaskState 设为 false。

然后写一个大循环，不断地请求任务。

先不断地获取新的 reply，然后根据 reply 里的 TaskType 来决定做什么。

TaskType 有四种可能，Map，Reduce，Wait，AllDone。

这个时候先更新一下 args，更新 TaskNumber 和 TaskType，然后把 TaskState 设为 false，默认失败，除非是完成了才设为 true。

#### map

如果是 map，map 的逻辑是先调用 mapf 函数，得到一堆 KeyValue 后，根据 Key 进行哈希分区，得到属于哪个 Bucket。

然后使用 json.NewEncoder 来写文件，中间文件命名格式为 mr-X-Y。X 是 Map 任务编号，Y 是 Reduce 任务编号

mapf 的传参应该是 filename 和文件内容，文件内容是通过 taskNumber 来算的。

所以第一步是解析 kvs，利用 mapf。

本来在那里研究怎么原子操作，后来想想先放着，先把基本逻辑写完。

一次性创建 nReduce 个文件句柄和对应的 json.Encoder，然后全部打开。

然后对 kvs 进行哈希分区，写入对应的文件后一次性全部关闭，最后把 args.TaskState 设为 true。

#### reduce

Reduce 任务其实就是读取所有的中间文件，一次性打开所有的 mr-X-Y 文件，其中 X 从 0 到 nMap-1，Y 是当前的 reduce 任务编号。

然后解码，把所有的 KeyValue 都放到一个切片里，对 Key 进行排序，最后调用 reducef 函数，得到结果，写入最终文件。

具体过程其实就是一次性打开，解析然后放在切片，之后——

sort 的逻辑直接抄串行实现即可。

然后创建输出文件，命名格式为 mr-out-X，其中 X 是 reduce 任务编号。

利用官网给出的写入逻辑，写入文件后输出，把 args.TaskState 设为 true。

### 处理 Wait 和 AllDone

Wait 就是休息一会儿，因为一定要所有的 map 任务做完，才能做 reduce 任务。

写一个 time.Sleep 即可。

AllDone 就是直接退出循环，结束工作。

### 写 coordinator

一开始看到 coordinator.go 觉得特别复杂，后来发现其实并没有那么复杂。

#### 设计结构体

按照论文里面的来看，coordinator 得记录 WorkerId，State 和 Category。

但是因为 Coordinator 要记录的任务是全局的，所以 Coordinator 结构体里应该有 Map 和 Reduce 两个切片，分别记录所有的 map 任务和 reduce 任务。

所以应该设计一个结构体 AWorker，里面有 WorkerId，State 和 Category。

Map 和 Reduce 是 AWorker 类型的切片。

> 后续在本次设计中增加了 files，mu，和 IsMapDone，见下文

#### 初始化

初始化就是在 MakeCoordinator 里初始化 Map 和 Reduce 切片。

开两个循环，把所有的状态都设为待命状态和对应的 id 以及种类即可。

然后发现其实有传参 file，所以 Coordinator 结构体里应该还有 Files 字段。

#### 主逻辑

设计的是 Handler 函数，用来处理 worker 的请求。

最开始的逻辑是根据 args 里的 TaskNumber 先判断，如果是 -1 代表是第一次请求任务，那么就分配一个新的 map 任务。后来在写出来第一个版本之后发现这样的话还得处理如果 -1 是 map 已经做完了之后再出来的逻辑，很复杂。

最后的第一个版本就直接应该先判断 args.TaskState，如果是 true，那么把表内对任务的状态设为完成状态。因为 TaskNumber = -1 的情况，TaskState 肯定是 false，所以不会影响。

然后下一个判断是判断 TaskNumber!= -1 && !TaskType，这意味着不是第一次任务并且任务失败了，所以得把这个任务的状态设为待命状态，以便重新分配。

上面是一个 if，else，处理上次任务失败还是成功以便修改状态。

然后下面的 if，else 分配任务。

因为要先做 map，所以先判断 map 任务有没有做完。

这个时候为了判断 map 任务做完与否，coordinator 结构体里增加了 IsMapDone 字段，初值为 false。

如果是假，那么遍历 c.Map 然后分配，这个时候就根本不需要关心任务是不是第一次了。

如果发现 c.Map 所有都做完了，那么把 IsMapDone 设为 true，如果不是，那么设置为 Wait。

都做完了，进入下一个状态分配 Reduce。

同样的逻辑，遍历 c.Reduce 分配任务，如果没任务分配，检查所有的任务是否做完，如果没做完设置为 Wait，如果都做完了，那么设置为 AllDone。

那么整个逻辑就写完了

### 简单的优化与测试后的问题

事实上就是非常史的在主逻辑的最开始，写了一个 c.mu.Lock()，然后立刻写了一下defer c.mu.Unlock()。

本来的主逻辑有问题还很抽象，然后优化为了上面所说的版本后，就开始测试了。

测试后发现有几个问题，就详见[test1](./test1.md)了

主要的问题一个就是之前意识到的原子操作，但是周期性 ping 这事，写着写着就忘了，难绷啊。

噢还有一些级别的报错不需要直接让程序崩溃，应该处理一下。

没去处理 wait 的问题，修改为 wait 之后，就直接一直是 wait 了。

## 第二次尝试

其实第二次尝试已经是最后一次尝试了，修了非常多 bug，全部测试样例通过，其实感觉还是有优化空间的。

最后的结果是：[test2](./test2.md)

### 原子操作

文档里用的是 io/ioutil 包，后来发现这个包已经被废弃了，所以改成了 os。

和之前一样的逻辑，只是把一次性创建一堆文件改成了一次性创建一堆临时文件，写完之后再重命名。

reduce 操作也是类似的逻辑。

### 增加 out 机制

这里也是我觉得写的不是很好的地方，我写了一个 islive 函数，判断当前时间和上次心跳时间的差值是否大于 10 秒，如果大于 10 秒就代表这个 worker 死掉了，然后修改状态为待命状态。

主函数则是在每次分配任务之后，直接 go islive。

这样的话很耗资源就是了，感觉应该写一个函数周期性的去检查，这样会更好。

然后随之而来的就是竞态问题，所以把 islive 里的修改状态的部分也加了锁。

还有一个问题就是防止二次汇报成功，所以只需要增加一个检查即可，只能修改非 complete 状态的任务。

### done 的问题

随之而来的就是 done 的问题，这里自作聪明加了个锁，然后就死锁了。

所以删了。

### 处理 wait 机制

wait 机制之前的问题很大，主要有两个问题。

第一，停止完之前的 task 还是原本的 task，分配到 wait 之后，再次进入主逻辑，会再次进入一次 complete 任务的逻辑，导致任务被重复完成从而修改状态为 complete。（虽然不影响，但是感觉会出问题，而且我逻辑改过了，这样变成 complete 的逻辑很怪）修改完之后出现第二个问题。

第二，wait 之后永远是 wait。

所以这部分进行了简单的重构，如果进入 wait 状态，直接在 worker 那里改为是第一次来请求任务，这样就不用去处理任务的状态。

写这个的时候主要是忘记了 index 问题，忘了 map 和 reduce 的情况需要分开处理。

### 处理 map 和 reduce 的衔接

原本的逻辑是判断完 map 任务做完，设置 IsMapDone 为 true。

这部分逻辑是对的，但是我采用的是 if else 结构，所以当 map 做完之后，reduce 任务就会先进入判断 wait 以及 all done 的逻辑，从而产生奇怪的问题（和在不进入这两个状态就把状态写死为 wait 的逻辑联合产生的）

改成两个独立的 if 语句就可以解决，顺便把 all done 的逻辑改为了 if else，而非 if，然后直接执行。

## 后续

问题仍然有一些，比如说我并没有很好的去处理 log，几乎是完全没写 log。

而且一出现问题直接 fatal，例如读取文件失败直接 fatal，这样其实并不好。

噢还有之前说的大锁问题，锁的颗粒度太大，应该改进。

总的来说先过所有测试再说。
