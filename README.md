# mit6.5840 学习笔记

## 前言

这是西二 go 考核的 task7 应该，由于我本人要决定去考研，剩余的时间可能不足以写完 task 5-6，所以在 go 组组长的建议下，直接上 task7。

预计花费 20 天的时间写完。

应“不要把你做的 lab 源代码上传到互联网（例如 github），保持一定的学术诚信，这是为你，也是为了你的学弟学妹”的要求，本仓库不包含任何源码，只包含笔记以及个人感悟。

此外，稍微注意一下环境和版本问题。本人在 wsl 环境下运行，做的是 25 spring 版本，使用 go 1.23.5 版本。

## 目录

### lec: 课程笔记

- 课程笔记 lec1：[课程笔记 lec1](./lec/lec1-introduction.md)，课前需要阅读 [MapReduce 论文](./paper/MapReduce.md)
- 课程笔记 lec2：[课程笔记 lec2](./lec/lec2-RPC-and-Threads.md)
- 课程笔记 lec3：[课程笔记 lec3](./lec/lec3-GFS.md)，课前需要阅读 [FaultTolerant 论文](./paper/FaultTolerant.md)

### lab1: Map Reduce

看完论文就可以开始做 lab1 了。

- lab1 论文原文翻译：[MapReduce 论文](./paper/MapReduce.md)
- lab1 论文原文阅读笔记：[MapReduce 阅读笔记](./paper-read/paper1.md)
- lab1 实验要求翻译：[lab1 实验要求翻译](./lab1/mr.md)
- lab1 实验要求分析：[lab1 实验要求分析](./lab1/mr-exp.md)
- lab1 框架代码阅读：
  - 串行版本实现：[mr-seq](./lab1/code/mrsequential.go)
  - key-value 接口定义：[mr-kv](./lab1/code/wc.go)
- lab1 过程笔记：[mr-aar](./lab1/mr-aar.md)
- lab1 测试记录
  - mr 测试记录 1：[mr 测试记录 1](./lab1/test1.md)
  - mr 测试记录 2：[mr 测试记录 2](./lab1/test2.md)

### lab2: Key/Value Server

其实可以直接做。

2020 版本的课和 2025 的 lab 对应的有出入，虽然和《一种实用的容错虚拟机系统设计》这篇论文有出入，不过其实无所谓。

- lab2 实验要求翻译：[lab2 实验要求翻译](./lab2/kvsrv1.md)
- lab2 实验要求分析：[lab2 实验要求分析](./lab2/kvsrv1-exp.md)
- lab2 过程笔记：[kvsrv1-aar](./lab2/kvsrv1-aar.md)
- lab2 测试记录：[lab2 测试记录](./lab2/test.md)

### lab3：Raft

看完论文就可以开始做 lab3 了，不过我多看了一个 GFS 论文（看了一点看不下去了，先去写 lab3 去）。

- lab1 论文原文翻译：[Raft(Extend) 论文](./paper/Raft(Extend).md)
- lab1 论文原文阅读笔记：[Raft(Extend) 阅读笔记](./paper-read/paper3.md)
- lab3 实验要求翻译：[lab3 实验要求翻译](./lab3/raft.md)
- lab3 实验要求分析：[lab3 实验要求分析](./lab3/raft-exp.md)

### 论文阅读

- [MapReduce 论文](./paper/MapReduce.md)
- [MapReduce 阅读笔记](./paper-read/paper1.md)
- [FaultTolerant 论文](./paper/FaultTolerant.md)
- [FaultTolerant 阅读笔记](./paper-read/paper2.md)
- [Raft(Extend) 论文](./paper/Raft(Extend).md)
- [Raft(Extend) 阅读笔记](./paper-read/paper3.md)
- [GFS 论文](./paper/GFS.md)
- [GFS 阅读笔记](./paper-read/paper4.md)

### 代码练习

- [一些语法回顾](./practice/)

## 参考资料

1.[课程原网站](http://nil.csail.mit.edu/6.5840/2025/schedule.html)
2.[b 站 2020 MIT 6.824](https://www.bilibili.com/video/BV1R7411t71W?vd_source=8a950947d6bc6120547b345c6856e11b&spm_id_from=333.788.videopod.episodes)搭配 b 站 自带的翻译。事实上也可以直接使用油管的视频加上油管自带的翻译，不过 b  站 搬运的本本更经典一点，所以就用 b 站 的版本上。
