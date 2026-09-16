package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/petermattis/goid"
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

var thrd1Mu sync.Mutex
var thrd2Mu sync.Mutex
var thrd1Cond = sync.NewCond(&thrd1Mu)
var thrd2Cond = sync.NewCond(&thrd2Mu)

// 定义一个协程函数，模拟一个耗时操作
func myroutine1(done chan struct{}) {
	defer close(done)
	goid.Get()
	fmt.Println("我的协程函数1 begin execute")
	for i := 0; i < 3; i++ {
		fmt.Printf("协程1：第%d帧运行 \n", i)
		thrd1Mu.Lock()
		thrd1Cond.Wait()
		thrd1Mu.Unlock()
	}
	fmt.Println("我的协程函数1 end execute")
}

func myroutine2(done chan struct{}) {
	defer close(done)
	fmt.Println("我的协程函数2 begin execute")
	for i := 0; i < 3; i++ {
		fmt.Printf("协程2：第%d帧运行 \n", i)
		thrd2Mu.Lock()
		thrd2Cond.Wait()
		thrd2Mu.Unlock()
	}
	fmt.Println("我的协程函数2 end execute")
}

func main() {
	done1 := createRoutine(myroutine1)
	done2 := createRoutine(myroutine2)
	for {
		//先唤醒线程1
		thrd1Mu.Lock()
		thrd1Cond.Signal()
		thrd1Mu.Unlock()

		//唤醒线程2
		thrd2Mu.Lock()
		thrd2Cond.Signal()
		thrd2Mu.Unlock()

		//先唤醒线程1
		thrd1Mu.Lock()
		thrd1Cond.Signal()
		thrd1Mu.Unlock()

		//唤醒线程2
		thrd2Mu.Lock()
		thrd2Cond.Signal()
		thrd2Mu.Unlock()

		//先唤醒线程1
		thrd1Mu.Lock()
		thrd1Cond.Signal()
		thrd1Mu.Unlock()

		//唤醒线程2
		thrd2Mu.Lock()
		thrd2Cond.Signal()
		thrd2Mu.Unlock()

	}
	//waitWithTimeout(done1, -1*time.Second)
	//waitWithTimeout(done2, -1*time.Second)

	//fmt.Println("协程结束")
}
