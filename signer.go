package main

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
)

var wg = &sync.WaitGroup{}

func main() {
	//fmt.Println(fmt.Sprintf("%d", 1))
	//fmt.Println(DataSignerCrc32(fmt.Sprintf("%d", 1) + "4108050209~502633748"))
	//fmt.Println(MultiHash(SingleHash("0")) + "_" + MultiHash(SingleHash("1")))
	inputData := []int{0, 1, 1, 2, 3, 5, 8}
	jobs := []job{
		job(func(in, out chan interface{}) {
			for _, fibNum := range inputData {
				out <- fibNum
			}
		}),
		job(SingleHash),
		job(MultiHash),
		job(CombineResults),
		job(func(in, out chan interface{}) {
			dataRaw := <-in
			data, ok := dataRaw.(string)
			if !ok {
				fmt.Println("cant convert result data to string")
			}
			fmt.Println(data)
		}),
	}

	ExecutePipeline(jobs...)

	//var recieved uint32
	//freeFlowJobs := []job{
	//	job(func(in, out chan interface{}) {
	//		out <- uint32(1)
	//		out <- uint32(3)
	//		out <- uint32(4)
	//	}),
	//	job(func(in, out chan interface{}) {
	//		for val := range in {
	//			out <- val.(uint32) * 3
	//			time.Sleep(time.Millisecond * 100)
	//		}
	//	}),
	//	job(func(in, out chan interface{}) {
	//		for val := range in {
	//			fmt.Println("collected", val)
	//			atomic.AddUint32(&recieved, val.(uint32))
	//		}
	//	}),
	//}
	//
	//start := time.Now()
	//
	//ExecutePipeline(freeFlowJobs...)
	//
	//end := time.Since(start)
	//
	//expectedTime := time.Millisecond * 350
	//
	//fmt.Println(end, expectedTime)
}

var ExecutePipeline = func(jobs ...job) {
	first := make(chan interface{}, 100)
	tmpIn := first
	tmpOut := make(chan interface{}, 100)
	for _, job := range jobs {
		wg.Add(1)
		go PipelineStep(job, tmpIn, tmpOut)

		tmpIn = tmpOut
		tmpOut = make(chan interface{}, 100)
	}
	wg.Wait()
}

var PipelineStep = func(job job, in, out chan interface{}) {
	defer wg.Done()
	job(in, out)
	close(out)
}

var SingleHash = func(in, out chan interface{}) {
	for data := range in {
		str := fmt.Sprintf("%d", data)
		fmt.Println(data)

		crc32Chan := make(chan string)
		go AsyncDataSignerCrc32(str, crc32Chan, make(chan struct{}))
		runtime.Gosched()

		md5Chan := make(chan string)
		go AsyncDataSignerMd5(str, md5Chan, make(chan struct{}))
		runtime.Gosched()

		crc32md5Chan := make(chan string)
		cancelCh := make(chan struct{})
		result := ""
	LOOP:
		for {
			select {
			case crc32 := <-crc32Chan:
				result += crc32
			case crc32md5 := <-crc32md5Chan:
				result += "~" + crc32md5
			case md5 := <-md5Chan:
				go AsyncDataSignerCrc32(md5, crc32md5Chan, cancelCh)
				runtime.Gosched()
			case <-cancelCh:
				close(crc32Chan)
				close(md5Chan)
				close(crc32md5Chan)
				close(cancelCh)
				break LOOP
			}
		}

		fmt.Println(result)
		out <- result
		runtime.Gosched()
	}
}
var MultiHash = func(in, out chan interface{}) {
	wg := &sync.WaitGroup{}

	for data := range in {
		str := fmt.Sprintf("%s", data)
		crc32Chans := []chan string{
			make(chan string),
			make(chan string),
			make(chan string),
			make(chan string),
			make(chan string),
			make(chan string),
		}

		wg.Add(1)
		go func(out chan interface{}, crc32Chans []chan string) {
			defer wg.Done()
			//i := 0
			result := make(map[int]string)
			for {
				select {
				case tmp := <-crc32Chans[0]:
					fmt.Println(tmp)
					result[0] = tmp
					if len(result) == 6 {
						out <- stringMapToString(result)
						return
					}
				case tmp := <-crc32Chans[1]:
					fmt.Println(tmp)
					result[1] = tmp
					if len(result) == 6 {
						out <- stringMapToString(result)
						return
					}
				case tmp := <-crc32Chans[2]:
					fmt.Println(tmp)
					result[2] = tmp
					if len(result) == 6 {
						out <- stringMapToString(result)
						return
					}
				case tmp := <-crc32Chans[3]:
					fmt.Println(tmp)
					result[3] = tmp
					if len(result) == 6 {
						out <- stringMapToString(result)
						return
					}
				case tmp := <-crc32Chans[4]:
					fmt.Println(tmp)
					result[4] = tmp
					if len(result) == 6 {
						out <- stringMapToString(result)
						return
					}
				case tmp := <-crc32Chans[5]:
					fmt.Println(tmp)
					result[5] = tmp
					if len(result) == 6 {
						out <- stringMapToString(result)
						return
					}
				}
			}
		}(out, crc32Chans)

		for i := 0; i < 6; i++ {
			cCh := make(chan struct{})
			go AsyncDataSignerCrc32(fmt.Sprintf("%d", i)+str, crc32Chans[i], cCh)
		}
	}

	wg.Wait()
}
var CombineResults = func(in, out chan interface{}) {
	var tmp []string
	for val := range in {
		MultiHash := fmt.Sprintf("%s", val)
		fmt.Println(MultiHash)
		tmp = append(tmp, MultiHash)
	}

	result := ""
	length := len(tmp)
	sort.Strings(tmp)
	for i := range tmp {
		if i != length-1 {
			result += tmp[i] + string('_')
		} else {
			result += tmp[i]
		}
	}
	out <- result
}

var AsyncDataSignerCrc32 = func(str string, dataCh chan string, cancelCh chan struct{}) {
	dataCh <- DataSignerCrc32(str)
	cancelCh <- struct{}{}
}
var AsyncDataSignerMd5 = func(str string, dataCh chan string, cancelCh chan struct{}) {
	dataCh <- DataSignerMd5(str)
	cancelCh <- struct{}{}
}
var stringMapToString = func(m map[int]string) string {
	s := ""
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, key := range keys {
		s += fmt.Sprintf("%s", m[key])
	}
	return s
}
