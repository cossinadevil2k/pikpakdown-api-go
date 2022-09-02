package main

import (
	"context"
	"fmt"
	"time"
)

func fn1(ctx context.Context) {
	log("start fn1")
	defer log("done fn1")
	//for i := 1; i <= 10000; i++ {
	//	//select {
	//	//case <-ctx.Done():
	//	//	return
	//	//default:
	//	//	log("loop fn1")
	//	//	time.Sleep(1 * time.Second)
	//	//}
	//	log("loop fn1")
	//	time.Sleep(1 * time.Second)
	//}
	// 虽然传了 context 但是没有用...所以也没用
	for true {
		log("loop fn1")
		time.Sleep(1 * time.Second)
	}
}

func fn2(ctx context.Context) {
	log("start fn2")
	defer log("done fn2")
	for {
		select {
		case <-ctx.Done():
			fmt.Println("响应...退出")
			return
		default:
			log("loop fn2")
		}
		//log("loop fn2")
	}
}

func log(timing string) {
	fmt.Printf("%s second:%v\n", timing, time.Now().Second())
}

func main1() {
	log("start main")
	//defer log("done main")
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10000*time.Second)
	//defer cancel()

	go fn1(ctx)
	go fn2(ctx)
	cancel() //立马取消
	time.Sleep(5 * time.Second)

	log("done main")
	for true {

	}
}
