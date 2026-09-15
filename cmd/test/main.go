package main

import (
	"fmt"
	"runtime"
	"time"
)

// 超时等待函数，等待一个channel完成，或者超时
func waitWithTimeout(c chan struct{}, timeout time.Duration) bool {
	//如果timeout是负数，则表示永不超时
	if timeout < 0 {
		<-c // 直接阻塞等待，永不超时
		return true
	}

	select {
	case <-c:
		return true
	case <-time.After(timeout):
		return false
	}
}

// 定义一个协程创建的函数，传递一个协程函数，返回一个channel，用于等待协程完成
func createRoutine(f func(done chan struct{})) chan struct{} {
	done := make(chan struct{})
	go f(done)
	return done
}

// 定义一个协程函数，模拟一个耗时操作
func myroutine1(done chan struct{}) {
	defer close(done)
	fmt.Println("我的协程函数1 begin execute")
	for i := 0; i < 3; i++ {
		fmt.Printf("协程1：第%d帧运行 \n", i)
		runtime.Gosched()
	}
	fmt.Println("我的协程函数1 end execute")
}

func myroutine2(done chan struct{}) {
	defer close(done)
	fmt.Println("我的协程函数2 begin execute")
	for i := 0; i < 3; i++ {
		fmt.Printf("协程2：第%d帧运行 \n", i)
		runtime.Gosched()
	}
	fmt.Println("我的协程函数2 end execute")
}

func main() {
	// 创建两个协程，并返回对应的channel
	done1 := createRoutine(myroutine1)
	done2 := createRoutine(myroutine2)

	// 等待协程完成，设置超时时间为2秒
	if waitWithTimeout(done1, 2*time.Second) {
		fmt.Println("协程1任务完成")
	} else {
		fmt.Println("协程1等待超时")
	}
	// 等待协程完成，设置超时时间为2秒
	if waitWithTimeout(done2, 2*time.Second) {
		fmt.Println("协程2任务完成")
	} else {
		fmt.Println("协程2等待超时")
	}
}
