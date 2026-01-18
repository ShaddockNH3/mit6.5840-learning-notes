package grammar_test

import (
	"fmt"
	"testing"
	"time"
	"sync"
)

func worker_worker_pool(id int, jobs <-chan int, results chan<- int) {
	for j := range jobs {
		fmt.Printf("Worker %d started job %d\n", id, j)
		time.Sleep(time.Second) // 模拟工作时间
		fmt.Printf("Worker %d finished job %d\n", id, j)
		results <- j * 2
	}
}

// 工作池
func TestWorkerPool(t *testing.T) {
	const numJobs = 5
	jobs := make(chan int, numJobs)
	results := make(chan int, numJobs)

	// 启动 3 个工作协程，也就是招募 3 个工人
	// 每个工人从 jobs 通道接收任务，完成后将结果发送到 results 通道
	// 只是招募，他们会时时刻刻地从 jobs 通道中等待任务
	// 因为有一个 for range 循环在不停地监听 jobs 通道
	// 直到 jobs 通道被关闭
	for w := 1; w <= 3; w++ {
		go worker_worker_pool(w, jobs, results)
	}
	for j := 1; j <= numJobs; j++ {
		jobs <- j
	}
	close(jobs)

	for a := 1; a <= numJobs; a++ {
		<-results
	}
}

func worker_wait_group(id int) {
	fmt.Printf("Worker %d starting\n", id)
	time.Sleep(time.Second) // 模拟工作时间
	fmt.Printf("Worker %d done\n", id)
}

// WaitGroups 等待组
// WaitGroups 是用来等待一组协程完成的同步原语
// wg.Go 属于理想中的 API，实际并不存在
// wg 只有三个方法：Add、Done 和 Wait
// Add 用于增加等待的协程数量
// Done 用于减少等待的协程数量
// Wait 用于阻塞直到等待的协程数量为零
func TestWaitGroups(t *testing.T) {
	var wg sync.WaitGroup
	for i := 1; i <= 100; i++ {
		// wg.Go(func ()  {
		// 	worker_wait_group(i)
		// })
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			worker_wait_group(id)
		}(i)
	}
	wg.Wait()
	fmt.Println("All workers done")
}

// 速率限制器
// 用于突发场景
// 例如，固定的 timer 只允许每秒处理 3 个请求
// 但是有时会有突发的 5 个请求同时到达
// 所以这个时候可以提前预存一些令牌，允许突发的请求处理
// 这里设置 burstyLimiter 为 3
// 前三个请求几乎是同时处理的，后续的请求则是按照每 200ms 一个的速率处理的
func TestRateLimiter_Bursty(t *testing.T) {
	// 普通的速率限制器
	requests := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		requests <- i
	}
	close(requests)

	limiter := time.Tick(200 * time.Millisecond)

	for req := range requests {
		<-limiter
		fmt.Printf("Request %d processed at %v\n", req, time.Now())
	}

	// 突发的速率限制器
	burstyLimiter := make(chan time.Time, 3)

	for range 3{
		burstyLimiter <- time.Now()
	}

	go func() {
		for t := range time.Tick(200 * time.Millisecond) {
			burstyLimiter <- t
		}
	}()

	burstyRequests := make(chan int, 5)
	for i := 1; i <= 5; i++ {
		burstyRequests <- i
	}
	close(burstyRequests)
	for req := range burstyRequests {
		<-burstyLimiter
		fmt.Printf("Bursty request %d processed at %v\n", req, time.Now())
	}
}
