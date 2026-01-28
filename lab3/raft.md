# 6.5840 Lab 3: Raft (Raft 共识算法)

## 介绍

这是一系列实验中的第一个，在这些实验中，你将构建一个容错的键/值（Key/Value）存储系统。在这个实验中，你将实现 **Raft**，一种复制状态机（Replicated State Machine）协议。在下一个实验中，你将在 Raft 之上构建一个键/值服务。之后，你将把服务“分片（shard）”到多个复制状态机上，以获得更高的性能。

复制服务通过在多个副本服务器上存储其状态（即数据）的完整副本来实现容错。即使部分服务器发生故障（崩溃、网络中断或不稳定），复制也能让服务继续运行。挑战在于，故障可能会导致副本持有不同的数据副本。

Raft 将客户端请求组织成一个序列，称为**日志（log）**，并确保所有副本服务器看到相同的日志。每个副本按日志顺序执行客户端请求，将其应用（apply）到服务状态的本地副本中。由于所有活跃的副本都看到相同的日志内容，它们都以相同的顺序执行相同的请求，从而继续拥有相同的服务状态。如果服务器发生故障但后来恢复了，Raft 会负责将其日志更新到最新状态。只要至少**大多数（majority）**服务器处于活跃状态且能相互通信，Raft 就会继续运行。如果没有达到大多数，Raft 将不会取得进展，但一旦大多数服务器能再次通信，它将从中断的地方继续运行。

在这个实验中，你将把 Raft 实现为一个具有相关方法的 Go 对象类型，旨在作为更大服务中的一个模块使用。一组 Raft 实例通过 RPC 相互通信以维护复制的日志。你的 Raft 接口将支持无限序列的编号命令，也称为**日志条目（log entries）**。条目使用索引号（index numbers）进行编号。具有给定索引的日志条目最终会被**提交（committed）**。此时，你的 Raft 应该将该日志条目发送给更大的服务以供其执行。

你应该遵循 [Raft 扩展论文](http://nil.csail.mit.edu/6.824/2022/papers/raft-extended.pdf) 中的设计，特别注意**图 2**。你将实现论文中的大部分内容，包括保存持久状态，并在节点故障重启后读取它。你**不需要**实现集群成员变更（第 6 节）。

本实验分为四个部分。你必须在相应的截止日期前提交每一部分。

## 入门指南

执行 `git pull` 以获取最新的实验软件。

如果你已经完成了 Lab 1，你应该已经有了实验源代码的副本。如果没有，你可以在 Lab 1 的说明中找到通过 git 获取源代码的指引。

我们为你提供了骨架代码 `src/raft/raft.go`。我们还提供了一套测试，你应该使用这些测试来驱动你的实现工作，我们将使用这些测试来评估你提交的实验。测试位于 `src/raft/raft_test.go` 中。

当我们给你的提交评分时，我们将运行不带 `-race` 标志的测试。但是，你应该在开发解决方案时通过带 `-race` 标志运行测试，以检查你的代码是否存在数据竞争。

要开始运行，请执行以下命令。别忘了 `git pull` 获取最新软件。

```bash
$ cd ~/6.5840
$ git pull
...
$ cd src/raft
$ go test
Test (3A): initial election (reliable network)...
Fatal: expected one leader, got none
--- FAIL: TestInitialElection3A (4.90s)
Test (3A): election after network failure (reliable network)...
Fatal: expected one leader, got none
--- FAIL: TestReElection3A (5.05s)
...
$
```

## 代码

通过向 `raft/raft.go` 添加代码来实现 Raft。在该文件中，你会找到骨架代码，以及如何发送和接收 RPC 的示例。

你的实现必须支持以下接口，测试器和（最终的）键/值服务器将使用该接口。你可以在 `raft.go` 的注释中找到更多细节。

```go
// 创建一个新的 Raft 服务器实例：
rf := Make(peers, me, persister, applyCh)

// 开始就一个新的日志条目达成共识：
rf.Start(command interface{}) (index, term, isleader)

// 询问 Raft 其当前的任期，以及它是否认为自己是 Leader
rf.GetState() (term, isLeader)

// 每当一个新的条目被提交到日志时，每个 Raft 节点
// 都应该向服务（或测试器）发送一个 ApplyMsg。
type ApplyMsg
```

服务调用 `Make(peers, me, …)` 来创建一个 Raft 节点（Peer）。`peers` 参数是 Raft 节点（包括这一个）的网络标识符数组，用于 RPC。`me` 参数是此节点在 `peers` 数组中的索引。`Start(command)` 请求 Raft 开始处理将命令追加到复制日志中。`Start()` 应该立即返回，而不等待日志追加完成。服务期望你的实现对于每个新提交的日志条目，都向传递给 `Make()` 的 `applyCh` 通道参数发送一个 `ApplyMsg`。

`raft.go` 包含发送 RPC（`sendRequestVote()`）和处理传入 RPC（`RequestVote()`）的示例代码。你的 Raft 节点应该使用 `labrpc` Go 包（源码在 `src/labrpc`）交换 RPC。测试器可以告诉 `labrpc` 延迟 RPC、重新排序 RPC 并丢弃 RPC，以模拟各种网络故障。虽然你可以临时修改 `labrpc`，但请确保你的 Raft 能与原始的 `labrpc` 一起工作，因为我们将使用它来测试和评分你的实验。你的 Raft 实例必须仅通过 RPC 进行交互；例如，不允许它们使用共享的 Go 变量或文件进行通信。

后续实验建立在本实验之上，因此给自己足够的时间编写稳固的代码非常重要。

---

## 第 3A 部分：Leader 选举（中等难度）

实现 Raft Leader 选举和心跳（没有日志条目的 `AppendEntries` RPC）。3A 部分的目标是选出一个 Leader，如果没有故障，该 Leader 应保持其地位；如果旧 Leader 故障或者往返旧 Leader 的数据包丢失，则由新 Leader 接管。运行 `go test -run 3A` 来测试你的 3A 代码。

1. 你无法轻松地直接运行你的 Raft 实现；相反，你应该通过测试器运行它，即 `go test -run 3A`。
2. 遵循论文的**图 2**。此时你需要关注发送和接收 `RequestVote` RPC、与选举相关的服务器规则以及与 Leader 选举相关的状态。
3. 将图 2 中用于 Leader 选举的状态添加到 `raft.go` 中的 `Raft` 结构体中。你还需要定义一个结构体来保存有关每个日志条目的信息。
4. 填充 `RequestVoteArgs` 和 `RequestVoteReply` 结构体。修改 `Make()` 创建一个后台 goroutine，当它在一段时间内没有收到其他节点的消息时，通过发送 `RequestVote` RPC 来周期性地启动 Leader 选举。实现 `RequestVote()` RPC 处理程序，以便服务器可以相互投票。
5. 要实现心跳，定义一个 `AppendEntries` RPC 结构体（尽管你可能还不需要所有参数），并让 Leader 周期性地发送它们。编写一个 `AppendEntries` RPC 处理方法。
6. 测试器要求 Leader 每秒发送心跳 RPC 的次数**不超过十次**。
7. 测试器要求你的 Raft 在旧 Leader 故障后的**五秒内**选出一个新 Leader（如果大多数节点仍能通信）。
8. 论文第 5.2 节提到选举超时范围为 150 到 300 毫秒。只有当 Leader 发送心跳的频率远高于每 150 毫秒一次（例如，每 10 毫秒一次）时，这样的范围才有意义。因为测试器限制你每秒只能发送几十次心跳，所以你必须使用比论文中的 150 到 300 毫秒**更大**的选举超时，但也不能太大，否则你可能无法在五秒内选出 Leader。
9. 你可能会发现 Go 的 `rand` 包很有用。
10. 你需要编写周期性执行或延时执行的代码。最简单的方法是创建一个带有循环的 goroutine，并在循环中调用 `time.Sleep()`；参见 `Make()` 创建的 `ticker()` goroutine。**不要使用** Go 的 `time.Timer` 或 `time.Ticker`，它们很难正确使用。
11. 如果你的代码难以通过测试，请再次阅读论文的图 2；Leader 选举的完整逻辑分布在图中的多个部分。
12. 别忘了实现 `GetState()`。
13. 当测试器永久关闭一个实例时，会调用你的 Raft 的 `rf.Kill()`。你可以使用 `rf.killed()` 检查 `Kill()` 是否已被调用。你可能希望在所有循环中都这样做，以避免已死亡的 Raft 实例打印令人困惑的消息。
14. Go RPC 仅发送名称以**大写字母**开头的结构体字段。子结构体的字段名也必须大写（例如数组中的日志记录字段）。`labgob` 包会对此发出警告；不要忽略这些警告。
15. 本实验最具挑战性的部分可能是调试。花点时间让你的实现易于调试。请参阅“指导”页面获取调试提示。
16. 如果你未通过测试，测试器会生成一个文件，可视化显示标记了事件的时间轴，包括网络分区、崩溃的服务器和执行的检查。这里有一个可视化的示例。此外，你可以通过编写代码添加自己的注释，例如 `tester.Annotate("Server 0", "short description", "details")`。这是我们今年添加的新功能，如果你对可视化器有任何反馈（例如错误报告、你认为有帮助的注释 API、你希望可视化器显示的信息等），请告诉我们！

在提交 3A 部分之前，请确保你通过了 3A 测试，以便看到类似以下内容：

```bash
$ go test -run 3A
Test (3A): initial election (reliable network)...
  ... Passed --   3.6  3   106    0
Test (3A): election after network failure (reliable network)...
  ... Passed --   7.6  3   304    0
Test (3A): multiple elections (reliable network)...
  ... Passed --   8.4  7   954    0
PASS
ok      6.5840/raft1    19.834s
```

每个 "Passed" 行包含五个数字；分别是测试所用的时间（秒）、Raft 节点的数量、测试期间发送的 RPC 数量、RPC 消息的总字节数以及 Raft 报告已提交的日志条目数。你的数字将与此处显示的有所不同。如果你愿意，可以忽略这些数字，但它们可能会帮助你对实现发送的 RPC 数量进行健全性检查。对于所有的实验 3、4 和 5，如果所有测试（`go test`）总共花费超过 600 秒，或者任何单个测试花费超过 120 秒，评分脚本将判定你的解决方案失败。

当我们给你的提交评分时，我们将运行不带 `-race` 标志的测试。但是，你应该确保你的代码在带 `-race` 标志的情况下能持续通过测试。

---

## 第 3B 部分：日志（困难）

实现 Leader 和 Follower 代码以追加新的日志条目，从而使 `go test -run 3B` 测试通过。

1. 运行 `git pull` 以获取最新的实验软件。
2. Raft 日志是 **1 起始索引（1-indexed）**的，但我们建议你将其视为 0 起始索引，并以一个任期（term）为 0 的条目（在索引 = 0 处）开始。这允许第一个 `AppendEntries` RPC 包含 0 作为 `PrevLogIndex`，并且是日志中的有效索引。
3. 你的第一个目标应该是通过 `TestBasicAgree3B()`。首先实现 `Start()`，然后按照图 2 编写代码，通过 `AppendEntries` RPC 发送和接收新的日志条目。在每个节点上通过 `applyCh` 发送每个新提交的条目。
4. 你需要实现**选举限制**（论文第 5.4.1 节）。
5. 你的代码可能有重复检查某些事件的循环。不要让这些循环在不暂停的情况下持续执行，因为这会极大地拖慢你的实现，导致测试失败。使用 Go 的条件变量（condition variables），或者在每次循环迭代中插入 `time.Sleep(10 * time.Millisecond)`。
6. 为了未来的实验着想，请编写（或重写）干净清晰的代码。如果需要思路，请重温我们的“指导”页面，了解有关如何开发和调试代码的提示。
7. 如果你未通过测试，请查看 `raft_test.go` 并从那里追踪测试代码，以了解正在测试的内容。
8. 即将到来的实验测试可能会因为你的代码运行太慢而失败。你可以使用 `time` 命令检查你的解决方案使用了多少实际时间和 CPU 时间。典型的输出如下：

```bash
$ time go test -run 3B
Test (3B): basic agreement (reliable network)...
  ... Passed --   1.3  3    18    0
Test (3B): RPC byte count (reliable network)...
  ... Passed --   2.8  3    56    0
...
PASS
ok      6.5840/raft1    48.353s
go test -run 3B  1.37s user 0.74s system 4% cpu 48.865 total
```

"ok 6.5840/raft 35.557s" 意味着 Go 测量的 3B 测试花费了 35.557 秒的实际（墙上时钟）时间。"user 0m2.556s" 意味着代码消耗了 2.556 秒的 CPU 时间，即实际执行指令的时间（而不是等待或睡眠）。如果你的解决方案在 3B 测试中使用了远超一分钟的实际时间，或远超 5 秒的 CPU 时间，你稍后可能会遇到麻烦。检查是否在睡眠或等待 RPC 超时上花费了时间、是否有在不睡眠或不等待条件/通道消息的情况下运行的循环，或者是否发送了大量的 RPC。

---

## 第 3C 部分：持久化（困难）

如果基于 Raft 的服务器重启，它应该从中断的地方恢复服务。这要求 Raft 保持能在重启后存活的持久状态。论文的图 2 提到了哪些状态应该是持久的。

真实的实现会在每次状态变更时将 Raft 的持久状态写入磁盘，并在重启后从磁盘读取状态。你的实现不使用磁盘；相反，它将从 `Persister` 对象（参见 `persister.go`）保存和恢复持久状态。调用 `Raft.Make()` 的人会提供一个 `Persister`，它最初持有 Raft 最近持久化的状态（如果有）。Raft 应该从该 `Persister` 初始化其状态，并应在每次状态变更时使用它来保存持久状态。使用 `Persister` 的 `ReadRaftState()` 和 `Save()` 方法。

通过添加保存和恢复持久状态的代码来完成 `raft.go` 中的 `persist()` 和 `readPersist()` 函数。你需要将状态编码（或“序列化”）为字节数组，以便将其传递给 `Persister`。使用 `labgob` 编码器；参见 `persist()` 和 `readPersist()` 中的注释。`labgob` 类似于 Go 的 `gob` 编码器，但如果你尝试编码带有小写字段名的结构体，它会打印错误消息。目前，将 `nil` 作为第二个参数传递给 `persister.Save()`。在你的实现更改持久状态的地方插入对 `persist()` 的调用。一旦你完成了这些，并且如果你的其余实现是正确的，你应该通过所有 3C 测试。

你可能需要一种优化，即一次回退 `nextIndex` 超过一个条目。请查看扩展版 Raft 论文的第 7 页底部和第 8 页顶部（由灰线标记）。论文对细节描述得很模糊；你需要填补空白。一种可能性是让拒绝消息包含：

* `XTerm`: 冲突条目的任期（如果有）
* `XIndex`: 该任期的第一个条目的索引（如果有）
* `XLen`: 日志长度

然后 Leader 的逻辑可以是这样的：

* **情况 1**: Leader 没有 `XTerm`：
  `nextIndex = XIndex`
* **情况 2**: Leader 有 `XTerm`：
  `nextIndex = (Leader 对 XTerm 的最后一个条目的索引) + 1`
* **情况 3**: Follower 的日志太短：
  `nextIndex = XLen`

其他一些提示：

1. 运行 `git pull` 以获取最新的实验软件。
2. 3C 测试比 3A 或 3B 的测试要求更高，失败可能是由你的 3A 或 3B 代码中的问题引起的。
3. 你的代码应该通过所有 3C 测试（如下所示），以及 3A 和 3B 测试。

```bash
$ go test -run 3C
...
PASS
ok      6.5840/raft1    126.054s
```

在提交之前多次运行测试并检查每次运行是否都打印 PASS 是个好主意。

```bash
for i in {0..10}; do go test; done
```

---

## 第 3D 部分：日志压缩（困难）

按照目前的情况，重启的服务器会重放完整的 Raft 日志以恢复其状态。然而，对于长期运行的服务来说，永远记住完整的 Raft 日志是不切实际的。相反，你将修改 Raft 以便与那些不时持久化存储其状态“快照（snapshot）”的服务进行协作，此时 Raft 会丢弃快照之前的日志条目。结果是持久化数据量更小，重启速度更快。然而，现在 Follower 可能会落后太多，以至于 Leader 已经丢弃了它赶上进度所需的日志条目；此时 Leader 必须发送一个快照加上从快照时间开始的日志。扩展版 Raft 论文的第 7 节概述了该方案；你必须设计细节。

你的 Raft 必须提供以下函数，服务可以使用其状态的序列化快照来调用该函数：

`Snapshot(index int, snapshot []byte)`

在 Lab 3D 中，测试器会周期性地调用 `Snapshot()`。在 Lab 4 中，你将编写一个调用 `Snapshot()` 的键/值服务器；快照将包含完整的键/值对表。服务层会在每个节点（不仅仅是 Leader）上调用 `Snapshot()`。

`index` 参数表示快照中反映的最高日志条目。Raft 应该丢弃该点之前的日志条目。你需要修改你的 Raft 代码，使其能在仅存储日志尾部的情况下运行。

你需要实现论文中讨论的 `InstallSnapshot` RPC，它允许 Raft Leader 告诉落后的 Raft 节点用快照替换其状态。你可能需要仔细思考 `InstallSnapshot` 应如何与图 2 中的状态和规则交互。

当 Follower 的 Raft 代码收到 `InstallSnapshot` RPC 时，它可以使用 `applyCh` 将快照以 `ApplyMsg` 的形式发送给服务。`ApplyMsg` 结构体定义已经包含了你需要的字段（也是测试器期望的）。请注意，这些快照只能推进服务的状态，不要导致其倒退。

如果服务器崩溃，它必须从持久化数据重启。你的 Raft 应该同时持久化 Raft 状态和相应的快照。使用 `persister.Save()` 的第二个参数来保存快照。如果没有快照，则传递 `nil` 作为第二个参数。

当服务器重启时，应用层读取持久化的快照并恢复其保存的状态。

实现 `Snapshot()` 和 `InstallSnapshot` RPC，以及对 Raft 的更改以支持这些功能（例如，使用截断的日志进行操作）。当你的解决方案通过 3D 测试（以及所有之前的 Lab 3 测试）时，即算完成。

1. `git pull` 以确保你有最新的软件。
2. 一个好的开始是修改你的代码，使其能够仅存储从某个索引 X 开始的日志部分。最初你可以将 X 设置为零并运行 3B/3C 测试。然后让 `Snapshot(index)` 丢弃 `index` 之前的日志，并将 X 设置为等于 `index`。如果一切顺利，你应该能通过第一个 3D 测试。
3. 未能通过第一个 3D 测试的一个常见原因是 Follower 花费太长时间才赶上 Leader。
4. 接下来：如果 Leader 没有使 Follower 更新所需的日志条目，让 Leader 发送 `InstallSnapshot` RPC。
5. 在单个 `InstallSnapshot` RPC 中发送整个快照。不要实现图 13 中用于分割快照的偏移量机制。
6. Raft 必须以允许 Go 垃圾回收器释放和重用内存的方式丢弃旧的日志条目；这要求没有可达的引用（指针）指向已丢弃的日志条目。
7. 不带 `-race` 运行全套 Lab 3 测试（3A+3B+3C+3D）的合理耗时是 6 分钟的实际时间和 1 分钟的 CPU 时间。带 `-race` 运行时，大约是 10 分钟的实际时间和 2 分钟的 CPU 时间。
8. 你的代码应该通过所有 3D 测试（如下所示），以及 3A、3B 和 3C 测试。

```bash
$ go test -run 3D
...
PASS
ok      6.5840/raft1    195.006s
```
