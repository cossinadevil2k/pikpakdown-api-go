package main

import (
	"encoding/hex"
	"fmt"
)

func hexConvert() {
	input := []byte{172, 80, 85, 37, 168, 146, 153, 228, 82, 28, 6, 0, 192, 186}
	in := hex.EncodeToString(input)
	fmt.Println("in:%s", in)
}

func main() {

	hexConvert()

}
