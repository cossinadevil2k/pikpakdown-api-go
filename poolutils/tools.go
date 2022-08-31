package main

import (
	"fmt"
	"golang.org/x/net/context"
	"time"
)

func AsncCall() {
	timeout, cancelFunc := context.WithTimeout(context.Background(), time.Second*10)
	defer cancelFunc()
	for i := 0; i < 10; i++ {
		go func(i int, timeout context.Context) {
			time.Sleep(time.Second * time.Duration(i))
			fmt.Printf("call %d\n", i)
		}(i, timeout)
	}

	select {
	case <-timeout.Done():
		fmt.Printf("call successFully!")
		return
	case <-time.After(time.Second * time.Duration(60)):
		fmt.Printf("call error!")
		return
	}
}

func main() {
	AsncCall()
}
