package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func main() {

}

type WorkerPool struct {
	WorkerCount int32
	Wg          *sync.WaitGroup
	WorkerWg    *sync.WaitGroup
	Stopping    bool
	Input       chan func()
	Remove      chan bool
}

func StartWorkerPool(cnt int, input chan func(), remove chan bool, tasks []func()) (*WorkerPool, error) {
	if cnt < 0 {
		return nil, fmt.Errorf("Worker count must not be negative")
	}

	wg := &sync.WaitGroup{}
	wg.Add(len(tasks))
	workerWg := &sync.WaitGroup{}
	workerWg.Add(cnt)

	wp := &WorkerPool{
		WorkerWg: workerWg,
		Wg:       wg,
		Stopping: false,
		Input:    input,
		Remove:   remove}

	wp.WorkerCount = int32(cnt)

	for i := 0; i < cnt; i++ {
		go wp.StartWorker()
	}

	for _, t := range tasks {
		input <- t
	}

	return wp, nil
}

func (w *WorkerPool) AddWorker() error {
	if w.Stopping {
		return fmt.Errorf("Worker pool is stopping now")
	}

	atomic.AddInt32(&w.WorkerCount, 1)
	w.WorkerWg.Add(1)
	go w.StartWorker()
	return nil
}

func (w *WorkerPool) RemoveWorker() error {
	if w.Stopping {
		return fmt.Errorf("Worker pool is stopping now")
	}

	if atomic.LoadInt32(&w.WorkerCount) == 1 {
		return fmt.Errorf("Worker count cannot be less than 1")
	}

	w.Remove <- true
	return nil
}

func (w *WorkerPool) AddTask(task func()) error {
	if w.Stopping {
		return fmt.Errorf("Worker pool is stopping now")
	}

	w.Wg.Add(1)
	w.Input <- task

	return nil
}

func (w *WorkerPool) Wait() {
	w.Wg.Wait()
}

func (w *WorkerPool) Stop() error {
	// надо остаток запросов на старт принять
	if w.Stopping {
		return fmt.Errorf("Already stopping")
	}

	w.Stopping = true

	// сначала ждём выполнения всех тасков
	w.Wg.Wait()

	// потом ждём завершения воркеров

	cnt := int(atomic.LoadInt32(&w.WorkerCount))
	for i := 0; i < cnt; i++ {
		w.Remove <- true
	}

	w.WorkerWg.Wait()

	// и закрываем каналы
	close(w.Input)
	close(w.Remove)

	return nil
}

func (w *WorkerPool) StartWorker() {
	for {
		select {
		case executeJob := <-w.Input:
			executeJob()
			w.Wg.Done()
		case <-w.Remove:
			atomic.AddInt32(&w.WorkerCount, -1)
			defer w.WorkerWg.Done()
			return
		}
	}
}
