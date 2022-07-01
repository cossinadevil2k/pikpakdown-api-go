package utils

// If 模拟三目
//interface{}类型的变量可以指向任意类型的值
func If(isTrue bool, a, b interface{}) interface{} {
	if isTrue {
		return a
	}
	return b
}
