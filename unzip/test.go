package main

import (
	"fmt"
	"github.com/mrxtryagin/pikpakdown-api-go/myzip"
	"strconv"
	"time"
)

func test1(url string) {

	reader, err := myzip.GetZipReaderFromUrl(url)
	if err != nil {
		panic(err)
	}
	myzip.PrintZipFiles(reader)
	folder := fmt.Sprintf("%s", strconv.Itoa(int(time.Now().Unix())))
	_, err = myzip.UnZipAllFiles(url, reader, folder)
	if err != nil {
		panic(err)
	}
}

func main() {
	//u1 := "https://va-trialdist.azureedge.net/stella_trial.zip"
	u2 := "https://storage1.lathercraft.net/akabeesoft2/roleplayer2/ab2_roleplayer_tororoshimai_webtrial.zip"
	folder := fmt.Sprintf("%s", strconv.Itoa(int(time.Now().Unix())))
	props := &myzip.UnzipProps{
		Url:         u2,
		UnzipAll:    true,
		RangeStart:  1,
		RangeEnd:    10,
		Numbers:     nil,
		CharsetName: myzip.ShiftJIS,
		TargetPath:  folder,
	}
	reader, err := props.GetZipReader()
	if err != nil {
		panic(err)
	}
	err = props.InfoPrint(reader)
	if err != nil {
		panic(err)
	}
	// 内存过大
	_, results, err := props.Unzip(reader)
	if err != nil {
		panic(err)
	}
	myzip.ResultPrint(results)

}
