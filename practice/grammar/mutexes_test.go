package grammar_test

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
)

// 原子计数器
// cpu 级别的原子操作，不需要加锁，性能更高
// 但是只能操作简单类型的数据
func TestAtomicCounter(t *testing.T) {
	var ops atomic.Uint64
	var wg sync.WaitGroup
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range 1000 {
				ops.Add(1)
			}
		}()

	}
	wg.Wait()
	fmt.Println("ops:", ops.Load())

	var num int64 = 0
	for range 50 {
		wg.Add(1)
		go func() {
			defer wg.Done() 
			for range 1000 {
				num += 1 // 这里没有加锁，最后的结果是不确定的
			}
		}()
	}
	wg.Wait()
	fmt.Println("num:", num)
}

/* 
互斥锁计数器 
这个东西的话，其实和原子计数器的功能是一样的，
但是性能会差一些，因为加锁和解锁是需要时间的。
不过互斥锁的好处是可以保护更复杂的数据结构，
比如说 map、slice 之类的。
type 里有一个互斥锁 mu，每次进行操作的时候（例如写操作）
都会先加锁，操作完毕之后再解锁
以保证在同一时间
只有一个 goroutine 能够访问和修改 counters 字段
这样可以保证数据的一致性，避免竞态条件的发生。
*/

type Container struct{
	mu sync.Mutex
	counters map[string]int
}

func (c *Container) Inc(name string){
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counters[name] += 1
}

func TestMutexCounter(t *testing.T){
	c := Container{
		counters: make(map[string]int),
	}
	var wg sync.WaitGroup
	doIncrements := func(name string, n int){
		for range n {
			c.Inc(name)
		}
	}
	wg.Add(1)
	go func(){
		defer wg.Done()
		doIncrements("a",10000)
	}()

	wg.Add(1)
	go func(){
		defer wg.Done()
		doIncrements("a",10000)
	}()

	wg.Add(1)
	go func(){
		defer wg.Done()
		doIncrements("b",10000)
	}()

	wg.Wait()
	fmt.Println("counters:", c.counters)
}