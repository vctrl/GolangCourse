package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	thCnt = 6
)

// ExecutePipeline тащит движ
func ExecutePipeline(jobs ...job) {
	wg := &sync.WaitGroup{}
	wg.Add(len(jobs))

	in := make(chan interface{})
	for i := 0; i < len(jobs); i++ {
		out := make(chan interface{})
		go func(i int, in, out chan interface{}) {
			jobs[i](in, out)
			close(out)
			wg.Done()
		}(i, in, out)

		in = out
	}

	wg.Wait()
}

// SingleHash считает один хэш
func SingleHash(in, out chan interface{}) {
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for v := range in {
		wg.Add(1)
		go func(v interface{}) {
			n := v.(int)
			s := strconv.Itoa(n)

			crc32md5ch := make(chan string)
			go func(n int, c chan string) {
				mu.Lock()
				md5val := DataSignerMd5(s)
				mu.Unlock()

				c <- DataSignerCrc32(md5val)
			}(n, crc32md5ch)

			out <- DataSignerCrc32(s) + "~" + <-crc32md5ch
			wg.Done()
		}(v)
	}

	wg.Wait()
}

// MultiHash считает много хэшей
func MultiHash(in, out chan interface{}) {
	wg1 := &sync.WaitGroup{}
	for v := range in {
		wg1.Add(1)
		data := v.(string)
		go func(data string) {
			crc32ch := make([]chan string, thCnt)
			for i := range crc32ch {
				crc32ch[i] = make(chan string)
			}

			wg := &sync.WaitGroup{}

			for th := 0; th < thCnt; th++ {
				wg.Add(1)
				th := th
				go func() {
					crc32ch[th] <- DataSignerCrc32(strconv.Itoa(th) + data)
				}()
			}

			mu := &sync.Mutex{}
			res := make([]string, thCnt)
			go func() {
				for i := 0; i < thCnt; i++ {
					i := i
					go func() {
						mu.Lock()
						res[i] = <-crc32ch[i]
						mu.Unlock()
						wg.Done()
					}()
				}
			}()

			wg.Wait()
			out <- strings.Join(res, "")
			wg1.Done()
		}(data)

	}

	wg1.Wait()
}

// CombineResults комбинирует
func CombineResults(in, out chan interface{}) {
	res := make([]string, 0)

	for v := range in {
		v := v.(string)
		res = append(res, v)
	}

	sort.Strings(res)
	out <- strings.Join(res, "_")
}
