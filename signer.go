package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const TH = 6

var ExecutePipeline = func(jobs ...job) {

	wg := &sync.WaitGroup{}
	in := make(chan interface{})

	for _, job := range jobs {

		wg.Add(1)

		out := make(chan interface{})
		go jobWorker(job, in, out, wg)
		in = out
	}

	wg.Wait()
}

func jobWorker(job job, in, out chan interface{}, wg *sync.WaitGroup) {

	defer wg.Done()
	defer close(out)

	job(in, out)
}

var SingleHash = func(in, out chan interface{}) {

	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}

	for i := range in {

		wg.Add(1)

		go singleHashWorker(i, out, wg, mu)
	}

	wg.Wait()
}

func singleHashWorker(in interface{}, out chan interface{}, wg *sync.WaitGroup, mu *sync.Mutex) {

	defer wg.Done()

	data := strconv.Itoa(in.(int))

	mu.Lock()
	md5Data := DataSignerMd5(data)
	mu.Unlock()

	crc32DataChan := make(chan string)
	go crc32Parallel(data, crc32DataChan)
	crc32Md5Data := DataSignerCrc32(md5Data)
	crc32Data := <-crc32DataChan

	fmt.Printf("%s SingleHash data %s\n", data, data)
	fmt.Printf("%s SingleHash md5(data) %s\n", data, md5Data)
	fmt.Printf("%s SingleHash crc32(md5(data)) %s\n", data, crc32Md5Data)
	fmt.Printf("%s SingleHash crc32(data) %s\n", data, crc32Data)
	fmt.Printf("%s SingleHash result %s\n", data, crc32Data+"~"+crc32Md5Data)

	out <- crc32Data + "~" + crc32Md5Data
}

func crc32Parallel(data string, out chan string) {

	out <- DataSignerCrc32(data)
}

var MultiHash = func(in, out chan interface{}) {

	wg := &sync.WaitGroup{}

	for i := range in {

		wg.Add(1)

		go multiHashWorker(i.(string), out, wg)
	}

	wg.Wait()
}

func multiHashWorker(in string, out chan interface{}, wg *sync.WaitGroup) {

	defer wg.Done()

	mu := &sync.Mutex{}
	wgCrc32 := &sync.WaitGroup{}

	concatArray := make([]string, TH)

	for i := 0; i < TH; i++ {

		wgCrc32.Add(1)

		data := strconv.Itoa(i) + in

		go func(concatArray []string, data string, index int, wg *sync.WaitGroup, mu *sync.Mutex) {

			defer wg.Done()

			data = DataSignerCrc32(data)

			mu.Lock()
			concatArray[index] = data
			fmt.Printf("%s MultiHash: crc32(th+step1)) %d %s\n", in, index, data)
			mu.Unlock()

		}(concatArray, data, i, wgCrc32, mu)
	}

	wgCrc32.Wait()

	result := strings.Join(concatArray, "")

	fmt.Printf("%s MultiHash result: %s\n", in, result)

	out <- result
}

var CombineResults = func(in, out chan interface{}) {

	var array []string

	for i := range in {

		array = append(array, i.(string))
	}

	sort.Strings(array)
	result := strings.Join(array, "_")

	fmt.Printf("CombineResults \n%s\n", result)

	out <- result
}
