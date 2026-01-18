package grammar_test

import (
	"fmt"
	"testing"
	"time"
)

func TestGoroutine(t *testing.T) {
	go func(s string){
		fmt.Println(s)
	}("Hello, goroutine!")

	time.Sleep(1 * time.Second) // Wait for the goroutine to finish
}

func PrintTripleString(s string) {
	for i := 0; i < 3; i++ {
		fmt.Println(s)
	}
}

func TestGoroutineWithPrintTripleString(t *testing.T) {
	go PrintTripleString("Hello, goroutine with PrintTripleString!")
	go PrintTripleString("Hello again from PrintTripleString!")
	time.Sleep(1 * time.Second) // Wait for the goroutine to finish
}

func PrintHundredTimesString(s string) {
	for i := 0; i < 100; i++ {
		fmt.Println(s)
	}
}

// 显著的出现输出混乱现象
func TestGoroutineWithPrintHundredTimesString(t *testing.T) {
	go PrintHundredTimesString("Hello, goroutine with PrintHundredTimesString!")
	go PrintHundredTimesString("Hello again from PrintHundredTimesString!")
	time.Sleep(2 * time.Second) // Wait for the goroutine to finish
}
