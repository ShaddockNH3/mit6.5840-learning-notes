# mr

<https://pdos.csail.mit.edu/6.824/labs/lab-mr.html>

## Introduction (简介)

在这个实验中，你将构建一个 MapReduce 系统。你需要实现一个 **Worker（工作进程）**，它调用应用程序的 Map 和 Reduce 函数，并处理文件的读写；你还需要实现一个 **Coordinator（协调者进程）**，负责向 Worker 分派任务并处理 Worker 的故障。你将构建一个类似于 MapReduce 论文中所描述的系统。（注意：本实验使用“Coordinator”一词代替论文中的“Master”）。

## Getting started (开始上手)

你需要配置 Go 语言环境来完成实验。

使用 git（一种版本控制系统）获取初始实验软件。如需了解更多关于 git 的信息，请查阅 Pro Git 书籍或 git 用户手册。

```bash
$ git clone git://g.csail.mit.edu/6.5840-golabs-2025 6.5840
$ cd 6.5840
$ ls
Makefile src
$
```

我们在 `src/main/mrsequential.go` 中为你提供了一个简单的串行 MapReduce 实现。它在一个进程中一次运行一个 Map 和 Reduce 任务。我们还为你提供了几个 MapReduce 应用程序：`mrapps/wc.go` 中的单词计数（word-count），以及 `mrapps/indexer.go` 中的文本索引器。你可以按如下方式串行运行单词计数：

```bash
$ cd ~/6.5840
$ cd src/main
$ go build -buildmode=plugin ../mrapps/wc.go
$ rm mr-out*
$ go run mrsequential.go wc.so pg*.txt
$ more mr-out-0
A 509
ABOUT 2
ACT 8
...
```

`mrsequential.go` 将其输出保留在文件 `mr-out-0` 中。输入来自名为 `pg-xxx.txt` 的文本文件。

你可以随意借用 `mrsequential.go` 中的代码。你也应该看看 `mrapps/wc.go`，了解 MapReduce 应用程序代码是什么样的。

对于本实验和所有其他实验，我们可能会发布代码更新。为了确保你能获取这些更新并使用 `git pull` 轻松合并它们，最好将我们提供的代码保留在原始文件中。你可以按照实验说明的指示在我们提供的代码中添加内容；只是不要移动它。将你自己的新函数放在新文件中是可以的。

## Your Job (moderate/hard) (你的任务 - 中等/困难)

你的工作是实现一个分布式 MapReduce，由两个程序组成：**Coordinator（协调者）** 和 **Worker（工作者）**。将只有一个 Coordinator 进程，和一个或多个并行执行的 Worker 进程。在实际系统中，Worker 会运行在一堆不同的机器上，但对于本实验，你将在同一台机器上运行所有进程。Worker 将通过 RPC（远程过程调用）与 Coordinator 通信。每个 Worker 进程将在一个循环中向 Coordinator 索取任务，从一个或多个文件中读取任务的输入，执行任务，将任务的输出写入一个或多个文件，然后再次向 Coordinator 索取新任务。Coordinator 应该能注意到如果一个 Worker 在合理的时间内（本实验使用 **10秒**）没有完成任务，并将同一任务交给不同的 Worker。

我们已经为你提供了一些起始代码。Coordinator 和 Worker 的“main”例程分别在 `main/mrcoordinator.go` 和 `main/mrworker.go` 中；**不要更改这些文件**。你应该将你的实现放在 `mr/coordinator.go`、`mr/worker.go` 和 `mr/rpc.go` 中。

以下是如何在单词计数 MapReduce 应用程序上运行你的代码。首先，确保单词计数插件是新构建的：

```bash
go build -buildmode=plugin ../mrapps/wc.go
```

在 `main` 目录下，运行 Coordinator：

```bash
rm mr-out*
go run mrcoordinator.go pg-*.txt
```

传递给 `mrcoordinator.go` 的 `pg-*.txt` 参数是输入文件；每个文件对应一个“Split（切片）”，并且是一个 Map 任务的输入。

在一个或多个其他窗口中，运行一些 Worker：

```bash
go run mrworker.go wc.so
```

当 Worker 和 Coordinator 完成后，查看 `mr-out-*` 中的输出。当你完成实验后，输出文件的排序联合（sorted union）应该与串行输出匹配，如下所示：

```bash
$ cat mr-out-* | sort | more
A 509
ABOUT 2
ACT 8
...
```

我们在 `main/test-mr.sh` 中为你提供了一个测试脚本。测试会检查当给定 `pg-xxx.txt` 文件作为输入时，`wc` 和 `indexer` MapReduce 应用程序是否产生正确的输出。测试还会检查你的实现是否并行运行 Map 和 Reduce 任务，以及你的实现是否能从运行任务时崩溃的 Worker 中恢复。

如果你现在运行测试脚本，它会挂起，因为 Coordinator 永远不会完成：

```bash
$ cd ~/6.5840/src/main
$ bash test-mr.sh
*** Starting wc test.
```

你可以将 `mr/coordinator.go` 中的 `Done` 函数里的 `ret := false` 改为 `true`，以便 Coordinator 立即退出。然后：

```bash
$ bash test-mr.sh
*** Starting wc test.
sort: No such file or directory
cmp: EOF on mr-wc-all
--- wc output is not the same as mr-correct-wc.txt
--- wc test: FAIL
$
```

测试脚本期望在名为 `mr-out-X` 的文件中看到输出，每个 Reduce 任务对应一个文件。`mr/coordinator.go` 和 `mr/worker.go` 的空实现不会产生这些文件（或者做任何其他事情），所以测试失败。

当你完成后，测试脚本的输出应该如下所示：

```bash
$ bash test-mr.sh
*** Starting wc test.
--- wc test: PASS
*** Starting indexer test.
--- indexer test: PASS
*** Starting map parallelism test.
--- map parallelism test: PASS
*** Starting reduce parallelism test.
--- reduce parallelism test: PASS
*** Starting job count test.
--- job count test: PASS
*** Starting early exit test.
--- early exit test: PASS
*** Starting crash test.
--- crash test: PASS
*** PASSED ALL TESTS
$
```

你可能会看到一些来自 Go RPC 包的错误，看起来像这样：
`2019/12/16 13:27:09 rpc.Register: method "Done" has 1 input parameters; needs exactly three`
忽略这些消息；注册 Coordinator 作为 RPC 服务器会检查其所有方法是否适合 RPC（即有3个输入）；我们知道 `Done` 不是通过 RPC 调用的。

此外，根据你终止 Worker 进程的策略，你可能会看到一些如下形式的错误：
`2025/02/11 16:21:32 dialing:dial unix /var/tmp/5840-mr-501: connect: connection refused`
每个测试看到少数几条这样的消息是可以的；它们出现在 Worker 在 Coordinator 退出后无法联系 Coordinator RPC 服务器时。

### A few rules (一些规则)

* Map 阶段应该将中间键（intermediate keys）划分为 ( nReduce ) 个桶，用于 ( nReduce ) 个 reduce 任务，其中 ( nReduce ) 是 reduce 任务的数量——这是 `main/mrcoordinator.go` 传递给 `MakeCoordinator()` 的参数。每个 Mapper 应该创建 ( nReduce ) 个中间文件供 reduce 任务通过。
* Worker 的实现应该将第 ( X ) 个 reduce 任务的输出放在文件 `mr-out-X` 中。
* `mr-out-X` 文件应该包含每行一个 Reduce 函数的输出。该行应该使用 Go 的 `"%v %v"` 格式生成，传入键和值。查看 `main/mrsequential.go` 中注释为 "this is the correct format"（这是正确格式）的行。如果你的实现偏离这种格式太多，测试脚本将失败。
* 你可以修改 `mr/worker.go`、`mr/coordinator.go` 和 `mr/rpc.go`。你可以为了测试临时修改其他文件，但要确保你的代码能与原始版本一起工作；我们将使用原始版本进行测试。
* Worker 应该将中间 Map 输出放在**当前目录**的文件中，以便你的 Worker 稍后可以读取它们作为 Reduce 任务的输入。
* `main/mrcoordinator.go` 期望 `mr/coordinator.go` 实现一个 `Done()` 方法，当 MapReduce 作业完全完成时返回 `true`；此时，`mrcoordinator.go` 将退出。
* 当作业完全完成后，Worker 进程应该退出。实现这一点的一个简单方法是使用 `call()` 的返回值：如果 Worker 未能联系到 Coordinator，它可以假设 Coordinator 已经退出，因为作业已完成，所以 Worker 也可以终止。根据你的设计，你可能也会发现有一个“请退出（please exit）”的伪任务让 Coordinator 给到 Worker 很有帮助。

## Hints (提示)

* **指导（Guidance）页面**有一些关于开发和调试的技巧。
* 开始的一种方法是修改 `mr/worker.go` 的 `Worker()`，发送一个 RPC 到 Coordinator 请求一个任务。然后修改 Coordinator 以回复一个尚未开始的 Map 任务的文件名。然后修改 Worker 读取该文件并调用应用程序的 Map 函数，就像在 `mrsequential.go` 中一样。
* 应用程序的 Map 和 Reduce 函数是在运行时使用 Go 的 plugin 包加载的，文件名以 `.so` 结尾。
* 如果你更改了 `mr/` 目录中的任何内容，你可能需要重新构建你使用的任何 MapReduce 插件，例如 `go build -buildmode=plugin ../mrapps/wc.go`。
* 本实验依赖于 Worker 共享文件系统。当所有 Worker 在同一台机器上运行时，这很简单，但如果 Worker 在不同的机器上运行，则需要像 GFS 这样的全局文件系统。
* 中间文件的一个合理的命名约定是 `mr-X-Y`，其中 ( X ) 是 Map 任务编号，( Y ) 是 reduce 任务编号。
* Worker 的 map 任务代码需要一种方法将中间键/值对存储在文件中，以便在 reduce 任务期间正确读回。一种可能性是使用 Go 的 `encoding/json` 包。要将键/值对以 JSON 格式写入打开的文件：

  ```go
  enc := json.NewEncoder(file)
  for _, kv := ... {
    err := enc.Encode(&kv)
  ```

  以及读回这样的文件：

  ```go
  dec := json.NewDecoder(file)
  for {
    var kv KeyValue
    if err := dec.Decode(&kv); err != nil {
      break
    }
    kva = append(kva, kv)
  }
  ```

* 你的 Worker 的 map 部分可以使用 `ihash(key)` 函数（在 `worker.go` 中）为给定的键选择 reduce 任务。
* 你可以从 `mrsequential.go` 中窃取一些代码，用于读取 Map 输入文件，在 Map 和 Reduce 之间对中间键/值对进行排序，以及将 Reduce 输出存储在文件中。
* Coordinator 作为一个 RPC 服务器，将是并发的；**不要忘记锁定共享数据**。
* 使用 Go 的竞态检测器（race detector），使用 `go run -race`。`test-mr.sh` 在开头有一个注释，告诉你如何使用 `-race` 运行它。当我们给你的实验评分时，我们**不会**使用竞态检测器。尽管如此，如果你的代码有竞态条件，即使没有竞态检测器，它也很可能在我们要测试时失败。
* Worker 有时需要等待，例如，在最后一个 map 完成之前，reduce 无法开始。一种可能性是 Worker 定期向 Coordinator 请求工作，在每次请求之间使用 `time.Sleep()` 睡眠。另一种可能性是 Coordinator 中的相关 RPC 处理程序有一个循环等待，使用 `time.Sleep()` 或 `sync.Cond`。Go 为每个 RPC 在其自己的线程中运行处理程序，因此一个处理程序正在等待的事实不需要阻止 Coordinator 处理其他 RPC。
* Coordinator 无法可靠地分辨崩溃的 Worker、活着但因某种原因停滞的 Worker、以及正在执行但速度太慢而无用的 Worker。你能做的最好的事情是让 Coordinator 等待一段时间，然后放弃并将任务重新发布给不同的 Worker。对于本实验，让 Coordinator 等待 **10秒**；之后 Coordinator 应假设 Worker 已死亡（当然，它可能没有）。
* 如果你选择实现**备份任务（Backup Tasks，第3.6节）**，请注意我们要测试的是：当 Worker 执行任务且没有崩溃时，你的代码不会调度多余的任务。备份任务应该只在一段相对较长的时间（例如 10秒）之后才被调度。
* 要测试崩溃恢复，你可以使用 `mrapps/crash.go` 应用程序插件。它会在 Map 和 Reduce 函数中随机退出。
* 为了确保在存在崩溃的情况下没有人观察到部分写入的文件，MapReduce 论文提到了使用临时文件并在完全写入后原子性地重命名它的技巧。你可以使用 `ioutil.TempFile`（或者如果你运行的是 Go 1.17 或更高版本，使用 `os.CreateTemp`）来创建一个临时文件，并使用 `os.Rename` 来原子性地重命名它。
* `test-mr.sh` 在子目录 `mr-tmp` 中运行其所有进程，所以如果出了问题你想查看中间或输出文件，请去那里看。你可以随意临时修改 `test-mr.sh` 以在测试失败后退出，这样脚本就不会继续测试（并覆盖输出文件）。
* `test-mr-many.sh` 连续多次运行 `test-mr.sh`，你可能想这样做是为了发现低概率的 bug。它接受一个参数作为运行测试的次数。你不应该并行运行多个 `test-mr.sh` 实例，因为 Coordinator 会重用同一个套接字，导致冲突。
* Go RPC 仅发送名称以大写字母开头的结构体字段。子结构体也必须具有大写的字段名称。
* 当调用 RPC `call()` 函数时，reply 结构体应该包含所有默认值。RPC 调用应该看起来像这样：

  ```go
  reply := SomeType{}
  call(..., &reply)
  ```

  而在调用之前不要设置 `reply` 的任何字段。如果你传递具有非默认字段的 `reply` 结构，RPC 系统可能会默默地返回不正确的值。

## No-credit challenge exercises (无学分挑战练习)

* 实现你自己的 MapReduce 应用程序（参见 `mrapps/*` 中的示例），例如，分布式 Grep（MapReduce 论文第 2.3 节）。
* 让你的 MapReduce Coordinator 和 Worker 在单独的机器上运行，就像在实践中一样。你需要将你的 RPC 设置为通过 TCP/IP 而不是 Unix 套接字进行通信（参见 `Coordinator.server()` 中注释掉的行），并使用共享文件系统读取/写入文件。例如，你可以 ssh 进入 MIT 的多台 Athena 集群机器，它们使用 AFS 共享文件；或者你可以租用几个 AWS 实例并使用 S3 进行存储。
