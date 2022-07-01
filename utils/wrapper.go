package utils

import (
	"time"
)

//FuncTimer 记录函数耗时
func Timer(after func(startTime time.Time)) func() {
	start := time.Now()
	return func() {
		after(start)
	}
}

func FunctionTimer(funcName string) func() {
	handler := func(startTime time.Time) {
		Log().Info("Time taken by %s function is %s", funcName, CostTimeInfo(startTime))
	}
	return Timer(handler)
}

//Wrapper 通用
func Wrapper(start, end func()) func() {
	start()
	return func() {
		end()
	}
}
