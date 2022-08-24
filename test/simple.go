package main

import (
	"encoding/hex"
	"fmt"
	"github.com/imroc/req/v3"
	"golang.org/x/net/html/charset"
)

func hexConvert() {
	input := []byte{172, 80, 85, 37, 168, 146, 153, 228, 82, 28, 6, 0, 192, 186}
	in := hex.EncodeToString(input)
	fmt.Println("in:%s", in)
}

func transform() {
	client := req.C()
	get, err := client.R().Get("http://getchu.com/")
	if err != nil {
		panic(err)
	}
	fmt.Println(get.String())
	e, name, certain := charset.DetermineEncoding(get.Bytes(), "")
	fmt.Printf("编码:%v\n名称:%s\n 确定: %t\n", e, name, certain)
}

func main() {

	transform()

}
