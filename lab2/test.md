# test

任务一

```bash
shaddocknh3@LAPTOP-JGAAJ56H:~/6.5840/src/kvsrv1$ go test -race
One client and reliable Put (reliable network)...
  ... Passed --  time  0.0s #peers 1 #RPCs     5 #Ops    0
Test: many clients racing to put values to the same key (reliable network)...
  ... Passed --  time  3.4s #peers 1 #RPCs  4751 #Ops 4751
Test: memory use many put clients (reliable network)...
--- FAIL: TestMemPutManyClientsReliable (225.09s)
    config.go:179: test took longer than 120 seconds
One client (unreliable network)...
  ... Passed --  time  8.9s #peers 1 #RPCs   260 #Ops  213
FAIL
exit status 1
FAIL    6.5840/kvsrv1   237.670s
shaddocknh3@LAPTOP-JGAAJ56H:~/6.5840/src/kvsrv1$ 
```

任务二

```bash
shaddocknh3@LAPTOP-JGAAJ56H:~/6.5840/src/kvsrv1/lock$ go test -v -run Reliable
=== RUN   TestOneClientReliable
Test: 1 lock clients (reliable network)...
  ... Passed --  time  2.0s #peers 1 #RPCs  1135 #Ops    0
--- PASS: TestOneClientReliable (2.01s)
=== RUN   TestManyClientsReliable
Test: 10 lock clients (reliable network)...
  ... Passed --  time  2.1s #peers 1 #RPCs 46845 #Ops    0
--- PASS: TestManyClientsReliable (2.13s)
PASS
ok      6.5840/kvsrv1/lock      4.146s
```

任务三/四

```bash
shaddocknh3@LAPTOP-JGAAJ56H:~/6.5840/src/kvsrv1$ go test -v
=== RUN   TestReliablePut
One client and reliable Put (reliable network)...
  ... Passed --  time  0.0s #peers 1 #RPCs     5 #Ops    0
--- PASS: TestReliablePut (0.00s)
=== RUN   TestPutConcurrentReliable
Test: many clients racing to put values to the same key (reliable network)...
  ... Passed --  time  2.2s #peers 1 #RPCs 16489 #Ops 16489
--- PASS: TestPutConcurrentReliable (2.15s)
=== RUN   TestMemPutManyClientsReliable
Test: memory use many put clients (reliable network)...
  ... Passed --  time 28.0s #peers 1 #RPCs 100000 #Ops    0
--- PASS: TestMemPutManyClientsReliable (52.99s)
=== RUN   TestUnreliableNet
One client (unreliable network)...
  ... Passed --  time  9.5s #peers 1 #RPCs   268 #Ops  212
--- PASS: TestUnreliableNet (9.50s)
PASS
ok      6.5840/kvsrv1   64.701s
shaddocknh3@LAPTOP-JGAAJ56H:~/6.5840/src/kvsrv1$ 
```
