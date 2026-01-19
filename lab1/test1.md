# test1.md

好久没有写过大狗屎代码了，今天算是重新体会到了。

第一次测试结果如下：

成功的测试

1. Starting wc test。单词计数测试，说明基础功能已经没问题了
2. Starting indexer test。开始倒排索引测试，说明系统能处理不同类型的 Map/Reduce 任务，不仅仅是单词计数。
3. Starting map parallelism test。Map 并行度测试，说明多个 Worker 能同时领走 Map 任务并行开工，没有冲突，也就是锁是没问题的，虽然我采用了非常史的大锁。
4. Starting job count test。作业计数测试，说明 Master 能正确分配任务数量，没有多发或者少发任务。

失败的测试

1. Starting reduce parallelism test。Reduce 并行度测试，并没有两只以上的 Reduce 在同时干活，有三个问题：
    1. Coordinator 在 Map 阶段还没彻底结束时，就把 Reduce 任务卡住了，导致 Reduce Worker 拿不到活，只能在那干等（Waiting）。
    2. 锁加得太重了，导致 Worker 们排队领任务，失去了并发的意义。
    3. Worker 拿到 Reduce 任务后迅速崩溃了，没来得及写文件。
2. Starting early exit test。提前退出测试，说明缺乏原子写入机制。Worker 写文件写到一半早退跑路了，在磁盘上留下了写了一半的残缺文件，导致后来接手的 Worker 读到了脏数据或者重复写入，最终结果不对。
3. Starting crash test。崩溃/容错测试，当 Worker 在做任务途中挂掉或者断连后，Coordinator 还在无限期等待，没有超时重试机制把任务抢回来重新分给活着的 Worker。

```bash
shaddocknh3@LAPTOP-JGAAJ56H:~/6.5840/src/main$ bash test-mr.sh
*** Starting wc test.
2026/01/18 16:01:49 No tasks available, worker is waiting.
2026/01/18 16:01:49 No tasks available, worker is waiting.
2026/01/18 16:01:49 All tasks are done. Worker exiting.
2026/01/18 16:01:50 All tasks are done. Worker exiting.
2026/01/18 16:01:50 All tasks are done. Worker exiting.
--- wc test: PASS
*** Starting indexer test.
2026/01/18 16:01:53 No tasks available, worker is waiting.
2026/01/18 16:01:53 All tasks are done. Worker exiting.
2026/01/18 16:01:55 All tasks are done. Worker exiting.
--- indexer test: PASS
*** Starting map parallelism test.
2026/01/18 16:02:01 No tasks available, worker is waiting.
2026/01/18 16:02:01 All tasks are done. Worker exiting.
2026/01/18 16:02:02 All tasks are done. Worker exiting.
--- map parallelism test: PASS
*** Starting reduce parallelism test.
2026/01/18 16:02:05 No tasks available, worker is waiting.
2026/01/18 16:02:05 All tasks are done. Worker exiting.
cat: 'mr-out*': No such file or directory
--- too few parallel reduces.
--- reduce parallelism test: FAIL
2026/01/18 16:02:16 All tasks are done. Worker exiting.
*** Starting job count test.
2026/01/18 16:02:37 No tasks available, worker is waiting.
2026/01/18 16:02:37 All tasks are done. Worker exiting.
2026/01/18 16:02:37 All tasks are done. Worker exiting.
2026/01/18 16:02:37 All tasks are done. Worker exiting.
--- job count test: PASS
2026/01/18 16:02:38 All tasks are done. Worker exiting.
*** Starting early exit test.
2026/01/18 16:02:39 No tasks available, worker is waiting.
2026/01/18 16:02:39 No tasks available, worker is waiting.
2026/01/18 16:02:39 All tasks are done. Worker exiting.
sort: cannot read: 'mr-out*': No such file or directory
2026/01/18 16:02:43 All tasks are done. Worker exiting.
2026/01/18 16:02:43 All tasks are done. Worker exiting.
cmp: EOF on mr-wc-all-initial which is empty
--- output changed after first worker exited
--- early exit test: FAIL
*** Starting crash test.
2026/01/18 16:02:46 No tasks available, worker is waiting.
2026/01/18 16:02:46 No tasks available, worker is waiting.
2026/01/18 16:02:46 No tasks available, worker is waiting.
2026/01/18 16:02:46 All tasks are done. Worker exiting.
2026/01/18 16:02:47 All tasks are done. Worker exiting.
2026/01/18 16:02:48 All tasks are done. Worker exiting.
2026/01/18 16:02:48 All tasks are done. Worker exiting.
2026/01/18 16:02:49 All tasks are done. Worker exiting.
2026/01/18 16:02:49 All tasks are done. Worker exiting.
2026/01/18 16:02:49 All tasks are done. Worker exiting.
2026/01/18 16:02:50 All tasks are done. Worker exiting.
2026/01/18 16:02:50 All tasks are done. Worker exiting.
2026/01/18 16:02:51 All tasks are done. Worker exiting.
2026/01/18 16:02:51 All tasks are done. Worker exiting.
2026/01/18 16:02:52 All tasks are done. Worker exiting.
2026/01/18 16:02:53 All tasks are done. Worker exiting.
2026/01/18 16:02:53 All tasks are done. Worker exiting.
2026/01/18 16:02:53 All tasks are done. Worker exiting.
2026/01/18 16:02:54 All tasks are done. Worker exiting.
2026/01/18 16:02:54 All tasks are done. Worker exiting.
2026/01/18 16:02:54 All tasks are done. Worker exiting.
2026/01/18 16:02:55 All tasks are done. Worker exiting.
2026/01/18 16:02:55 All tasks are done. Worker exiting.
2026/01/18 16:02:55 All tasks are done. Worker exiting.
2026/01/18 16:02:56 All tasks are done. Worker exiting.
2026/01/18 16:02:56 All tasks are done. Worker exiting.
2026/01/18 16:02:56 All tasks are done. Worker exiting.
2026/01/18 16:02:57 All tasks are done. Worker exiting.
2026/01/18 16:02:57 All tasks are done. Worker exiting.
2026/01/18 16:02:57 All tasks are done. Worker exiting.
2026/01/18 16:02:58 All tasks are done. Worker exiting.
2026/01/18 16:02:58 All tasks are done. Worker exiting.
2026/01/18 16:02:58 All tasks are done. Worker exiting.
2026/01/18 16:02:59 All tasks are done. Worker exiting.
2026/01/18 16:02:59 All tasks are done. Worker exiting.
2026/01/18 16:02:59 All tasks are done. Worker exiting.
2026/01/18 16:03:00 All tasks are done. Worker exiting.
2026/01/18 16:03:00 All tasks are done. Worker exiting.
2026/01/18 16:03:01 All tasks are done. Worker exiting.
2026/01/18 16:03:01 All tasks are done. Worker exiting.
2026/01/18 16:03:01 All tasks are done. Worker exiting.
2026/01/18 16:03:02 All tasks are done. Worker exiting.
2026/01/18 16:03:02 All tasks are done. Worker exiting.
2026/01/18 16:03:02 All tasks are done. Worker exiting.
2026/01/18 16:03:03 All tasks are done. Worker exiting.
2026/01/18 16:03:03 All tasks are done. Worker exiting.
2026/01/18 16:03:03 All tasks are done. Worker exiting.
2026/01/18 16:03:04 All tasks are done. Worker exiting.
2026/01/18 16:03:04 All tasks are done. Worker exiting.
2026/01/18 16:03:05 All tasks are done. Worker exiting.
2026/01/18 16:03:05 All tasks are done. Worker exiting.
2026/01/18 16:03:05 All tasks are done. Worker exiting.
2026/01/18 16:03:06 All tasks are done. Worker exiting.
2026/01/18 16:03:06 All tasks are done. Worker exiting.
2026/01/18 16:03:06 All tasks are done. Worker exiting.
2026/01/18 16:03:07 All tasks are done. Worker exiting.
2026/01/18 16:03:07 All tasks are done. Worker exiting.
2026/01/18 16:03:07 All tasks are done. Worker exiting.
2026/01/18 16:03:08 All tasks are done. Worker exiting.
2026/01/18 16:03:08 All tasks are done. Worker exiting.
2026/01/18 16:03:08 All tasks are done. Worker exiting.
2026/01/18 16:03:09 All tasks are done. Worker exiting.
2026/01/18 16:03:09 All tasks are done. Worker exiting.
2026/01/18 16:03:09 All tasks are done. Worker exiting.
2026/01/18 16:03:10 All tasks are done. Worker exiting.
2026/01/18 16:03:10 All tasks are done. Worker exiting.
2026/01/18 16:03:10 All tasks are done. Worker exiting.
2026/01/18 16:03:15 All tasks are done. Worker exiting.
2026/01/18 16:03:15 All tasks are done. Worker exiting.
2026/01/18 16:03:15 All tasks are done. Worker exiting.
2026/01/18 16:03:16 All tasks are done. Worker exiting.
2026/01/18 16:03:16 All tasks are done. Worker exiting.
2026/01/18 16:03:16 All tasks are done. Worker exiting.
2026/01/18 16:03:17 All tasks are done. Worker exiting.
2026/01/18 16:03:17 All tasks are done. Worker exiting.
2026/01/18 16:03:17 All tasks are done. Worker exiting.
2026/01/18 16:03:18 All tasks are done. Worker exiting.
2026/01/18 16:03:18 All tasks are done. Worker exiting.
2026/01/18 16:03:19 All tasks are done. Worker exiting.
2026/01/18 16:03:19 All tasks are done. Worker exiting.
2026/01/18 16:03:19 All tasks are done. Worker exiting.
2026/01/18 16:03:20 All tasks are done. Worker exiting.
2026/01/18 16:03:20 All tasks are done. Worker exiting.
2026/01/18 16:03:21 All tasks are done. Worker exiting.
2026/01/18 16:03:21 All tasks are done. Worker exiting.
2026/01/18 16:03:21 All tasks are done. Worker exiting.
2026/01/18 16:03:22 All tasks are done. Worker exiting.
2026/01/18 16:03:22 All tasks are done. Worker exiting.
2026/01/18 16:03:22 All tasks are done. Worker exiting.
2026/01/18 16:03:23 All tasks are done. Worker exiting.
2026/01/18 16:03:23 All tasks are done. Worker exiting.
2026/01/18 16:03:23 All tasks are done. Worker exiting.
2026/01/18 16:03:24 All tasks are done. Worker exiting.
2026/01/18 16:03:24 All tasks are done. Worker exiting.
2026/01/18 16:03:24 All tasks are done. Worker exiting.
2026/01/18 16:03:25 All tasks are done. Worker exiting.
2026/01/18 16:03:25 All tasks are done. Worker exiting.
2026/01/18 16:03:25 All tasks are done. Worker exiting.
2026/01/18 16:03:26 All tasks are done. Worker exiting.
2026/01/18 16:03:26 All tasks are done. Worker exiting.
2026/01/18 16:03:26 All tasks are done. Worker exiting.
2026/01/18 16:03:27 All tasks are done. Worker exiting.
2026/01/18 16:03:27 All tasks are done. Worker exiting.
2026/01/18 16:03:27 All tasks are done. Worker exiting.
2026/01/18 16:03:28 All tasks are done. Worker exiting.
2026/01/18 16:03:28 All tasks are done. Worker exiting.
2026/01/18 16:03:28 All tasks are done. Worker exiting.
2026/01/18 16:03:29 All tasks are done. Worker exiting.
2026/01/18 16:03:29 All tasks are done. Worker exiting.
2026/01/18 16:03:29 All tasks are done. Worker exiting.
2026/01/18 16:03:30 All tasks are done. Worker exiting.
2026/01/18 16:03:30 All tasks are done. Worker exiting.
2026/01/18 16:03:30 All tasks are done. Worker exiting.
2026/01/18 16:03:31 All tasks are done. Worker exiting.
2026/01/18 16:03:31 All tasks are done. Worker exiting.
2026/01/18 16:03:31 All tasks are done. Worker exiting.
2026/01/18 16:03:32 All tasks are done. Worker exiting.
2026/01/18 16:03:32 All tasks are done. Worker exiting.
2026/01/18 16:03:33 All tasks are done. Worker exiting.
2026/01/18 16:03:33 All tasks are done. Worker exiting.
2026/01/18 16:03:33 All tasks are done. Worker exiting.
2026/01/18 16:03:34 All tasks are done. Worker exiting.
2026/01/18 16:03:34 All tasks are done. Worker exiting.
2026/01/18 16:03:34 All tasks are done. Worker exiting.
2026/01/18 16:03:35 All tasks are done. Worker exiting.
2026/01/18 16:03:35 All tasks are done. Worker exiting.
2026/01/18 16:03:35 All tasks are done. Worker exiting.
2026/01/18 16:03:36 All tasks are done. Worker exiting.
2026/01/18 16:03:36 All tasks are done. Worker exiting.
2026/01/18 16:03:36 All tasks are done. Worker exiting.
2026/01/18 16:03:37 All tasks are done. Worker exiting.
2026/01/18 16:03:37 All tasks are done. Worker exiting.
2026/01/18 16:03:37 All tasks are done. Worker exiting.
2026/01/18 16:03:38 All tasks are done. Worker exiting.
2026/01/18 16:03:38 All tasks are done. Worker exiting.
2026/01/18 16:03:38 All tasks are done. Worker exiting.
2026/01/18 16:03:39 All tasks are done. Worker exiting.
2026/01/18 16:03:39 All tasks are done. Worker exiting.
2026/01/18 16:03:39 All tasks are done. Worker exiting.
2026/01/18 16:03:40 All tasks are done. Worker exiting.
2026/01/18 16:03:40 All tasks are done. Worker exiting.
2026/01/18 16:03:40 All tasks are done. Worker exiting.
2026/01/18 16:03:41 All tasks are done. Worker exiting.
2026/01/18 16:03:41 All tasks are done. Worker exiting.
2026/01/18 16:03:41 All tasks are done. Worker exiting.
2026/01/18 16:03:42 All tasks are done. Worker exiting.
2026/01/18 16:03:42 All tasks are done. Worker exiting.
2026/01/18 16:03:42 All tasks are done. Worker exiting.
2026/01/18 16:03:43 All tasks are done. Worker exiting.
2026/01/18 16:03:43 All tasks are done. Worker exiting.
2026/01/18 16:03:43 All tasks are done. Worker exiting.
2026/01/18 16:03:44 All tasks are done. Worker exiting.
2026/01/18 16:03:44 All tasks are done. Worker exiting.
2026/01/18 16:03:44 All tasks are done. Worker exiting.
2026/01/18 16:03:45 All tasks are done. Worker exiting.
2026/01/18 16:03:45 All tasks are done. Worker exiting.
2026/01/18 16:03:45 All tasks are done. Worker exiting.
2026/01/18 16:03:46 All tasks are done. Worker exiting.
2026/01/18 16:03:46 All tasks are done. Worker exiting.
2026/01/18 16:03:46 All tasks are done. Worker exiting.
2026/01/18 16:03:50 All tasks are done. Worker exiting.
2026/01/18 16:03:50 All tasks are done. Worker exiting.
2026/01/18 16:03:50 All tasks are done. Worker exiting.
2026/01/18 16:03:51 All tasks are done. Worker exiting.
2026/01/18 16:03:51 All tasks are done. Worker exiting.
2026/01/18 16:03:51 All tasks are done. Worker exiting.
2026/01/18 16:03:52 All tasks are done. Worker exiting.
2026/01/18 16:03:52 All tasks are done. Worker exiting.
2026/01/18 16:03:52 All tasks are done. Worker exiting.
2026/01/18 16:03:53 All tasks are done. Worker exiting.
2026/01/18 16:03:53 All tasks are done. Worker exiting.
2026/01/18 16:03:53 All tasks are done. Worker exiting.
2026/01/18 16:03:54 All tasks are done. Worker exiting.
2026/01/18 16:03:54 All tasks are done. Worker exiting.
2026/01/18 16:03:54 All tasks are done. Worker exiting.
2026/01/18 16:03:55 All tasks are done. Worker exiting.
2026/01/18 16:03:55 All tasks are done. Worker exiting.
2026/01/18 16:03:55 All tasks are done. Worker exiting.
2026/01/18 16:03:56 All tasks are done. Worker exiting.
2026/01/18 16:03:56 All tasks are done. Worker exiting.
2026/01/18 16:03:56 All tasks are done. Worker exiting.
2026/01/18 16:03:57 All tasks are done. Worker exiting.
2026/01/18 16:03:57 All tasks are done. Worker exiting.
2026/01/18 16:03:57 All tasks are done. Worker exiting.
2026/01/18 16:03:58 All tasks are done. Worker exiting.
2026/01/18 16:03:58 All tasks are done. Worker exiting.
2026/01/18 16:03:58 All tasks are done. Worker exiting.
2026/01/18 16:03:59 All tasks are done. Worker exiting.
2026/01/18 16:03:59 All tasks are done. Worker exiting.
2026/01/18 16:03:59 All tasks are done. Worker exiting.
2026/01/18 16:04:00 All tasks are done. Worker exiting.
2026/01/18 16:04:00 All tasks are done. Worker exiting.
2026/01/18 16:04:01 All tasks are done. Worker exiting.
2026/01/18 16:04:01 All tasks are done. Worker exiting.
2026/01/18 16:04:01 All tasks are done. Worker exiting.
2026/01/18 16:04:02 All tasks are done. Worker exiting.
2026/01/18 16:04:02 All tasks are done. Worker exiting.
2026/01/18 16:04:02 All tasks are done. Worker exiting.
2026/01/18 16:04:03 All tasks are done. Worker exiting.
2026/01/18 16:04:03 All tasks are done. Worker exiting.
2026/01/18 16:04:04 All tasks are done. Worker exiting.
2026/01/18 16:04:04 All tasks are done. Worker exiting.
2026/01/18 16:04:04 All tasks are done. Worker exiting.
2026/01/18 16:04:05 All tasks are done. Worker exiting.
2026/01/18 16:04:05 All tasks are done. Worker exiting.
2026/01/18 16:04:05 All tasks are done. Worker exiting.
2026/01/18 16:04:06 All tasks are done. Worker exiting.
2026/01/18 16:04:06 All tasks are done. Worker exiting.
2026/01/18 16:04:06 All tasks are done. Worker exiting.
2026/01/18 16:04:07 All tasks are done. Worker exiting.
2026/01/18 16:04:07 All tasks are done. Worker exiting.
2026/01/18 16:04:07 All tasks are done. Worker exiting.
2026/01/18 16:04:08 All tasks are done. Worker exiting.
2026/01/18 16:04:08 All tasks are done. Worker exiting.
2026/01/18 16:04:08 All tasks are done. Worker exiting.
2026/01/18 16:04:09 All tasks are done. Worker exiting.
2026/01/18 16:04:09 All tasks are done. Worker exiting.
2026/01/18 16:04:09 All tasks are done. Worker exiting.
2026/01/18 16:04:10 All tasks are done. Worker exiting.
2026/01/18 16:04:10 All tasks are done. Worker exiting.
2026/01/18 16:04:10 All tasks are done. Worker exiting.
2026/01/18 16:04:11 All tasks are done. Worker exiting.
2026/01/18 16:04:11 All tasks are done. Worker exiting.
2026/01/18 16:04:11 All tasks are done. Worker exiting.
2026/01/18 16:04:12 All tasks are done. Worker exiting.
2026/01/18 16:04:12 All tasks are done. Worker exiting.
2026/01/18 16:04:12 All tasks are done. Worker exiting.
2026/01/18 16:04:13 All tasks are done. Worker exiting.
2026/01/18 16:04:13 All tasks are done. Worker exiting.
2026/01/18 16:04:13 All tasks are done. Worker exiting.
2026/01/18 16:04:14 All tasks are done. Worker exiting.
2026/01/18 16:04:14 All tasks are done. Worker exiting.
2026/01/18 16:04:14 All tasks are done. Worker exiting.
2026/01/18 16:04:15 All tasks are done. Worker exiting.
2026/01/18 16:04:15 All tasks are done. Worker exiting.
2026/01/18 16:04:15 All tasks are done. Worker exiting.
2026/01/18 16:04:16 All tasks are done. Worker exiting.
2026/01/18 16:04:16 All tasks are done. Worker exiting.
2026/01/18 16:04:17 All tasks are done. Worker exiting.
2026/01/18 16:04:17 All tasks are done. Worker exiting.
2026/01/18 16:04:17 All tasks are done. Worker exiting.
2026/01/18 16:04:18 All tasks are done. Worker exiting.
2026/01/18 16:04:18 All tasks are done. Worker exiting.
2026/01/18 16:04:18 All tasks are done. Worker exiting.
2026/01/18 16:04:19 All tasks are done. Worker exiting.
2026/01/18 16:04:19 All tasks are done. Worker exiting.
2026/01/18 16:04:19 All tasks are done. Worker exiting.
2026/01/18 16:04:20 All tasks are done. Worker exiting.
2026/01/18 16:04:20 All tasks are done. Worker exiting.
2026/01/18 16:04:20 All tasks are done. Worker exiting.
2026/01/18 16:04:21 All tasks are done. Worker exiting.
2026/01/18 16:04:21 All tasks are done. Worker exiting.
2026/01/18 16:04:21 All tasks are done. Worker exiting.
2026/01/18 16:04:22 All tasks are done. Worker exiting.
2026/01/18 16:04:25 All tasks are done. Worker exiting.
2026/01/18 16:04:25 All tasks are done. Worker exiting.
2026/01/18 16:04:26 All tasks are done. Worker exiting.
2026/01/18 16:04:26 All tasks are done. Worker exiting.
2026/01/18 16:04:26 All tasks are done. Worker exiting.
2026/01/18 16:04:27 All tasks are done. Worker exiting.
2026/01/18 16:04:27 All tasks are done. Worker exiting.
2026/01/18 16:04:27 All tasks are done. Worker exiting.
2026/01/18 16:04:28 All tasks are done. Worker exiting.
2026/01/18 16:04:28 All tasks are done. Worker exiting.
2026/01/18 16:04:28 All tasks are done. Worker exiting.
2026/01/18 16:04:29 All tasks are done. Worker exiting.
2026/01/18 16:04:29 All tasks are done. Worker exiting.
2026/01/18 16:04:29 All tasks are done. Worker exiting.
2026/01/18 16:04:30 All tasks are done. Worker exiting.
2026/01/18 16:04:30 All tasks are done. Worker exiting.
2026/01/18 16:04:30 All tasks are done. Worker exiting.
2026/01/18 16:04:31 All tasks are done. Worker exiting.
2026/01/18 16:04:31 All tasks are done. Worker exiting.
2026/01/18 16:04:31 All tasks are done. Worker exiting.
2026/01/18 16:04:32 All tasks are done. Worker exiting.
2026/01/18 16:04:32 All tasks are done. Worker exiting.
2026/01/18 16:04:32 All tasks are done. Worker exiting.
2026/01/18 16:04:33 All tasks are done. Worker exiting.
2026/01/18 16:04:33 All tasks are done. Worker exiting.
2026/01/18 16:04:33 All tasks are done. Worker exiting.
2026/01/18 16:04:34 All tasks are done. Worker exiting.
2026/01/18 16:04:34 All tasks are done. Worker exiting.
2026/01/18 16:04:34 All tasks are done. Worker exiting.
2026/01/18 16:04:35 All tasks are done. Worker exiting.
2026/01/18 16:04:35 All tasks are done. Worker exiting.
2026/01/18 16:04:35 All tasks are done. Worker exiting.
2026/01/18 16:04:36 All tasks are done. Worker exiting.
2026/01/18 16:04:36 All tasks are done. Worker exiting.
2026/01/18 16:04:36 All tasks are done. Worker exiting.
2026/01/18 16:04:37 All tasks are done. Worker exiting.
2026/01/18 16:04:37 All tasks are done. Worker exiting.
2026/01/18 16:04:37 All tasks are done. Worker exiting.
2026/01/18 16:04:38 All tasks are done. Worker exiting.
2026/01/18 16:04:38 All tasks are done. Worker exiting.
2026/01/18 16:04:38 All tasks are done. Worker exiting.
2026/01/18 16:04:39 All tasks are done. Worker exiting.
2026/01/18 16:04:39 All tasks are done. Worker exiting.
2026/01/18 16:04:39 All tasks are done. Worker exiting.
2026/01/18 16:04:40 All tasks are done. Worker exiting.
2026/01/18 16:04:40 All tasks are done. Worker exiting.
2026/01/18 16:04:40 All tasks are done. Worker exiting.
2026/01/18 16:04:41 All tasks are done. Worker exiting.
2026/01/18 16:04:41 All tasks are done. Worker exiting.
2026/01/18 16:04:41 All tasks are done. Worker exiting.
2026/01/18 16:04:42 All tasks are done. Worker exiting.
2026/01/18 16:04:42 All tasks are done. Worker exiting.
2026/01/18 16:04:42 All tasks are done. Worker exiting.
2026/01/18 16:04:43 All tasks are done. Worker exiting.
2026/01/18 16:04:43 All tasks are done. Worker exiting.
2026/01/18 16:04:43 All tasks are done. Worker exiting.
2026/01/18 16:04:44 All tasks are done. Worker exiting.
2026/01/18 16:04:44 All tasks are done. Worker exiting.
2026/01/18 16:04:44 All tasks are done. Worker exiting.
2026/01/18 16:04:45 All tasks are done. Worker exiting.
2026/01/18 16:04:45 All tasks are done. Worker exiting.
2026/01/18 16:04:45 All tasks are done. Worker exiting.
2026/01/18 16:04:46 All tasks are done. Worker exiting.
2026/01/18 16:04:46 All tasks are done. Worker exiting.
2026/01/18 16:04:46 All tasks are done. Worker exiting.
2026/01/18 16:04:47 All tasks are done. Worker exiting.
2026/01/18 16:04:47 All tasks are done. Worker exiting.
2026/01/18 16:04:47 All tasks are done. Worker exiting.
2026/01/18 16:04:48 All tasks are done. Worker exiting.
2026/01/18 16:04:48 All tasks are done. Worker exiting.
2026/01/18 16:04:49 All tasks are done. Worker exiting.
2026/01/18 16:04:49 All tasks are done. Worker exiting.
2026/01/18 16:04:49 All tasks are done. Worker exiting.
2026/01/18 16:04:50 All tasks are done. Worker exiting.
2026/01/18 16:04:50 All tasks are done. Worker exiting.
2026/01/18 16:04:50 All tasks are done. Worker exiting.
2026/01/18 16:04:51 All tasks are done. Worker exiting.
2026/01/18 16:04:51 All tasks are done. Worker exiting.
2026/01/18 16:04:51 All tasks are done. Worker exiting.
2026/01/18 16:04:52 All tasks are done. Worker exiting.
2026/01/18 16:04:52 All tasks are done. Worker exiting.
2026/01/18 16:04:53 All tasks are done. Worker exiting.
2026/01/18 16:04:53 All tasks are done. Worker exiting.
2026/01/18 16:04:53 All tasks are done. Worker exiting.
2026/01/18 16:04:54 All tasks are done. Worker exiting.
2026/01/18 16:04:54 All tasks are done. Worker exiting.
2026/01/18 16:04:54 All tasks are done. Worker exiting.
2026/01/18 16:04:55 All tasks are done. Worker exiting.
2026/01/18 16:04:55 All tasks are done. Worker exiting.
2026/01/18 16:04:55 All tasks are done. Worker exiting.
cmp: EOF on mr-crash-all after byte 268, line 3
--- crash output is not the same as mr-correct-crash.txt
--- crash test: FAIL
*** FAILED SOME TESTS
```
