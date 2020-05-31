package main

import (
	"fmt"
	"runtime"
)

func main()  {
	//fmt.Println(fmt.Sprintf("%d", 1))
	//fmt.Println(DataSignerCrc32(fmt.Sprintf("%d", 1) + "4108050209~502633748"))
	//fmt.Println(MultiHash(SingleHash("0")) + "_" + MultiHash(SingleHash("1")))
	inputData := []int{0,1}
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

	ExecutePipeline(jobs)
	fmt.Scanln()
}

var SingleHash = func(in, out chan interface{}) {
	data := fmt.Sprintf("%d", <-in)
	fmt.Println(data)
	result := DataSignerCrc32(data) + "~" + DataSignerCrc32(DataSignerMd5(data))
	fmt.Println(result)
	out <- result
}
var MultiHash = func(in, out chan interface{}) {
	data := fmt.Sprintf("%s", <-in)
	result := ""
	for i := 0; i < 6; i++ {
		tmp := DataSignerCrc32(fmt.Sprintf("%d", i) + data)
		fmt.Println(tmp)
		result += tmp
		runtime.Gosched()
	}
	out <- result
}

var CombineResults = func(in, out chan interface{}) {
	result := ""
	for val := range in {
		result += fmt.Sprintf("%s", val)
		fmt.Println(result)
		runtime.Gosched()
	}
	out <- result
}

var ExecutePipeline = func(jobs []job) {
	//out := make(chan interface{}, 1)

	//pipeline := jobs[:len(jobs) - 1]
	//lastJob := jobs[len(jobs) - 1:]

	fmt.Println("test")
	tmpIn := make(chan interface{}, 100)
	tmpOut := make(chan interface{})
	for _, job := range jobs {
		go job(tmpIn, tmpOut)
		tmpIn = tmpOut
		tmpOut = make(chan interface{})
	}
	//go lastJob[0](tmpIn, out)
}