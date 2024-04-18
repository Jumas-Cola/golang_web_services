package main

import (
	"fmt"
	"sort"
	"sync"
)

var in, out chan interface{}

// // Рабочая, но медленная
// func SingleHash(in, out chan interface{}) {
// 	hash1Ch := make(chan string)
// 	hash2Ch := make(chan string)
// 	for i := range in {
// 		data := fmt.Sprintf("%v", i)
// 		go func(data string) {
// 			hash1Ch <- DataSignerCrc32(data)
// 		}(data)
// 		go func(data string) {
// 			hash2Ch <- DataSignerCrc32(DataSignerMd5(data))
// 		}(data)
// 		out <- fmt.Sprintf("%v~%v", <-hash1Ch, <-hash2Ch)
// 	}
// }

func SingleHash(in, out chan interface{}) {
	hash1Slice := make([]chan string, 0)
	hash2Slice := make([]chan string, 0)
	dataSlice := make([]string, 0)

	index := 0
	for i := range in {
		data := fmt.Sprintf("%v", i)
		hash1Slice = append(hash1Slice, make(chan string))
		hash2Slice = append(hash2Slice, make(chan string))
		dataSlice = append(dataSlice, data)
		index++
	}

	for i := 0; i < index; i++ {
		go func(data string, index int) {
			hash1 := DataSignerCrc32(data)
			hash1Slice[index] <- fmt.Sprintf("%v", hash1)
		}(dataSlice[i], i)
		md5 := DataSignerMd5(dataSlice[i])
		go func(md5 string, index int) {
			hash2 := DataSignerCrc32(md5)
			hash2Slice[index] <- fmt.Sprintf("%v", hash2)
		}(md5, i)
	}

	for i := 0; i < index; i++ {
		out <- fmt.Sprintf("%v~%v", <-hash1Slice[i], <-hash2Slice[i])
	}
}

// // Рабочая, но медленная
// func MultiHash(in, out chan interface{}) {
// 	for i := range in {
// 		data := fmt.Sprintf("%v", i)
// 		res := ""
// 		var resArr [6]chan string
// 		for i := 0; i < 6; i++ {
// 			resArr[i] = make(chan string)
// 			go func(data string, i int, resArr [6]chan string) {
// 				resArr[i] <- DataSignerCrc32(fmt.Sprintf("%v%v", i, data))
// 			}(data, i, resArr)
// 		}
//
// 		for i := 0; i < 6; i++ {
// 			res += <-resArr[i]
// 		}
// 		out <- res
// 	}
// }

func MultiHash(in, out chan interface{}) {
	hashSlice := make([]chan string, 0)
	dataSlice := make([]string, 0)

	index := 0
	for i := range in {
		data := fmt.Sprintf("%v", i)
		dataSlice = append(dataSlice, data)
		hashSlice = append(hashSlice, make(chan string))
		index++
	}

	for i := 0; i < index; i++ {
		go func(index int) {
			data := dataSlice[index]
			res := ""
			var resArr [6]chan string
			for i := 0; i < 6; i++ {
				resArr[i] = make(chan string)
				go func(data string, i int, resArr [6]chan string) {
					resArr[i] <- DataSignerCrc32(fmt.Sprintf("%v%v", i, data))
				}(data, i, resArr)
			}

			for i := 0; i < 6; i++ {
				res += <-resArr[i]
			}
			hashSlice[index] <- res
		}(i)
	}

	for i := 0; i < index; i++ {
		out <- (<-hashSlice[i])
	}
}

func CombineResults(in, out chan interface{}) {
	res := ""
	resArr := make([]string, 0)
	for i := range in {
		resArr = append(resArr, fmt.Sprintf("%v", i))
	}

	sort.Strings(resArr)

	for _, data := range resArr {
		if res != "" {
			res += "_" + data
		} else {
			res += data
		}
	}
	out <- res
}

func ExecutePipeline(jobs ...job) {
	var wg sync.WaitGroup
	in := make(chan interface{}, 100)
	var out chan interface{}

	for _, j := range jobs {
		out = make(chan interface{}, 100)
		wg.Add(1)
		go func(j job, in, out chan interface{}) {
			defer wg.Done()
			j(in, out)
			close(out)
		}(j, in, out)
		in = out
	}
	wg.Wait()
}

// func main() {
// 	inputData := []int{0, 1, 1, 2, 3, 5, 8}
// 	// inputData := []int{0, 1}
//
// 	hashSignJobs := []job{
// 		job(func(in, out chan interface{}) {
// 			for _, fibNum := range inputData {
// 				out <- fmt.Sprintf("%v", fibNum)
// 			}
// 		}),
// 		job(SingleHash),
// 		job(MultiHash),
// 		job(CombineResults),
// 		job(func(in, out chan interface{}) {
// 			for o := range in {
// 				fmt.Println(o)
// 			}
// 		}),
// 	}
//
// 	ExecutePipeline(hashSignJobs...)
// }
