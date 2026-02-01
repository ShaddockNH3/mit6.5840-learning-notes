# kvraft

## 引言 (Introduction)

在这个实验中，你将使用你在实验 3 中构建的 Raft 库来构建一个容错的键/值存储服务。对客户端而言，该服务看起来与实验 2 的服务器类似。然而，它不再是单个服务器，而是由一组使用 Raft 来协助它们维护相同数据库的服务器组成。只要大多数（majority）服务器存活且能互相通信，即使发生其他故障或网络分区，你的键/值服务也应继续处理客户端请求。在完成实验 4 后，你将实现 Raft 交互图中所示的所有部分（Clerk、Service 和 Raft）。

客户端将像在实验 2 中一样，通过一个 `Clerk` 与你的键/值服务进行交互。`Clerk` 实现了 `Put` 和 `Get` 方法，其语义与实验 2 相同：`Put` 操作是“至多一次”（at-most-once）的，且 `Put`/`Get` 必须形成一个线性化（linearizable）的历史。

对于单个服务器来说，提供线性化相对容易。但对于复制服务来说则更难，因为所有服务器必须为并发请求选择相同的执行顺序，必须避免使用未更新的状态回复客户端，并且在故障后恢复状态时必须保留所有已确认的客户端更新。

本实验分为三个部分。

* **Part A**：你将使用你的 Raft 实现来实现一个复制状态机包 `rsm`；`rsm` 对它所复制的请求内容是不可知的（agnostic）。
* **Part B**：你将使用 `rsm` 来实现一个复制的键/值服务，但不使用快照。
* **Part C**：你将使用你在实验 3D 中实现的快照功能，这将允许 Raft 丢弃旧的日志条目。
  请在各自的截止日期前提交每个部分。

你应该复习 **扩展版 Raft 论文**，特别是 **第 7 节**（但不需要看第 8 节）。为了获得更广阔的视角，可以看看 Chubby, Paxos Made Live, Spanner, Zookeeper, Harp, Viewstamped Replication 以及 Bolosky 等人的论文。

**尽早开始。**

## 入门指南 (Getting Started)

我们在 `src/kvraft1` 中为你提供了骨架代码和测试。骨架代码使用骨架包 `src/kvraft1/rsm` 来复制服务器。服务器必须实现 `rsm` 中定义的 `StateMachine` 接口，以便利用 `rsm` 进行自我复制。你的大部分工作将是实现 `rsm` 以提供与服务器无关的复制功能。你还需要修改 `kvraft1/client.go` 和 `kvraft1/server.go` 以实现服务器特定的部分。这种拆分允许你在下一个实验中重用 `rsm`。你可能能够重用一些实验 2 的代码（例如，通过复制或导入 "src/kvsrv1" 包来重用服务器代码），但这并不是必须的。

要开始运行，请执行以下命令。别忘了执行 `git pull` 以获取最新软件。

```bash
$ cd ~/6.5840
$ git pull
..
```

## Part A: 复制状态机 (RSM) (中等/困难)

```bash
$ cd src/kvraft1/rsm
$ go test -v
=== RUN   TestBasic
Test RSM basic (reliable network)...
..
    config.go:147: one: took too long
```

在使用 Raft 进行复制的客户端/服务器服务的常见情况中，服务以两种方式与 Raft 交互：服务 Leader 通过调用 `raft.Start()` 提交客户端操作，所有服务副本通过 Raft 的 `applyCh` 接收已提交的操作并执行它们。在 Leader 上，这就产生了交互：在任何给定时间，一些服务器 goroutine 正在处理客户端请求，它们调用了 `raft.Start()`，并且每一个都在等待其操作被提交并获知执行结果。随着已提交的操作出现在 `applyCh` 上，每一个操作都需要由服务执行，结果需要交给调用 `raft.Start()` 的那个 goroutine，以便它可以将结果返回给客户端。

`rsm` 包封装了上述交互。它位于服务（例如键/值数据库）和 Raft 之间。在 `rsm/rsm.go` 中，你需要实现一个“读取器 (reader)” goroutine 来读取 `applyCh`，以及一个 `rsm.Submit()` 函数，该函数为客户端操作调用 `raft.Start()`，然后等待读取器 goroutine 将该操作的执行结果交给它。

使用 `rsm` 的服务对 `rsm` 读取器 goroutine 表现为一个提供了 `DoOp()` 方法的 `StateMachine` 对象。读取器 goroutine 应该将每个已提交的操作交给 `DoOp()`；`DoOp()` 的返回值应提供给相应的 `rsm.Submit()` 调用以供其返回。`DoOp()` 的参数和返回值类型为 `any`；实际值应分别与服务传递给 `rsm.Submit()` 的参数和返回值类型相同。

服务应将每个客户端操作传递给 `rsm.Submit()`。为了帮助读取器 goroutine 将 `applyCh` 消息与等待的 `rsm.Submit()` 调用匹配起来，`Submit()` 应将每个客户端操作封装在一个 `Op` 结构中，并附带一个唯一标识符。然后 `Submit()` 应等待直到操作已提交并被执行，最后返回执行结果（即 `DoOp()` 返回的值）。如果 `raft.Start()` 指示当前节点不是 Raft Leader，`Submit()` 应返回一个 `rpc.ErrWrongLeader` 错误。`Submit()` 应该检测并处理在调用 `raft.Start()` 后领导权立即变更导致操作丢失（从未提交）的情况。

对于 Part A，`rsm` 测试器充当服务的角色，提交它解释为对由单个整数组成的状态进行递增的操作。在 Part B 中，你将使用 `rsm` 作为实现 `StateMachine`（和 `DoOp()`）的键/值服务的一部分，并调用 `rsm.Submit()`。

如果一切顺利，客户端请求的事件序列如下：

1. 客户端向服务 Leader 发送请求。
2. 服务 Leader 带着请求调用 `rsm.Submit()`。
3. `rsm.Submit()` 带着请求调用 `raft.Start()`，然后等待。
4. Raft 提交请求并将其发送到所有节点的 `applyChs` 上。
5. 每个节点上的 `rsm` 读取器 goroutine 从 `applyCh` 读取请求并将其传递给服务的 `DoOp()`。
6. 在 Leader 上，`rsm` 读取器 goroutine 将 `DoOp()` 的返回值交给最初提交请求的 `Submit()` goroutine，然后 `Submit()` 返回该值。

你的服务器不应直接通信；它们应仅通过 Raft 相互交互。

**任务：** 实现 `rsm.go`：`Submit()` 方法和一个读取器 goroutine。如果你通过了 rsm 4A 测试，则表示你已完成此任务：

```bash
$ cd src/kvraft1/rsm
$ go test -v -run 4A
=== RUN   TestBasic4A
Test RSM basic (reliable network)...
  ... Passed --   1.2  3    48    0
--- PASS: TestBasic4A (1.21s)
=== RUN   TestLeaderFailure4A
  ... Passed --  9223372036.9  3    31    0
--- PASS: TestLeaderFailure4A (1.50s)
PASS
ok      6.5840/kvraft1/rsm      2.887s
```

* 你不需要向 Raft `ApplyMsg` 或 Raft RPC（如 `AppendEntries`）添加任何字段，但你可以这样做。
* 你的解决方案需要处理一种情况：`rsm` Leader 为一个用 `Submit()` 提交的请求调用了 `Start()`，但在请求提交到日志之前就失去了领导权。

  * 一种处理方式是让 `rsm` 检测到它已经失去了领导权（通过注意到 Raft 的任期 term 已经改变，或者在 `Start()` 返回的索引处出现了不同的请求），并从 `Submit()` 返回 `rpc.ErrWrongLeader`。
  * 如果前任 Leader 被网络分区隔离了，它将不知道新的 Leader 是谁；但同一分区中的任何客户端也无法与新 Leader 通信，所以在这种情况下服务器无限期等待直到分区修复是没问题的。
* 测试器在关闭一个节点时会调用你的 Raft 的 `rf.Kill()`。Raft 应该关闭 `applyCh`，以便你的 `rsm` 获知关闭操作，并能退出所有循环。

## Part B: 无快照的键/值服务 (中等)

```bash
$ cd src/kvraft1
$ go test -v -run TestBasic4B
=== RUN   TestBasic4B
Test: one client (4B basic) (reliable network)...
    kvtest.go:62: Wrong error 
$
```

现在你将使用 `rsm` 包来复制一个键/值服务器。每个服务器（"kvservers"）都将有一个关联的 `rsm`/Raft 节点。`Clerk` 向关联的 Raft 为 Leader 的 `kvserver` 发送 `Put()` 和 `Get()` RPC。`kvserver` 代码将 `Put`/`Get` 操作提交给 `rsm`，`rsm` 使用 Raft 复制这些操作，并在每个节点上调用你的服务器的 `DoOp`，后者应该将操作应用到该节点的键/值数据库中；目的是让所有服务器维护完全相同的键/值数据库副本。

一个 `Clerk` 有时不知道哪个 `kvserver` 是 Raft Leader。如果 `Clerk` 向错误的 `kvserver` 发送 RPC，或者无法连接到 `kvserver`，`Clerk` 应该通过向不同的 `kvserver` 发送请求来重试。如果键/值服务将操作提交到了其 Raft 日志（并因此将操作应用到了键/值状态机），Leader 通过响应 RPC 向 `Clerk` 报告结果。如果操作未能提交（例如，如果 Leader 被替换了），服务器报告错误，并且 `Clerk` 使用不同的服务器重试。

你的 `kvserver` 不应直接通信；它们应仅通过 Raft 相互交互。

**你的第一个任务**是实现一个在没有消息丢失和没有服务器故障时能工作的解决方案。

* 你可以将实验 2 的客户端代码（`kvsrv1/client.go`）复制到 `kvraft1/client.go` 中。你需要添加逻辑来决定向哪个 `kvserver` 发送每个 RPC。
* 你还需要在 `server.go` 中实现 `Put()` 和 `Get()` RPC 处理程序。这些处理程序应该使用 `rsm.Submit()` 将请求提交给 Raft。
* 当 `rsm` 包从 `applyCh` 读取命令时，它应该调用 `DoOp` 方法，你需要像在 `server.go` 中实现它。
* 当你能通过测试套件中的第一个测试（`go test -v -run TestBasic4B`）时，你就完成了这个任务。

**注意：**

* 如果 `kvserver` 不是大多数（majority）的一部分，它不应完成 `Get()` RPC（以避免提供陈旧数据）。一个简单的解决方案是使用 `Submit()` 将每个 `Get()`（以及每个 `Put()`）都写入 Raft 日志。你不需要实现第 8 节中描述的只读操作优化。
* 最好从一开始就添加锁，因为避免死锁的需求有时会影响整体代码设计。使用 `go test -race` 检查你的代码是否没有数据竞争（race-free）。

**现在你应该修改你的解决方案**，以应对网络和服务器故障。

* 你会面临的一个问题是，`Clerk` 可能需要多次发送 RPC，直到找到一个回复积极的 `kvserver`。
* 如果一个 Leader 刚刚向 Raft 日志提交一个条目后就失败了，`Clerk` 可能不会收到回复，因此可能会将请求重新发送给另一个 Leader。对 `Clerk.Put()` 的每次调用应该只对应一个特定版本号的单一执行（线性一致性）。

**添加代码以处理故障：**

* 你的 `Clerk` 可以使用与实验 2 类似的重试计划，包括如果重试的 `Put` RPC 的响应丢失则返回 `ErrMaybe`。
* 当你能可靠地通过所有 4B 测试时（`go test -v -run 4B`），你就完成了任务。
* 回想一下，`rsm` Leader 可能会失去领导权并从 `Submit()` 返回 `rpc.ErrWrongLeader`。在这种情况下，你应该安排 `Clerk` 将请求重新发送给其他服务器，直到找到新的 Leader。
* 你可能需要修改你的 `Clerk` 以记住哪个服务器是上一个 RPC 的 Leader，并优先将下一个 RPC 发送给该服务器。这将避免在每个 RPC 上浪费时间寻找 Leader，这有助于你足够快地通过某些测试。

你的代码现在应该通过 Lab 4B 测试，如下所示：

```bash
$ cd kvraft1
$ go test -run 4B
Test: one client (4B basic) ...
  ... Passed --   3.2  5  1041  183
Test: one client (4B speed) ...
  ... Passed --  15.9  3  3169    0
Test: many clients (4B many clients) ...
  ... Passed --   3.9  5  3247  871
Test: unreliable net, many clients (4B unreliable net, many clients) ...
  ... Passed --   5.3  5  1035  167
Test: unreliable net, one client (4B progress in majority) ...
  ... Passed --   2.9  5   155    3
Test: no progress in minority (4B) ...
  ... Passed --   1.6  5   102    3
Test: completion after heal (4B) ...
  ... Passed --   1.3  5    67    4
Test: partitions, one client (4B partitions, one client) ...
  ... Passed --   6.2  5   958  155
Test: partitions, many clients (4B partitions, many clients (4B)) ...
  ... Passed --   6.8  5  3096  855
Test: restarts, one client (4B restarts, one client 4B ) ...
  ... Passed --   6.7  5   311   13
Test: restarts, many clients (4B restarts, many clients) ...
  ... Passed --   7.5  5  1223   95
Test: unreliable net, restarts, many clients (4B unreliable net, restarts, many clients ) ...
  ... Passed --   8.4  5   804   33
Test: restarts, partitions, many clients (4B restarts, partitions, many clients) ...
  ... Passed --  10.1  5  1308  105
Test: unreliable net, restarts, partitions, many clients (4B unreliable net, restarts, partitions, many clients) ...
  ... Passed --  11.9  5  1040   33
Test: unreliable net, restarts, partitions, random keys, many clients (4B unreliable net, restarts, partitions, random keys, many clients) ...
  ... Passed --  12.1  7  2801   93
PASS
ok      6.5840/kvraft1  103.797s
```

*每个 Passed 后面的数字分别是：实际时间（秒）、节点数量、发送的 RPC 数量（包括客户端 RPC）、执行的键/值操作数量（Clerk Get/Put 调用）。*

## Part C: 带快照的键/值服务 (中等)

目前的状况是，你的键/值服务器没有调用 Raft 库的 `Snapshot()` 方法，因此重启的服务器必须重放完整的持久化 Raft 日志才能恢复其状态。现在，你将修改 `kvserver` 和 `rsm`，配合 Raft 使用实验 3D 中的 `Snapshot()` 来节省日志空间并减少重启时间。

测试器将 `maxraftstate` 传递给你的 `StartKVServer()`，后者将其传递给 `rsm`。`maxraftstate` 指示你的持久化 Raft 状态允许的最大字节数（包括日志，但不包括快照）。你应该将 `maxraftstate` 与 `rf.PersistBytes()` 进行比较。每当你的 `rsm` 检测到 Raft 状态大小接近此阈值时，它应该通过调用 Raft 的 `Snapshot` 来保存快照。`rsm` 可以通过调用 `StateMachine` 接口的 `Snapshot` 方法来获取 `kvserver` 的快照从而创建此快照。如果 `maxraftstate` 为 -1，你不需要进行快照。`maxraftstate` 限制适用于你的 Raft 作为第一个参数传递给 `persister.Save()` 的 GOB 编码字节。

你可以在 `tester1/persister.go` 中找到 `persister` 对象的源代码。

**任务：**

* 修改你的 `rsm`，使其能检测持久化的 Raft 状态何时增长过大，然后将快照交给 Raft。
* 当一个 `rsm` 服务器重启时，它应该使用 `persister.ReadSnapshot()` 读取快照，并且如果快照的长度大于零，则将快照传递给 `StateMachine` 的 `Restore()` 方法。
* 如果你通过了 `rsm` 中的 `TestSnapshot4C`，你就完成了此任务。

```bash
$ cd kvraft1/rsm
$ go test -run TestSnapshot4C
=== RUN   TestSnapshot4C
  ... Passed --  9223372036.9  3   230    0
--- PASS: TestSnapshot4C (3.88s)
PASS
ok      6.5840/kvraft1/rsm      3.882s
```

* 思考一下 `rsm` 应该在什么时候对状态进行快照，以及除了服务器状态之外，快照中还应该包含什么。
* Raft 使用 `Save()` 将每个快照连同相应的 Raft 状态一起存储在 `persister` 对象中。你可以使用 `ReadSnapshot()` 读取最新存储的快照。
* **将存储在快照中的结构体所有字段首字母大写**（以便 GOB 编码）。
* 实现 `kvraft1/server.go` 中的 `Snapshot()` 和 `Restore()` 方法，供 `rsm` 调用。
* 修改 `rsm` 以处理包含快照的 `applyCh` 消息。

这个任务可能会暴露你的 Raft 和 `rsm` 库中的 Bug。如果你对 Raft 实现进行了更改，请确保它继续通过所有 Lab 3 的测试。
Lab 4 测试的合理用时为 400 秒实际时间和 700 秒 CPU 时间。

你的代码应该通过 4C 测试（如下例所示）以及 4A+B 测试（并且你的 Raft 必须继续通过 Lab 3 测试）。

```bash
$ go test -run 4C
Test: snapshots, one client (4C SnapshotsRPC) ...
Test: InstallSnapshot RPC (4C) ...
  ... Passed --   4.5  3   241   64
Test: snapshots, one client (4C snapshot size is reasonable) ...
  ... Passed --  11.4  3  2526  800
Test: snapshots, one client (4C speed) ...
  ... Passed --  14.2  3  3149    0
Test: restarts, snapshots, one client (4C restarts, snapshots, one client) ...
  ... Passed --   6.8  5   305   13
Test: restarts, snapshots, many clients (4C restarts, snapshots, many clients ) ...
  ... Passed --   9.0  5  5583  795
Test: unreliable net, snapshots, many clients (4C unreliable net, snapshots, many clients) ...
  ... Passed --   4.7  5   977  155
Test: unreliable net, restarts, snapshots, many clients (4C unreliable net, restarts, snapshots, many clients) ...
  ... Passed --   8.6  5   847   33
Test: unreliable net, restarts, partitions, snapshots, many clients (4C unreliable net, restarts, partitions, snapshots, many clients) ...
  ... Passed --  11.5  5   841   33
Test: unreliable net, restarts, partitions, snapshots, random keys, many clients (4C unreliable net, restarts, partitions, snapshots, random keys, many clients) ...
  ... Passed --  12.8  7  2903   93
PASS
ok      6.5840/kvraft1  83.543s
```
