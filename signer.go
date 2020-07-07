package main

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
	"sync"
)

var wg = sync.WaitGroup{}

type HashOrderPair struct {
	Hash  string
	Order int
}

func main() {
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
	crc32Chan := make(chan []string, 100)

	var wg1 sync.WaitGroup
	for data := range in {
		str := fmt.Sprintf("%d", data)

		md5 := DataSignerMd5(str)

		wg1.Add(1)
		go GroupDataSignerCrc32(crc32Chan, &wg1, []string{str, md5}...)
	}

	wg1.Wait()
	close(crc32Chan)

LOOP:
	for {
		select {
		case crc32, ok := <-crc32Chan:
			if !ok {
				break LOOP
			}
			out <- strings.Join(crc32, "~")
		}
	}
}

var MultiHash = func(in, out chan interface{}) {
	wg2 := sync.WaitGroup{}

	crc32Chan := make(chan []string, 100)
	for data := range in {
		str := fmt.Sprintf("%s", data)

		var stringSlice []string
		for i := 0; i < 6; i++ {
			stringSlice = append(stringSlice, fmt.Sprintf("%d", i)+str)
		}
		wg2.Add(1)
		go GroupDataSignerCrc32(crc32Chan, &wg2, stringSlice...)
	}

	wg2.Wait()
	close(crc32Chan)

LOOP:
	for {
		select {
		case tmp, ok := <-crc32Chan:
			if !ok {
				break LOOP
			}
			str := strings.Join(tmp, "")
			out <- str
		}
	}
}
var CombineResults = func(in, out chan interface{}) {
	var tmp []string
	for val := range in {
		MultiHash := fmt.Sprintf("%s", val)
		tmp = append(tmp, MultiHash)
	}

	sort.Strings(tmp)
	out <- strings.Join(tmp, "_")
}

var GroupDataSignerCrc32 = func(dataCh chan []string, wg *sync.WaitGroup, strings ...string) {
	defer wg.Done()
	hashMap := make(map[int]string)
	var output []string
	var wg3 sync.WaitGroup
	ch := make(chan HashOrderPair, 600)
	for key, str := range strings {
		wg3.Add(1)
		go AsyncGroupDataSignerCrc32(ch, &wg3, str, key)
	}

	wg3.Wait()
	close(ch)

LOOP:
	for {
		select {
		case hash, ok := <-ch:
			if !ok {
				break LOOP
			}
			hashMap[hash.Order] = hash.Hash
		}
		runtime.Gosched()
	}

	keys := make([]int, 0, len(hashMap))
	for k := range hashMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, key := range keys {
		output = append(output, hashMap[key])
	}

	dataCh <- output
}

var AsyncGroupDataSignerCrc32 = func(dataCh chan HashOrderPair, wg *sync.WaitGroup, str string, order int) {
	defer wg.Done()
	pair := HashOrderPair{Hash: DataSignerCrc32(str), Order: order}
	dataCh <- pair
}
