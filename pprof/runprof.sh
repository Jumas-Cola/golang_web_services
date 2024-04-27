#!/bin/sh
go test -bench . -benchmem -cpuprofile=cpu.out -memprofile=mem.out -memprofilerate=1
# go tool pprof -http=:8083 ./hw3.test mem.out
go tool pprof -http=:8083 ./hw3.test cpu.out
