package grammar_test

import (
	"fmt"
	"testing"
	"time"
)

// 管道，将值从一个 goroutine 传递到另一个 goroutine 的通信机制。
func TestChannelBasics(t *testing.T) {
	message := make(chan string)
	go func() {
		message <- "Hello, Channels!"
	}()
	msg := <-message
	fmt.Println(msg)
}

// 尝试多次从管道接收值，并发执行，顺序不确定
// 如果尝试从空管道接收值，程序将会阻塞，直到有值可用。
func TestChannel_1(t *testing.T) {
	messages := make(chan string)
	go func() {
		messages <- "ping"
	}()
	go func() {
		messages <- "shenmi"
	}()
	msg := <-messages
	fmt.Println(msg)
	msg = <-messages
	fmt.Println(msg)
	// msg = <-messages
	// fmt.Println(msg)
}

// 管道缓冲 Buffer
func TestChannelBuffer(t *testing.T) {
	messages := make(chan string, 2)
	messages <- "buffered"
	messages <- "channel"
	fmt.Println(<-messages)
	fmt.Println(<-messages)
}

func TestChannel_2(t *testing.T) {
	messages := make(chan string, 2)
	messages <- "buffered"
	messages <- "channel"

	msg1 := <-messages
	msg2 := <-messages

	fmt.Println(msg1)
	fmt.Println(msg2)
	// fmt.Println(<-messages) // 这里会阻塞，因为管道已空
}

// 通道同步，<- 适合一对一等待，WaitGroup 适合多对多等待
func worker_3(done chan bool) {
	fmt.Println("working...")
	time.Sleep(1 * time.Second)
	fmt.Println("done")
	done <- true
}

func TestChannelSynchronization(t *testing.T) {
	done := make(chan bool, 1)
	go worker_3(done) // go 这个意思是主程序不用等待 worker_3 执行完再继续往下走
	<-done
}

func TestChannelSynchronization_2(t *testing.T) {
	done := make(chan bool, 2)
	go worker_3(done)
	go worker_3(done)
	<-done
	// 少了一个 <-done 会导致主程序提前结束，无法等待第二个 worker_3 完成
	// 当然如果是 test 环境下，可能会在 pass 后冒出来
}

func ping(pipeIn chan<- string, msg string) {
	pipeIn <- msg // 这里是作为接受者
}

func pong(pipeOut chan<- string, pipeIn <-chan string) {
	msg := <-pipeIn // 这里是作为发送者
	pipeOut <- msg // 这里是作为接受者
}

// 我这边一直误解了一个地方，就是 go 的 chan
// 事实上，之前我一直把 chan<- type 和 <-chan type 理解成谁往谁发送
// 不能这么理解，而是要分开来理解
// chan<- 代表这个管道在这个函数里只能发送数据，不能接收数据
// <-chan 代表这个管道在这个函数里只能接收数据，不能发送数据
// 而理解完了上面，再去理解 type，type 是这个管道的数据类型
// 这样就能理解清楚了

func TestChannelDirection_Buffered(t *testing.T) {
	pipeA := make(chan string, 1)
	pipeB := make(chan string, 1)

	ping(pipeA, "passed message")
	pong(pipeB, pipeA)

	fmt.Println(<-pipeB)
}

func TestChannelDirection_Unbuffered(t *testing.T) {
	pipeA := make(chan string)
	pipeB := make(chan string)

	go ping(pipeA, "passed message")
	go pong(pipeB, pipeA)

	fmt.Println(<-pipeB)
}
