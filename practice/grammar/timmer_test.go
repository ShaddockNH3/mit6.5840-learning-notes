package grammar_test

import (
	"fmt"
	"testing"
	"time"
)

// 第一部分的 timer 作用和 time.Sleep 差不多
// 但是是通过 channel 来通知的
// 第二部分的 timer 可以通过 Stop 方法来停止计时器
// 也就是可以反悔的 sleep
func TestTimers(t *testing.T) {
	timer1 := time.NewTimer(2 * time.Second)
	<-timer1.C
	fmt.Println("Timer 1 expired")

	timer2 := time.NewTimer(1 * time.Second)
	go func() {
		<-timer2.C
		fmt.Println("Timer 2 expired")
	}()
	stop2 := timer2.Stop()
	if stop2 {
		fmt.Println("Timer 2 stopped before expiration")
	}
	time.Sleep(3 * time.Second)
}

// tickers
// ticker 是一个定时器，可以周期性的触发事件
func TestTickers(t *testing.T) {
	ticker := time.NewTicker(500 * time.Millisecond)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("Ticker ticked at", t)
			}
		}
	}()

	time.Sleep(1600 * time.Millisecond)
	ticker.Stop()
	done <- true
	fmt.Println("Ticker stopped")
}