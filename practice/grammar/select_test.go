package grammar_test

import (
	"fmt"
	"testing"
	"time"
)

// select 允许一个 goroutine 等待多个通信操作
func TestSelectBasic(t *testing.T) {
	ch1 := make(chan string)
	ch2 := make(chan string)
	go func() {
		ch1 <- "Message from channel 1"
	}()
	go func() {
		ch2 <- "Message from channel 2"
	}()
	for range 2 {
		select {
		case msg1 := <-ch1:
			fmt.Println("Received:", msg1)
		case msg2 := <-ch2:
			fmt.Println("Received:", msg2)
		}
	}
}

func TestSelectBasic_2(t *testing.T) {
	ch1 := make(chan string)
	ch2 := make(chan string)
	go func() {
		ch1 <- "shenmi"
	}()
	go func() {
		ch2 <- "shenmidazhi"
	}()
	for range 1 {
		select {
		case msg1 := <-ch1:
			fmt.Println("received", msg1)
		case msg2 := <-ch2:
			fmt.Println("received", msg2)
		}
	}
}

// 如果放在通信的场景，直接写一个 for 死循环来处理不断到来的消息

// timeouts
// 超时机制，防止 goroutine 永远等待下去
func TestSelectTimeouts(t *testing.T) {
	ch1 := make(chan string, 1)
	go func() {
		time.Sleep(2 * time.Second)
		ch1 <- "result 1"
	}()
	select {
	case res := <-ch1:
		fmt.Println(res)
	case <-time.After(1 * time.Second):
		fmt.Println("timeout 1s")
	}

	ch2 := make(chan string, 1)
	go func() {
		time.Sleep(1 * time.Second)
		ch2 <- "result 2"
	}()
	select {
	case res := <-ch2:
		fmt.Println(res)
	case <-time.After(2 * time.Second):
		fmt.Println("timeout 2s")
	}
}

// 无阻塞信道操作
// 其实就是 select 的 default 分支
// 如果没有准备好的通信，default 分支就会被执行，从而实现非阻塞的信道操作
// 因为没有并行的程序，所以下面的例子中，信道操作都不会成功
// 如果要看到成功，得 go func 起来
func TestSelectNonBlocking(t *testing.T) {
	messages := make(chan string)
	signals := make(chan bool)

	select {
	case msg := <-messages:
		fmt.Println("received message", msg)
	default:
		fmt.Println("no message received")
	}

	msg := "hi"

	select {
	case messages <- msg:
		fmt.Println("received message", msg)
	default:
		fmt.Println("no message received")
	}

	select {
	case msg := <-messages:
		fmt.Println("received message", msg)
	case sig := <-signals:
		fmt.Println("received signal", sig)
	default:
		fmt.Println("no activity")
	}
}
