# raft

## 要做什么

主要是先阅读 raft 的骨架，然后实现创建节点 Make，发起共识 Start，查询状态 GetState 等。

必须使用 labrpc 包进行 rpc 通信，不可以使用共享变量或文件来在不同的 Raft 节点通信。

## 3A

根据论文在 Raft 结构体中添加论文里图 2 的所有状态。

完善 RPC 填充 RequestVoteArgs 和 RequestVoteReply 结构体。

实现 RequestVote RPC 处理函数，让大家可以相互投票

定义 AppendEntries RPC 结构体，本阶段是为了心跳，后面要用来传日志，并且编写它的处理方法

修改 Make 函数，启动一个后台 goroutine 来定期发送心跳（发不带日志的 AppendEntries）或发起选举

时间控制，心跳频率每秒不超过 10 次，选举超时得比论文建议的 150-300ms 更大一些，但是不能太大，必须保证在 5 秒内选出 Leader。

禁止使用 Go 的 time.Timer 或 time.Ticker，直接在循环里使用 time.Sleep 来控制时间间隔。

## 3B

运行 `git pull` 以获取最新的实验软件。

调整日志索引策略，Raft 日志虽然是 1 起始，但建议视为 0 起始，并在索引 0 处放置一个 Term 为 0 的占位条目，以便 `PrevLogIndex` 可以为 0。

实现 `Start()` 函数，允许新命令加入日志。

根据论文图 2，完善 `AppendEntries` RPC 的发送与接收逻辑，以支持日志条目的追加和一致性检查。

实现日志提交逻辑，当条目被提交时，通过 `applyCh` 发送给应用层（注意通过 `TestBasicAgree3B()`）。

实现**选举限制**（论文 5.4.1 节），确保只有拥有最新日志的 Candidate 才能赢得选举。

优化循环等待机制，禁止忙等待（busy loop），必须使用条件变量或在循环中插入 `time.Sleep(10 * time.Millisecond)`。

性能控制，确保测试运行时间不超过 1 分钟（实际时间）和 5 秒（CPU 时间），避免过多的睡眠或 RPC 泛滥。

## 3C

先实现持久化，然后实现多步回退。

## 3D
