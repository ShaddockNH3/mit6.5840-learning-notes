# krsrv1

<https://pdos.csail.mit.edu/6.824/labs/lab-kvsrv1.html>

## 🟢 简介 (Introduction)

在这个实验中，你将构建一个运行在单机上的键/值（Key/Value）服务器。即便在网络故障的情况下，该服务器也要确保每个 `Put`（写入）操作**至多执行一次 (at-most-once)**，并且所有操作必须是**线性一致的 (linearizable)**。你将利用这个 KV 服务器来实现一个“锁 (Lock)”。在后续的实验中，你将通过复制（Replicate）像这样的服务器来处理服务器崩溃的问题。

## 🗝️ KV 服务器 (KV server)

每个客户端都通过一个 `Clerk`（办事员/代理）与键/值服务器进行交互，`Clerk` 负责向服务器发送 RPC（远程过程调用）。客户端可以向服务器发送两种不同的 RPC：`Put(key, value, version)` 和 `Get(key)`。

服务器在内存中维护一个 Map，记录每个 Key 对应的 **(value, version)** 元组。Key 和 Value 都是字符串。版本号（Version）记录了该 Key 被写入的次数。

对于 `Put(key, value, version)` 操作：
只有当 `Put` 请求中的 `version` 恰好匹配服务器端该 Key 的当前 `version` 时，服务器才会安装或替换该 Key 的值。

* 如果版本号匹配，服务器还会将该 Key 的版本号增加 1。
* 如果版本号不匹配，服务器应该返回 `rpc.ErrVersion`。

客户端可以通过调用 `version` 为 0 的 `Put` 来创建一个新的 Key（此时服务器存储的结果版本将变为 1）。如果 `Put` 的版本号大于 0 但该 Key 在服务器上不存在，服务器应该返回 `rpc.ErrNoKey`。

对于 `Get(key)` 操作：
获取该 Key 的当前值及其关联的版本号。如果该 Key 在服务器上不存在，服务器应该返回 `rpc.ErrNoKey`。

为每个 Key 维护一个版本号，对于利用 `Put` 实现“锁”，以及在网络不可靠、客户端需要重传时确保 `Put` 的**至多一次 (at-most-once)** 语义非常有用。

当你完成了这个实验并通过了所有测试后，从调用 `Clerk.Get` 和 `Clerk.Put` 的客户端角度来看，你将拥有一个线性一致的键/值服务。
这意味着：

* 如果客户端的操作不是并发执行的，那么每个客户端的 `Clerk.Get` 和 `Clerk.Put` 都将观察到由前序操作序列所隐含的状态修改。
* 对于并发操作，其返回值和最终状态必须与“这些操作以某种顺序逐个执行”的结果相同。
* 所谓“并发操作”是指它们在时间上重叠：例如，客户端 X 调用 `Clerk.Put()`，客户端 Y 也调用 `Clerk.Put()`，然后客户端 X 的调用返回了。
* 一个操作必须能观察到在它开始之前就已经完成的所有操作的效果。（更多背景信息请参阅关于线性一致性的 FAQ）。

线性一致性对应用程序来说很方便，因为它的行为和你看到的一个“一次只处理一个请求”的单机服务器行为是一样的。例如，如果一个客户端收到了服务器对更新请求的成功响应，那么随后其他客户端发起的读取操作，都保证能看到那次更新的效果。对于单机服务器来说，提供线性一致性是相对容易的。

## 🚀 开始上手 (Getting Started)

我们在 `src/kvsrv1` 中为你提供了骨架代码和测试。

* `kvsrv1/client.go` 实现了一个 `Clerk`，客户端利用它来管理与服务器的 RPC 交互；`Clerk` 提供了 `Put` 和 `Get` 方法。
* `kvsrv1/server.go` 包含了服务器代码，其中包括实现 RPC 请求服务端的 `Put` 和 `Get` 处理函数。

你需要修改 `client.go` 和 `server.go`。
RPC 的请求、回复以及错误值定义在 `kvsrv1/rpc` 包中的 `kvsrv1/rpc/rpc.go` 文件里。你应该看一看这个文件，尽管你不需要修改它。

要开始运行，请执行以下命令。别忘了 `git pull` 以获取最新的软件。

```bash
$ cd ~/6.5840
$ git pull
...
$ cd src/kvsrv1
$ go test -v
=== RUN   TestReliablePut
One client and reliable Put (reliable network)...
    kvsrv_test.go:25: Put err ErrNoKey
...
```

## ✅ 任务一：可靠网络下的 Key/Value 服务器 (简单)

你的第一个任务是实现一个在**没有消息丢失**的情况下能工作的解决方案。
你需要向 `client.go` 中的 `Clerk` `Put`/`Get` 方法添加发送 RPC 的代码，并在 `server.go` 中实现 `Put` 和 `Get` 的 RPC 处理函数。

当你的代码通过了测试套件中的 `Reliable`（可靠）测试时，你就完成了这个任务：

```bash
$ go test -v -run Reliable
=== RUN   TestReliablePut
One client and reliable Put (reliable network)...
  ... Passed --   0.0  1     5    0
--- PASS: TestReliablePut (0.00s)
=== RUN   TestPutConcurrentReliable
Test: many clients racing to put values to the same key (reliable network)...
info: linearizability check timed out, assuming history is ok
  ... Passed --   3.1  1 90171 90171
--- PASS: TestPutConcurrentReliable (3.07s)
=== RUN   TestMemPutManyClientsReliable
Test: memory use many put clients (reliable network)...
  ... Passed --   9.2  1 100000    0
--- PASS: TestMemPutManyClientsReliable (16.59s)
PASS
ok   6.5840/kvsrv1 19.681s
```

（Passed 后面的数字分别是：实际耗时秒数、常数 1、发送的 RPC 数量（包括客户端 RPC）、执行的键/值操作数量（Clerk Get 和 Put 调用）。）

请使用 `go test -race` 检查你的代码是否存在数据竞争（race-free）。

## 🔒 任务二：使用 Key/Value Clerk 实现锁 (中等)

在许多分布式应用中，运行在不同机器上的客户端使用键/值服务器来协调它们的活动。例如，ZooKeeper 和 Etcd 允许客户端使用分布式锁进行协调，这类似于 Go 程序中的线程如何使用锁（即 `sync.Mutex`）进行协调。ZooKeeper 和 Etcd 是通过条件写入（conditional put）来实现这种锁的。

在这个练习中，你的任务是在客户端 `Clerk.Put` 和 `Clerk.Get` 调用的基础上实现一个锁。
这个锁支持两个方法：`Acquire`（获取）和 `Release`（释放）。
锁的规范是：同一时间只能有一个客户端成功获取锁；其他客户端必须等待，直到第一个客户端使用 `Release` 释放了锁。

我们在 `src/kvsrv1/lock/` 中为你提供了骨架代码和测试。你需要修改 `src/kvsrv1/lock/lock.go`。你的 `Acquire` 和 `Release` 代码可以通过调用 `lk.ck.Put()` 和 `lk.ck.Get()` 与你的键/值服务器通信。

如果一个客户端在持有锁的时候崩溃了，锁将永远不会被释放。在一个比本实验更复杂的设计中，客户端会给锁附加一个租约（Lease）。当租约过期时，锁服务器会代表客户端释放锁。在本实验中，客户端不会崩溃，你可以忽略这个问题。

实现 `Acquire` 和 `Release`。当你的代码通过了 `lock` 子目录下的 `Reliable` 测试时，你就完成了这个练习：

```bash
$ cd lock
$ go test -v -run Reliable
...
```

如果这一部分需要写的代码很少，但比起上一个练习，它需要更多**独立的思考**。

* 你需要为每个锁客户端提供一个唯一标识符；调用 `kvtest.RandValue(8)` 可以生成一个随机字符串。
* 锁服务应该使用一个特定的 Key 来存储“锁状态”（你需要决定锁状态具体是什么）。这个 Key 是通过 `src/kvsrv1/lock/lock.go` 中 `MakeLock` 的参数 `l` 传递进来的。

## 🌪️ 任务三：应对消息丢失的 Key/Value 服务器 (中等)

这个练习的主要挑战在于网络可能会**乱序、延迟或丢弃** RPC 请求和/或回复。为了从丢弃的请求/回复中恢复，`Clerk` 必须**不断重试**每个 RPC，直到它收到来自服务器的回复。

如果网络丢弃了一个 RPC **请求**消息，那么客户端重发请求就能解决问题：服务器将接收并执行刚刚重发的请求。

然而，网络可能丢弃的是 RPC **回复**消息。客户端不知道哪个消息被丢弃了；客户端只观察到它没有收到回复。

* 如果被丢弃的是回复，并且客户端重发了 RPC 请求，那么服务器将收到两份请求副本。
* 对于 `Get` 来说这没问题，因为 `Get` 不修改服务器状态。
* 对于 `Put`，重发带有相同版本号的 RPC 也是安全的，因为服务器是根据版本号**有条件地**执行 `Put` 的；如果服务器已经接收并执行了一个 `Put` RPC，它将对重传的副本回复 `rpc.ErrVersion`，而不是再次执行该 `Put`。

**一个棘手的情况是**：如果服务器对一个 `Clerk` **重试**的 RPC 回复了 `rpc.ErrVersion`。
在这种情况下，`Clerk` 无法知道它的 `Put` 到底是被服务器执行了还是没执行：

1. 可能是第一个 RPC 被服务器执行了，但成功的回复被网络丢弃了，所以服务器只对重传的 RPC 发送了 `rpc.ErrVersion`。
2. 也可能是另一个 `Clerk` 在你的第一个 RPC 到达之前就更新了 Key，导致服务器即没执行你的第一个 RPC，也没执行第二个，并对两者都回复了 `rpc.ErrVersion`。

因此，如果 `Clerk` 收到针对**重传**的 `Put` RPC 的 `rpc.ErrVersion`，`Clerk.Put` 必须向应用程序返回 `rpc.ErrMaybe` 而不是 `rpc.ErrVersion`，因为请求**可能**已经被执行了。然后由应用程序来处理这种情况。
如果服务器对**初始**（非重传）的 `Put` RPC 回复 `rpc.ErrVersion`，那么 `Clerk` 应该向应用程序返回 `rpc.ErrVersion`，因为该 RPC 肯定没有被服务器执行。

如果 `Put` 是“恰好一次 (exactly-once)”（即没有 `rpc.ErrMaybe` 错误），对应用程序开发者来说会更方便，但这在不在服务器端为每个 `Clerk` 维护状态的情况下很难保证。在本实验的最后一个练习中，你将使用你的 `Clerk` 实现一个锁，以此探索如何使用“至多一次 (at-most-once)”的 `Clerk.Put` 进行编程。

现在你应该修改 `kvsrv1/client.go`，以便在面对 RPC 请求和回复丢失时能继续工作。

* 客户端的 `ck.clnt.Call()` 返回 `true` 表示客户端收到了来自服务器的 RPC 回复。
* 返回 `false` 表示它没有收到回复（更准确地说，`Call()` 会等待回复一段时间，如果在该时间内没有回复到达则返回 `false`）。
* 你的 `Clerk` 应该不断重发 RPC 直到收到回复。
* 请记住上面关于 `rpc.ErrMaybe` 的讨论。
* 你的解决方案不需要修改服务器端代码。

在 `Clerk` 中添加代码，如果收不到回复则重试。如果你的代码通过了 `kvsrv1/` 下的所有测试，你就完成了此任务：

```bash
$ go test -v
=== RUN   TestReliablePut
...
=== RUN   TestUnreliableNet
One client (unreliable network)...
  ... Passed --   7.6  1   251  208
--- PASS: TestUnreliableNet (7.60s)
```

*在客户端重试之前，应该稍微等待一下；你可以使用 go 的 `time` 包并调用 `time.Sleep(100 * time.Millisecond)`。*

## 🔓 任务四：在不可靠网络下使用 Key/Value Clerk 实现锁 (简单)

修改你的锁实现，使其在网络不可靠的情况下能与你修改后的 Key/Value 客户端正确配合工作。当你的代码通过了 `kvsrv1/lock/` 下的所有测试（包括不可靠网络的测试）时，你就完成了这个练习：

```bash
$ cd lock
$ go test -v
...
=== RUN   TestOneClientUnreliable
...
=== RUN   TestManyClientsUnreliable
...
```
