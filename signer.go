package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)


// сюда писать код

var ExecutePipeline = func(jobs ...job) {
	input := make(chan interface{})
	in := input
	for _, job := range jobs {
		out := make(chan interface{})
		go job(in, out) // run job() as a goroutine
		in = out        // the output of one job is the input of the next one
	}
	close(input)
	time.Sleep(60 * time.Second)
}

var SingleHash = func(in, out chan interface{}) {
	for i := range in {
		data := strconv.Itoa(i.(int))
		md5Data := DataSignerMd5(data)
		crc32Md5Data := DataSignerCrc32(md5Data)
		crc32Data := DataSignerCrc32(data)
		fmt.Printf("%s SingleHash data %s\n", data, data)
		fmt.Printf("%s SingleHash md5(data) %s\n", data, md5Data)
		fmt.Printf("%s SingleHash crc32(md5(data)) %s\n", data, crc32Md5Data)
		fmt.Printf("%s SingleHash crc32(data) %s\n", data, crc32Data)
		fmt.Printf("%s SingleHash result %s\n", data, crc32Data+"~"+crc32Md5Data)
		out <- crc32Data + "~" + crc32Md5Data
	}
}

var MultiHash = func(in, out chan interface{}) {
	for i := range in {
		var sum = ""
		for j := 0; j < 6; j++ {
			crc32 := DataSignerCrc32(strconv.Itoa(j) + i.(string))
			fmt.Printf("%s MultiHash: crc32(th+step1)) %d %s\n", i.(string), j, crc32)
			sum += crc32
		}

		fmt.Printf("%s MultiHash result:\n%s\n", i.(string), sum)
		out <- sum
	}

}

var CombineResults = func(in, out chan interface{}) {
	var count = 0
	var array []string
	for i := range in {
		count++
		array = append(array, i.(string))
		if count > 6 {
		close(in)
		}
	}
	sort.Strings(array)
	result := strings.Join(array, "_")
	fmt.Printf("CombineResults \n%s\n", result)
	out <- result
}
