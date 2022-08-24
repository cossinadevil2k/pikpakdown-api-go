package main

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/imroc/req/v3"
	"github.com/mrxtryagin/pikpakdown-api-go/httpHandler"
	"github.com/mrxtryagin/pikpakdown-api-go/myzip"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"
)

const (
	EOCD_RECORD_SIZE        = 22
	ZIP64_EOCD_RECORD_SIZE  = 56
	ZIP64_EOCD_LOCATOR_SIZE = 20

	MAX_STANDARD_ZIP_SIZE = 4_294_967_295
)

func get_file_size(url string) int64 {
	client := httpHandler.NewClient()
	getResponse := client.Get(url, nil, httpHandler.WithProxy("http://127.0.0.1:7890"))
	if getResponse.Err != nil {
		panic(getResponse.Err)
	}
	//header := getResponse.Response.Header
	//fmt.Printf("response_headers: %v \n", header)
	length := getResponse.Response.ContentLength
	if length == -1 {
		panic("url is invaild")
	}
	return length

}

func getRangeBytes(url string, start, end int64) *[]byte {
	//headers := http.Header{
	//	"range": {fmt.Sprintf("bytes=%d-%d", start, end)}, // 左闭右闭
	//}
	//client := httpHandler.NewClient()
	//response, err := client.Get(
	//	url,
	//	nil,
	//	httpHandler.WithHeader(headers),
	//	httpHandler.WithProxy("http://127.0.0.1:7890"),
	//	httpHandler.WithRetry(3, func(repose *http.Response, otherError error) bool {
	//		if otherError != nil {
	//			return true
	//		}
	//		return repose.ContentLength == -1
	//	}),
	//).GetResponse()
	//if err != nil {
	//	panic(err)
	//}
	headers := map[string]string{
		"range": fmt.Sprintf("bytes=%d-%d", start, end),
	}
	client := req.C().
		SetCommonHeaders(headers).
		SetTimeout(60 * time.Second).
		SetProxyURL("http://127.0.0.1:7890").
		SetCommonRetryFixedInterval(2 * time.Second).
		SetCommonRetryCount(5).
		SetCommonRetryCondition(func(resp *req.Response, err error) bool {
			return err != nil || resp.ContentLength == -1
		}).
		EnableDebugLog()

	response, err := client.R().
		Get(url)
	if err != nil {
		panic(err)
	}
	result := response.Bytes()
	fmt.Printf("response_len:%d start:%d end:%d headers:%+v \n", len(result), start, end, headers)
	return &result

}

/**
Offset | Bytes | Description
12     | 4     | Size of central directory
16     | 4     | Offset of start of CD, relative to start of archive
*/
func parseToInt(eocd *[]byte) (int64, int64) {
	input := *eocd
	fmt.Printf("%v\n", input)
	cd_size := binary.LittleEndian.Uint32(input[12:16])
	cd_start := binary.LittleEndian.Uint32(input[16:20])
	return int64(cd_size), int64(cd_start)
}

/**
Offset | Bytes | Description
40     | 8     | Size of central directory
48     | 8     | Offset of start of CD, relative to start of archive
*/
func parseToInt64(eocd64 *[]byte) (int64, int64) {
	input := *eocd64
	fmt.Printf("%v\n", input)
	// 读8位
	cd_size := binary.LittleEndian.Uint64(input[40:48])
	cd_start := binary.LittleEndian.Uint64(input[48:56])
	return int64(cd_size), int64(cd_start)
}

func BytesCombine(pBytes ...[]byte) []byte {
	var buffer bytes.Buffer
	for _, pByte := range pBytes {
		buffer.Write(pByte)
	}
	return buffer.Bytes()
}

func getZipReader(url string) *myzip.Reader {
	total_size := get_file_size(url)
	eocdRecord := getRangeBytes(url, total_size-EOCD_RECORD_SIZE, total_size)
	// 如果是普通zip
	/*
	   totalFetch: central_directory + eocd_record
	*/
	if total_size <= MAX_STANDARD_ZIP_SIZE {
		cd_size, cd_start := parseToInt(eocdRecord)
		total_extra_size := cd_size + EOCD_RECORD_SIZE
		central_directory := getRangeBytes(url, cd_start, cd_start+cd_size-1)
		fmt.Printf("cd_start:%d,cd_size:%d,extra_size:%d\n,total_size:%d\n", cd_start, cd_size, total_extra_size, total_size)
		total_meta := BytesCombine(*central_directory, *eocdRecord)
		////print(total_meta)
		//fmt.Println("file_tree:")
		//metaReader := getZipList(&total_meta)
		//for index, file := range metaReader.File {
		//	fmt.Printf("%d: %s\n", index, file.Name)
		//}

		args := &myzip.InitArgs{
			IsZip64:              false,
			TotalSize:            total_size,
			EOCDSize:             EOCD_RECORD_SIZE,
			CDSize:               cd_size,
			Zip64EocdRecordSize:  ZIP64_EOCD_RECORD_SIZE,
			Zip64EocdLocatorSize: ZIP64_EOCD_LOCATOR_SIZE,
			ExtraSize:            int64(len(total_meta)),
		}
		reader, err := myzip.NewReaderFromArgs(bytes.NewReader(total_meta), args)
		if err != nil {
			panic(err)
		}
		return reader
	} else {
		/*
				As already mentioned,
				the ZIP and ZIP64 structure are a bit different.
				For the latter one,
				the algorithm looks also the same.
				The only difference is that you need to
				fetch an extra ZIP64 EOCD record and a ZIP64 EOCD locator.
				Then the four bytes blocks (CD+EOCD64 record+EOCD64 locator+EOCD) can be read and open as a ZIP file.
			    totalFetch: central_directory + zip64_eocd_record + zip64_eocd_locator + eocd_record
		*/
		//如果是zip64 超过4G的zip,还需要请求剩余的eocd
		zip64_eocd_record_start := total_size - (EOCD_RECORD_SIZE + ZIP64_EOCD_LOCATOR_SIZE + ZIP64_EOCD_RECORD_SIZE)
		zip64_eocd_record := getRangeBytes(url,
			zip64_eocd_record_start,
			zip64_eocd_record_start+ZIP64_EOCD_RECORD_SIZE-1,
		)
		zip64_eocd_locator_start := total_size - (EOCD_RECORD_SIZE + ZIP64_EOCD_LOCATOR_SIZE)
		zip64_eocd_locator := getRangeBytes(url,
			zip64_eocd_locator_start,
			zip64_eocd_locator_start+ZIP64_EOCD_LOCATOR_SIZE-1,
		)

		cd_size, cd_start := parseToInt64(zip64_eocd_record)
		central_directory := getRangeBytes(url, cd_start, cd_start+cd_size-1)
		total_meta := BytesCombine(*central_directory, *zip64_eocd_record, *zip64_eocd_locator, *eocdRecord)
		fmt.Printf("cd_start:%d,cd_size:%d,zip64_eocd_record_start:%d\n,zip64_eocd_locator_start:%d\n,total_size:%d\n", cd_start, cd_size, zip64_eocd_record_start, zip64_eocd_locator_start, total_size)
		args := &myzip.InitArgs{
			IsZip64:              true,
			TotalSize:            total_size,
			EOCDSize:             EOCD_RECORD_SIZE,
			CDSize:               cd_size,
			Zip64EocdRecordSize:  ZIP64_EOCD_RECORD_SIZE,
			Zip64EocdLocatorSize: ZIP64_EOCD_LOCATOR_SIZE,
			ExtraSize:            int64(len(total_meta)),
		}

		reader, err := myzip.NewReaderFromArgs(bytes.NewReader(total_meta), args)
		if err != nil {
			panic(err)
		}
		return reader
	}

}
func changeBytes(input, fromBytes *[]byte, start int64) {
	for _, b := range *fromBytes {
		(*input)[start] = b
		start++
	}
}

func extractPartFilesTo(url string, reader *myzip.Reader, startNo, endNo int, targetPath string) {
	files := reader.File
	// 找到对应的下标
	startIndex := startNo - 1
	endIndex := endNo - 1
	//判断下标是否越界
	lastIndex := len(files) - 1
	if startIndex < 0 || endIndex < 0 {
		panic(errors.New("start,end must be >= 0 "))
	}
	if endIndex < startIndex {
		panic(errors.New("start > end"))
	}
	if startIndex > lastIndex || endIndex > lastIndex {
		panic(errors.New("index out of range"))
	}
	/*
		寻找满足条件的文件,算法如下:
		1. startIndex < endIndex < lastIndex:
		 抽取 start 到 end+1个的bytes进行解压
		2. startIndex < endIndex = lastIndex:
		 因为end 就是最后一个 没有下一个偏移了 所以用eocd的值进行,也就是抽取 start 到 ecod.DirectoryOffset
		3.  startIndex = endIndex:
		 start 与 end 是同一个 start 同一
	*/
	cdOffest := reader.EOCD.DirectoryOffset
	token := make(chan int, 200)
	if startIndex < endIndex {
		if endIndex < lastIndex {
			rangeFiles := files[startIndex : endIndex+1]
			for i := 0; i < len(rangeFiles)-1; i++ {
				go func(i int) {
					token <- 1
					//fmt.Printf("第%d次\n", i)
					f := getNewFIle(url, rangeFiles[i], rangeFiles[i].HeaderOffset, rangeFiles[i+1].HeaderOffset-1)
					unzipFile(f, targetPath)
					<-token
				}(i)
				//fmt.Printf("第%d次,start:%d end:%d \n", i, rangeFiles[i].HeaderOffset, rangeFiles[i+1].HeaderOffset-1)
				//f := getNewFIle(url, rangeFiles[i], rangeFiles[i].HeaderOffset, rangeFiles[i+1].HeaderOffset-1)
				//unzipFile(f, targetPath)
			}
		} else if endIndex == lastIndex {
			rangeFiles := files[startIndex : endIndex+1]
			for i := 0; i < len(rangeFiles); i++ {
				go func(i int) {
					token <- 1
					var endOffset int64
					//如果是最后一个
					if i == len(rangeFiles)-1 {
						endOffset = int64(cdOffest)
					} else {
						endOffset = rangeFiles[i+1].HeaderOffset - 1
					}
					f := getNewFIle(url, rangeFiles[i], rangeFiles[i].HeaderOffset, endOffset)
					unzipFile(f, targetPath)
					<-token
				}(i)
			}
		}
	} else {
		rangeFiles := files[startIndex : endIndex+2]
		for i := 0; i < len(rangeFiles)-1; i++ {
			f := getNewFIle(url, rangeFiles[i], rangeFiles[i].HeaderOffset, rangeFiles[i+1].HeaderOffset-1)
			unzipFile(f, targetPath)
		}
	}
}

func unzipFiles(url string, reader *myzip.Reader, targetPath string, nos ...int) {
	if len(nos) == 0 {
		panic("nos is empty!")
	}
	lastIndex := len(reader.File) - 1
	token := make(chan int, 20)
	// ecod的偏移
	eocdOffest := reader.EOCD.DirectoryOffset
	for _, no := range nos {
		// 从no获得下标
		noIndex := no - 1
		if noIndex < 0 || noIndex > lastIndex {
			panic("noIndex < 0 or noIndex > lastIndex,noIndex is invalid")
		}
		go func(noIndex int) {
			// 令牌桶限制频率
			token <- 1
			nowFile := reader.File[noIndex]
			start := nowFile.HeaderOffset
			var end int64
			// 区分小于和等于的情况
			if noIndex < lastIndex {
				end = reader.File[noIndex+1].HeaderOffset - 1
			} else if noIndex == lastIndex {
				end = int64(eocdOffest - 1)
			}
			//如果是小于的,那直接取就行了 注意 左闭右闭 所以说是 start = 本文件的开始偏移 end = 下一个文件的开始偏移-1的区间
			f := getNewFIle(url, nowFile, start, end)
			unzipFile(f, targetPath)
			<-token
		}(noIndex)
	}
}

func getNewFIle(url string, f *myzip.File, firstOffset, secondOffset int64) *myzip.File {
	fileBytes := getRangeBytes(url, firstOffset, secondOffset)
	firstFileReader := bytes.NewReader(*fileBytes)
	// 设置file 从 0开始读 给的字节也是请求的字节
	// 拷贝一个f
	newf := &myzip.File{
		FileHeader:   f.FileHeader,
		Zip:          f.Zip,
		Zipr:         f.Zipr,
		Zip64:        f.Zip64,
		DescErr:      f.DescErr,
		HeaderOffset: f.HeaderOffset,
	}
	newf.Zip.R = firstFileReader
	newf.Zipr = firstFileReader
	newf.HeaderOffset = 0
	return newf
}

func getPrivateValue(obj interface{}, field string) reflect.Value {
	v := reflect.ValueOf(obj)
	return v.FieldByName(field)
}

func getZipReaderFromBytes(input *[]byte) *myzip.Reader {
	reader, err := myzip.NewReader(bytes.NewReader(*input), int64(len(*input)))
	if err != nil {
		panic(err)
	}
	return reader
}

func PathExists(path string) (bool, error) {
	/*
	  判断文件或文件夹是否存在
	    如果返回的错误为nil,说明文件或文件夹存在
	    如果返回的错误类型使用os.IsNotExist()判断为true,说明文件或文件夹不存在
	    如果返回的错误为其它类型,则不确定是否在存在
	*/
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func unzipFile(f *myzip.File, dst string) {
	//isExist, _ := PathExists(dst)
	//if !isExist {
	//	err := os.Mkdir(dst, os.ModePerm)
	//	if err != nil {
	//		panic(err)
	//	}
	//}
	fmt.Printf("%+v", *f)
	var decodeName string
	if f.Flags == 0 {
		//如果标致位是0  则是默认的本地编码   默认为gbk
		i := bytes.NewReader([]byte(f.Name))
		decoder := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
		content, _ := ioutil.ReadAll(decoder)
		decodeName = string(content)
	} else {
		//如果标志为是 1 << 11也就是 2048  则是utf-8编码
		decodeName = f.Name
	}

	destination := filepath.Join(dst, decodeName)
	if f.FileInfo().IsDir() {
		//如果这个文件是文件夹 直接创建文件夹即可
		fmt.Printf("成功创建文件夹%s", destination)
		os.MkdirAll(destination, os.ModePerm)
	} else {
		//如果是文件夹套的文件
		if err := os.MkdirAll(filepath.Dir(destination), os.ModePerm); err != nil {
			panic(err)
		}
		destinationFile, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			panic(err)
		}
		defer destinationFile.Close()
		sourceFile, err := f.Open()
		if err != nil {
			fmt.Printf("file:%+v\n is error:%s\n", f, err.Error())
			panic(err)
		}
		defer sourceFile.Close()

		n, err := io.Copy(destinationFile, sourceFile)
		if err != nil {
			panic(err)
		}
		fmt.Printf("成功解压 %s ，共写入了 %d 个字符的数据\n", destination, n)
	}

}

func readAll(url string) {
	client := httpHandler.NewClient()
	getResponse := client.Get(
		url,
		nil,
		httpHandler.WithProxy("http://127.0.0.1:7890"),
	)
	//header := getResponse.Response.Header
	//fmt.Printf("response_headers: %v \n", header)
	response, err := getResponse.GetResponse()
	if err != nil {
		panic(err)
	}
	reader, err := zip.NewReader(bytes.NewReader(response), int64(len(response)))
	if err != nil {
		panic(err)
	}
	for _, f := range reader.File {
		fmt.Printf("%+v\n", *f)
	}

}

func main() {
	//u1 := "https://lgte-my.sharepoint.com/personal/mrx_lostknife_win/_layouts/15/download.aspx?UniqueId=662af895-f37c-4f17-95a8-a0f1f35f47df&Translate=false&tempauth=eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJhdWQiOiIwMDAwMDAwMy0wMDAwLTBmZjEtY2UwMC0wMDAwMDAwMDAwMDAvbGd0ZS1teS5zaGFyZXBvaW50LmNvbUA2Nzg0NTYxYS1lNTJkLTRlZGUtYmY4Yy1lZjBmNjk0ZGU5ZjIiLCJpc3MiOiIwMDAwMDAwMy0wMDAwLTBmZjEtY2UwMC0wMDAwMDAwMDAwMDAiLCJuYmYiOiIxNjYxMTY4MjY4IiwiZXhwIjoiMTY2MTE3MTg2OCIsImVuZHBvaW50dXJsIjoib2MzRkc2TmozSDh1ZEx4dnNZcHh1cS9EQnl2dHAvWnBsTXRJMmFLdGkyMD0iLCJlbmRwb2ludHVybExlbmd0aCI6IjE0NSIsImlzbG9vcGJhY2siOiJUcnVlIiwiY2lkIjoiTVdSbU1HRTJPVEV0TkRBMU1DMDBOamt6TFRrNU1UUXRPRFppWmpFMU1qWTNaRE0yIiwidmVyIjoiaGFzaGVkcHJvb2Z0b2tlbiIsInNpdGVpZCI6IlpUUmxZamt4TlRFdE0yUTVNeTAwWXpNM0xXRmhZVEV0TlRBeFpUTmtNMlpoTjJKayIsImFwcF9kaXNwbGF5bmFtZSI6ImNsb3VkcmV2ZSIsImdpdmVuX25hbWUiOiJyeCIsImZhbWlseV9uYW1lIjoibSIsImFwcGlkIjoiNjQ0NWJkNWItZjI1OS00YmY2LTgxMTItZGFjODA2N2RmZjM5IiwidGlkIjoiNjc4NDU2MWEtZTUyZC00ZWRlLWJmOGMtZWYwZjY5NGRlOWYyIiwidXBuIjoibXJ4QGxvc3RrbmlmZS53aW4iLCJwdWlkIjoiMTAwMzIwMDBBNzQ4Q0IzOCIsImNhY2hla2V5IjoiMGguZnxtZW1iZXJzaGlwfDEwMDMyMDAwYTc0OGNiMzhAbGl2ZS5jb20iLCJzY3AiOiJhbGxmaWxlcy53cml0ZSIsInR0IjoiMiIsInVzZVBlcnNpc3RlbnRDb29raWUiOm51bGwsImlwYWRkciI6IjIwLjE5MC4xNDQuMTcyIn0.Y1B3RXNneE52K05OS3FSbzM0eW8yMzFUSy9GbWYyUjhCZHlKMnQ0a0x3VT0&ApiVersion=2.0"
	//u1 := "https://lgte-my.sharepoint.com/personal/mrx_lostknife_win/_layouts/15/download.aspx?UniqueId=662af895-f37c-4f17-95a8-a0f1f35f47df&Translate=false&tempauth=eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJhdWQiOiIwMDAwMDAwMy0wMDAwLTBmZjEtY2UwMC0wMDAwMDAwMDAwMDAvbGd0ZS1teS5zaGFyZXBvaW50LmNvbUA2Nzg0NTYxYS1lNTJkLTRlZGUtYmY4Yy1lZjBmNjk0ZGU5ZjIiLCJpc3MiOiIwMDAwMDAwMy0wMDAwLTBmZjEtY2UwMC0wMDAwMDAwMDAwMDAiLCJuYmYiOiIxNjYxMTcyMzE2IiwiZXhwIjoiMTY2MTE3NTkxNiIsImVuZHBvaW50dXJsIjoib2MzRkc2TmozSDh1ZEx4dnNZcHh1cS9EQnl2dHAvWnBsTXRJMmFLdGkyMD0iLCJlbmRwb2ludHVybExlbmd0aCI6IjE0NSIsImlzbG9vcGJhY2siOiJUcnVlIiwiY2lkIjoiTm1ReU16UXdPRE10TWpReE5DMDBNRGN3TFdJM01XSXRObUZoWWpjMk16ZGxNV0UyIiwidmVyIjoiaGFzaGVkcHJvb2Z0b2tlbiIsInNpdGVpZCI6IlpUUmxZamt4TlRFdE0yUTVNeTAwWXpNM0xXRmhZVEV0TlRBeFpUTmtNMlpoTjJKayIsImFwcF9kaXNwbGF5bmFtZSI6ImNsb3VkcmV2ZSIsImdpdmVuX25hbWUiOiJyeCIsImZhbWlseV9uYW1lIjoibSIsImFwcGlkIjoiNjQ0NWJkNWItZjI1OS00YmY2LTgxMTItZGFjODA2N2RmZjM5IiwidGlkIjoiNjc4NDU2MWEtZTUyZC00ZWRlLWJmOGMtZWYwZjY5NGRlOWYyIiwidXBuIjoibXJ4QGxvc3RrbmlmZS53aW4iLCJwdWlkIjoiMTAwMzIwMDBBNzQ4Q0IzOCIsImNhY2hla2V5IjoiMGguZnxtZW1iZXJzaGlwfDEwMDMyMDAwYTc0OGNiMzhAbGl2ZS5jb20iLCJzY3AiOiJhbGxmaWxlcy53cml0ZSIsInR0IjoiMiIsInVzZVBlcnNpc3RlbnRDb29raWUiOm51bGwsImlwYWRkciI6IjIwLjE5MC4xNDQuMTY5In0.Mi9jYnRobzN4K3NkT2lhc2Frb29waTdaTHBlbm1ibElFbFBDckN6cmpoOD0&ApiVersion=2.0"
	//extractPartZip(u1)
	//fmt.Println(BytesCombine([]byte{1, 2, 3}, []byte{4, 5, 6}))
	//u2 := "https://lgte-my.sharepoint.com/personal/mrx_lostknife_win/_layouts/15/download.aspx?UniqueId=26558f00-c9e9-484e-93e7-45d0d0b35db7&Translate=false&tempauth=eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJhdWQiOiIwMDAwMDAwMy0wMDAwLTBmZjEtY2UwMC0wMDAwMDAwMDAwMDAvbGd0ZS1teS5zaGFyZXBvaW50LmNvbUA2Nzg0NTYxYS1lNTJkLTRlZGUtYmY4Yy1lZjBmNjk0ZGU5ZjIiLCJpc3MiOiIwMDAwMDAwMy0wMDAwLTBmZjEtY2UwMC0wMDAwMDAwMDAwMDAiLCJuYmYiOiIxNjYxMDAyNjQ1IiwiZXhwIjoiMTY2MTAwNjI0NSIsImVuZHBvaW50dXJsIjoiUm1YMFlFM3JzWFZnVU02UnhMQTgxTU1vb1RlK3huWlhHODQ1QnRWWGFibz0iLCJlbmRwb2ludHVybExlbmd0aCI6IjE0NSIsImlzbG9vcGJhY2siOiJUcnVlIiwiY2lkIjoiWW1WbFlUYzBZVEF0T0RobE5pMDBNMkkzTFRsak16RXRaVEEzT0dVNVl6QTBaV0l3IiwidmVyIjoiaGFzaGVkcHJvb2Z0b2tlbiIsInNpdGVpZCI6IlpUUmxZamt4TlRFdE0yUTVNeTAwWXpNM0xXRmhZVEV0TlRBeFpUTmtNMlpoTjJKayIsImFwcF9kaXNwbGF5bmFtZSI6ImNsb3VkcmV2ZSIsImdpdmVuX25hbWUiOiJyeCIsImZhbWlseV9uYW1lIjoibSIsImFwcGlkIjoiNjQ0NWJkNWItZjI1OS00YmY2LTgxMTItZGFjODA2N2RmZjM5IiwidGlkIjoiNjc4NDU2MWEtZTUyZC00ZWRlLWJmOGMtZWYwZjY5NGRlOWYyIiwidXBuIjoibXJ4QGxvc3RrbmlmZS53aW4iLCJwdWlkIjoiMTAwMzIwMDBBNzQ4Q0IzOCIsImNhY2hla2V5IjoiMGguZnxtZW1iZXJzaGlwfDEwMDMyMDAwYTc0OGNiMzhAbGl2ZS5jb20iLCJzY3AiOiJhbGxmaWxlcy53cml0ZSIsInR0IjoiMiIsInVzZVBlcnNpc3RlbnRDb29raWUiOm51bGwsImlwYWRkciI6IjIwLjE5MC4xNDQuMTcwIn0.amlyRHJRbjFhcFp4cVRxa1ZXR0NadlcwT1Jrem5CRWZzWHUvUlkxZkkwRT0&ApiVersion=2.0"
	u1 := "https://ytplbi-my.sharepoint.com/personal/cp60007_ytplbi_onmicrosoft_com/_layouts/15/download.aspx?UniqueId=8c833a21-2051-422f-a54c-f72bd637d8ce&Translate=false&tempauth=eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJhdWQiOiIwMDAwMDAwMy0wMDAwLTBmZjEtY2UwMC0wMDAwMDAwMDAwMDAveXRwbGJpLW15LnNoYXJlcG9pbnQuY29tQGNkODJlZjE4LThlYWEtNGIzNy04NjhlLTA4YjFlNzZlNjkwNyIsImlzcyI6IjAwMDAwMDAzLTAwMDAtMGZmMS1jZTAwLTAwMDAwMDAwMDAwMCIsIm5iZiI6IjE2NjEyNDYwODkiLCJleHAiOiIxNjYxMjQ5Njg5IiwiZW5kcG9pbnR1cmwiOiJad25wUmRPbGk4eXVrU2RER0hMazhWdmdnWTNnT3lxNVl2bEUzaGwzK2swPSIsImVuZHBvaW50dXJsTGVuZ3RoIjoiMTYwIiwiaXNsb29wYmFjayI6IlRydWUiLCJjaWQiOiJOVE5pWWpsa1l6Y3RZMlU1TUMwMFltRmhMV0ZoT0dZdFlXSmtNalF5T0dVMk5tSmwiLCJ2ZXIiOiJoYXNoZWRwcm9vZnRva2VuIiwic2l0ZWlkIjoiTUdFeU5UQTRORGt0WTJRME15MDBZVFU1TFdJeFl6Z3RZek00TmpCbE5qazJNakl4IiwiYXBwX2Rpc3BsYXluYW1lIjoiY2xvdWRyZXZlIiwic2lnbmluX3N0YXRlIjoiW1wia21zaVwiXSIsImFwcGlkIjoiZGEyZDE3ZDUtMWZlNy00MTc3LWE2NjYtNjg5Njk5ODI0NzgzIiwidGlkIjoiY2Q4MmVmMTgtOGVhYS00YjM3LTg2OGUtMDhiMWU3NmU2OTA3IiwidXBuIjoiY3A2MDAwN0B5dHBsYmkub25taWNyb3NvZnQuY29tIiwicHVpZCI6IjEwMDMyMDAxNzRGRDZGRkQiLCJjYWNoZWtleSI6IjBoLmZ8bWVtYmVyc2hpcHwxMDAzMjAwMTc0ZmQ2ZmZkQGxpdmUuY29tIiwic2NwIjoiYWxsZmlsZXMud3JpdGUiLCJ0dCI6IjIiLCJ1c2VQZXJzaXN0ZW50Q29va2llIjpudWxsLCJpcGFkZHIiOiIyMC4xOTAuMTQ0LjE2OSJ9.aUZodEcvU2hVQjN4QVc4Z2JuU3RWUUZtSlZLNGYyOE4vMFNIYkp2Zkk0QT0&ApiVersion=2.0"
	zipReader := getZipReader(u1)
	//fmt.Println("file_tree:")
	for index, file := range zipReader.File {
		fmt.Printf("%d: %s\n", index, file.Name)
	}
	//marshal, err := json.Marshal(zipReader)
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("reader:%+v\n", string(marshal))
	folder := fmt.Sprintf("%s", strconv.Itoa(int(time.Now().Unix())))
	unzipFiles(u1, zipReader, folder, 1, 3, 5, 6, 7, 8, 158)
	for {

	}

}
