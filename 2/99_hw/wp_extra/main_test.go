package main

import (
	"runtime"
	"testing"
	"time"
)

func TestConcurrent(t *testing.T) {
	start := time.Now()
	wp, _ := CreateAndStart(5, 5, 1)
	// без этой строчки может дедлокнуть почему-то
	wp.Wait()
	end := time.Since(start)
	expectedTime := time.Second + time.Millisecond*10
	if end > expectedTime {
		t.Errorf("execition too long\nGot: %s\nExpected: <%s", end, expectedTime)
	}
}

func TestGoroutineLeak(t *testing.T) {
	goroutinesStart := runtime.NumGoroutine()

	wp, _ := CreateAndStart(5, 5, 1)

	wp.Stop()

	goroutinesCount := runtime.NumGoroutine() - goroutinesStart
	if goroutinesCount > 0 {
		t.Fatalf("looks like you have goroutines leak: %+v", goroutinesCount)
	}
}

func TestSequential(t *testing.T) {
	start := time.Now()
	wp, _ := CreateAndStart(1, 5, 1)

	expectedTimeMin := time.Second * 5
	expectedTimeMax := time.Second * 6

	wp.Wait()
	end := time.Since(start)

	if end < expectedTimeMin {
		t.Errorf("execition too short\nGot: %s\nExpected: <%s", end, expectedTimeMin)
	}

	if end > expectedTimeMax {
		t.Errorf("execition too long\nGot: %s\nExpected: <%s", end, expectedTimeMax)
	}

}

func TestWorkerCountCannotBeZero(t *testing.T) {
	_, err := CreateAndStart(0, 5, 1)
	if err != nil {
		t.Errorf("error should not be nil")
	}
}

func TestAddTasks(t *testing.T) {
	start := time.Now()
	wp, _ := CreateAndStart(1, 1, 1)
	wp.AddTask(func() { time.Sleep(time.Second) })
	wp.AddTask(func() { time.Sleep(time.Second) })
	wp.Wait()

	end := time.Since(start)
	expectedTime := time.Second*3 + time.Millisecond*20
	if end > expectedTime {
		t.Errorf("execition too long\nGot: %s\nExpected: <%s", end, expectedTime)
	}
}

func TestAddWorkers(t *testing.T) {
	wp, _ := CreateAndStart(1, 1, 1)
	wp.AddWorker()
	wp.AddWorker()
	wp.AddWorker()

	start := time.Now()
	wp.AddTask(func() { time.Sleep(time.Second) })
	wp.AddTask(func() { time.Sleep(time.Second) })
	wp.AddTask(func() { time.Sleep(time.Second) })

	wp.Wait()
	end := time.Since(start)
	expectedTime := time.Second*1 + time.Millisecond*10

	if end > expectedTime {
		t.Errorf("execution too long\nGot: %s\nExpected: <%s", end, expectedTime)
	}
}

func TestRemoveWorkers(t *testing.T) {
	wp, _ := CreateAndStart(4, 1, 1)
	wp.RemoveWorker()
	wp.RemoveWorker()
	wp.RemoveWorker()

	// должен остаться один воркер, который последовательно обработает таски

	start := time.Now()
	wp.AddTask(func() { time.Sleep(time.Second) })
	wp.AddTask(func() { time.Sleep(time.Second) })

	wp.Wait()

	end := time.Since(start)
	expectedTime := time.Second * 3

	if end < expectedTime {
		t.Errorf("execution too short\nGot: %s\nExpected: <%s", end, expectedTime)
	}
}

func createTasks(n, duration int) []func() {
	res := make([]func(), n)
	for i := 0; i < n; i++ {
		res[i] = func() {
			time.Sleep(time.Second * time.Duration(duration))
		}
	}

	return res
}

func CreateAndStart(goroutineNum, tasksNum, duration int) (*WorkerPool, error) {
	cnt := goroutineNum
	input := make(chan func(), 100)
	remove := make(chan bool)

	tasks := createTasks(tasksNum, duration)
	wp, err := StartWorkerPool(cnt, input, remove, tasks)

	return wp, err
}
