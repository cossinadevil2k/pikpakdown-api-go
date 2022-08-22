package main

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/mrxtryagin/pikpakdown-api-go/httpHandler"
	"github.com/mrxtryagin/pikpakdown-api-go/myzip"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"time"
)

const (
	EOCD_RECORD_SIZE = 22
)

func get_file_size(url string) int64 {
	client := httpHandler.NewClient()
	getResponse := client.Get(url, nil, httpHandler.WithProxy("http://127.0.0.1:7890"))
	//header := getResponse.Response.Header
	//fmt.Printf("response_headers: %v \n", header)
	return getResponse.Response.ContentLength

}

func getRangeBytes(url string, start, end int64) *[]byte {
	headers := http.Header{
		"range": {fmt.Sprintf("bytes=%d-%d", start, end)},
	}
	client := httpHandler.NewClient()
	response, err := client.Get(
		url,
		nil,
		httpHandler.WithHeader(headers),
		httpHandler.WithProxy("http://127.0.0.1:7890"),
	).GetResponse()
	if err != nil {
		panic(err)
	}
	return &response

}
func parseToInt(eocd64 *[]byte) (int64, int64) {
	input := *eocd64
	fmt.Printf("%v\n", input)
	cd_size := binary.LittleEndian.Uint32(input[12:16])
	cd_start := binary.LittleEndian.Uint32(input[16:20])
	return int64(cd_size), int64(cd_start)
}

func BytesCombine(pBytes ...[]byte) []byte {
	var buffer bytes.Buffer
	for _, pByte := range pBytes {
		buffer.Write(pByte)
	}
	return buffer.Bytes()
}

func extractPartZip(url string, start, end int) {
	total_size := get_file_size(url)
	eocdRecord := getRangeBytes(url, total_size-EOCD_RECORD_SIZE, total_size)
	cd_size, cd_start := parseToInt(eocdRecord)
	total_extra_size := cd_size + EOCD_RECORD_SIZE
	fmt.Printf("cd_start:%d,cd_size:%d,extra_size:%d\n,total_size:%d\n", cd_start, cd_size, total_extra_size, total_size)
	central_directory := getRangeBytes(url, cd_start, cd_start+cd_size-1)
	//start := getRangeBytes(url, 0, 1000)
	// 填充 todo: 这里全部都会读入内存 太大,考虑按照官方的思路重做
	// 这个buff是关键(其实就是directoryOffset 就是  cd_size + EOCD_RECORD_SIZE的偏移量 zip包是直接从这里解析的)
	//rs := io.NewSectionReader(r, 0, size)
	//	if _, err = rs.Seek(int64(end.directoryOffset), io.SeekStart); err != nil {
	//		return err
	//	}
	//	buf := bufio.NewReader(rs)
	//noneFull := make([]byte, total_size-total_extra_size)
	//archive, err := unarr.NewArchiveFromMemory(*central_directory)
	//if err != nil {
	//	panic(err)
	//}
	//list, err := archive.List()
	//if err != nil {
	//	panic(err)
	//}
	//fmt.Printf("result:%v", list)
	total_meta := BytesCombine(*central_directory, *eocdRecord)
	//print(total_meta)
	fmt.Println("file_tree:")
	metaReader := getZipList(&total_meta)
	//for _, file := range metaReader.File {
	//	fmt.Printf("%+v\n", *file)
	//}
	extractPartFilesTo(url, metaReader, start, end, fmt.Sprintf("%s", strconv.Itoa(int(time.Now().Unix()))))

	//// 拿到offset 还要通过反射
	//firstOffest := getPrivateValue(*first, "headerOffset").Int()
	//secondOffest := getPrivateValue(*second, "headerOffset").Int()
	//fmt.Printf("get %d to %d", firstOffest, secondOffest)
	//firstFileContent := getRangeBytes(url, firstOffest, secondOffest)
	//changeBytes(&total_meta, firstFileContent, firstOffest)

	//1.开一个新的
	//secondReader := getZipList(&total_meta)
	//fullFirst := secondReader.File[index]
	//unzipFile(fullFirst, "")
	//2. 直接修改 panic: reflect: reflect.Value.Set using value obtained using unexported field
	//v := reflect.ValueOf(bytes.NewReader(total_meta))
	//getPrivateValue(*first, "zipr").Set(v)
	//unzipFile(first, "")
	// 3. 自定 读两个偏移之间的是最有效果的 因为一个头部打开的过程涉及到 头 身体 和描述符这么多东西
	//first := metaReader.File[index]
	//second := metaReader.File[index+1]
	//firstHeadOffset := first.HeaderOffset
	//secodHeadOffset := second.HeaderOffset
	//// 两个head 偏移之间的内容就是第一个文件的全部内容 包括 头 + 身体 + 尾部数据描述符
	//firstFile := getRangeBytes(url, firstHeadOffset, secodHeadOffset)
	//firstFileReader := bytes.NewReader(*firstFile)
	//// 设置file 从 0开始读 给的字节也是请求的字节
	//first.Zip.R = firstFileReader
	//first.Zipr = firstFileReader
	//first.HeaderOffset = 0
	//unzipFile(first, "")

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
	if startIndex < endIndex {
		if endIndex < lastIndex {
			rangeFiles := files[startIndex : endIndex+1]
			for i := 0; i < len(rangeFiles)-1; i++ {
				go func(i int) {
					f := getNewFIle(url, rangeFiles[i], rangeFiles[i].HeaderOffset, rangeFiles[i+1].HeaderOffset)
					unzipFile(f, targetPath)
				}(i)
			}
		} else if endIndex == lastIndex {
			rangeFiles := files[startIndex:endIndex]
			for i := 0; i < len(rangeFiles); i++ {
				go func(i int) {
					var endOffset int64
					//如果是最后一个
					if i == len(rangeFiles)-1 {
						endOffset = int64(cdOffest)
					} else {
						endOffset = rangeFiles[i+1].HeaderOffset
					}
					f := getNewFIle(url, rangeFiles[i], rangeFiles[i].HeaderOffset, endOffset)
					unzipFile(f, targetPath)
				}(i)
			}
		}
	} else {
		rangeFiles := files[startIndex : endIndex+2]
		for i := 0; i < len(rangeFiles)-1; i++ {
			f := getNewFIle(url, rangeFiles[i], rangeFiles[i].HeaderOffset, rangeFiles[i+1].HeaderOffset)
			unzipFile(f, targetPath)
		}
	}
}
func getNewFIle(url string, f *myzip.File, firstOffset, secondOffset int64) *myzip.File {
	fileBytes := getRangeBytes(url, firstOffset, secondOffset)
	firstFileReader := bytes.NewReader(*fileBytes)
	// 设置file 从 0开始读 给的字节也是请求的字节
	f.Zip.R = firstFileReader
	f.Zipr = firstFileReader
	f.HeaderOffset = 0
	return f
}

func getPrivateValue(obj interface{}, field string) reflect.Value {
	v := reflect.ValueOf(obj)
	return v.FieldByName(field)
}

func getZipList(input *[]byte) *myzip.Reader {
	print(input)
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
	isExist, _ := PathExists(dst)
	if !isExist {
		err := os.Mkdir(dst, 0777)
		if err != nil {
			panic(err)
		}
	}
	destination := filepath.Join(dst, f.Name)
	destinationFile, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
	if err != nil {
		panic(err)
	}
	defer destinationFile.Close()
	targetFile, err := f.Open()
	if err != nil {
		panic(err)
	}
	defer targetFile.Close()

	n, err := io.Copy(destinationFile, targetFile)
	if err != nil {
		panic(err)
	}
	fmt.Printf("成功解压 %s ，共写入了 %d 个字符的数据\n", destination, n)

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
	u1 := "https://lgte-my.sharepoint.com/personal/mrx_lostknife_win/_layouts/15/download.aspx?UniqueId=662af895-f37c-4f17-95a8-a0f1f35f47df&Translate=false&tempauth=eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJhdWQiOiIwMDAwMDAwMy0wMDAwLTBmZjEtY2UwMC0wMDAwMDAwMDAwMDAvbGd0ZS1teS5zaGFyZXBvaW50LmNvbUA2Nzg0NTYxYS1lNTJkLTRlZGUtYmY4Yy1lZjBmNjk0ZGU5ZjIiLCJpc3MiOiIwMDAwMDAwMy0wMDAwLTBmZjEtY2UwMC0wMDAwMDAwMDAwMDAiLCJuYmYiOiIxNjYxMTU0ODMwIiwiZXhwIjoiMTY2MTE1ODQzMCIsImVuZHBvaW50dXJsIjoib2MzRkc2TmozSDh1ZEx4dnNZcHh1cS9EQnl2dHAvWnBsTXRJMmFLdGkyMD0iLCJlbmRwb2ludHVybExlbmd0aCI6IjE0NSIsImlzbG9vcGJhY2siOiJUcnVlIiwiY2lkIjoiWVdVMFl6TTJZV0l0TjJNM09TMDBabVk1TFdFMVpqTXRORFUyTlRFMVpEQXlOelkwIiwidmVyIjoiaGFzaGVkcHJvb2Z0b2tlbiIsInNpdGVpZCI6IlpUUmxZamt4TlRFdE0yUTVNeTAwWXpNM0xXRmhZVEV0TlRBeFpUTmtNMlpoTjJKayIsImFwcF9kaXNwbGF5bmFtZSI6ImNsb3VkcmV2ZSIsImdpdmVuX25hbWUiOiJyeCIsImZhbWlseV9uYW1lIjoibSIsImFwcGlkIjoiNjQ0NWJkNWItZjI1OS00YmY2LTgxMTItZGFjODA2N2RmZjM5IiwidGlkIjoiNjc4NDU2MWEtZTUyZC00ZWRlLWJmOGMtZWYwZjY5NGRlOWYyIiwidXBuIjoibXJ4QGxvc3RrbmlmZS53aW4iLCJwdWlkIjoiMTAwMzIwMDBBNzQ4Q0IzOCIsImNhY2hla2V5IjoiMGguZnxtZW1iZXJzaGlwfDEwMDMyMDAwYTc0OGNiMzhAbGl2ZS5jb20iLCJzY3AiOiJhbGxmaWxlcy53cml0ZSIsInR0IjoiMiIsInVzZVBlcnNpc3RlbnRDb29raWUiOm51bGwsImlwYWRkciI6IjIwLjE5MC4xNDQuMTcwIn0.V1pEZnlQRXNWWVBPYmJsdE1FZFpYRGtyMU5ueEV6VlZReUpzR0huVlJLST0&ApiVersion=2.0"
	//extractPartZip(u1)
	//fmt.Println(BytesCombine([]byte{1, 2, 3}, []byte{4, 5, 6}))
	//u2 := "https://lgte-my.sharepoint.com/personal/mrx_lostknife_win/_layouts/15/download.aspx?UniqueId=26558f00-c9e9-484e-93e7-45d0d0b35db7&Translate=false&tempauth=eyJ0eXAiOiJKV1QiLCJhbGciOiJub25lIn0.eyJhdWQiOiIwMDAwMDAwMy0wMDAwLTBmZjEtY2UwMC0wMDAwMDAwMDAwMDAvbGd0ZS1teS5zaGFyZXBvaW50LmNvbUA2Nzg0NTYxYS1lNTJkLTRlZGUtYmY4Yy1lZjBmNjk0ZGU5ZjIiLCJpc3MiOiIwMDAwMDAwMy0wMDAwLTBmZjEtY2UwMC0wMDAwMDAwMDAwMDAiLCJuYmYiOiIxNjYxMDAyNjQ1IiwiZXhwIjoiMTY2MTAwNjI0NSIsImVuZHBvaW50dXJsIjoiUm1YMFlFM3JzWFZnVU02UnhMQTgxTU1vb1RlK3huWlhHODQ1QnRWWGFibz0iLCJlbmRwb2ludHVybExlbmd0aCI6IjE0NSIsImlzbG9vcGJhY2siOiJUcnVlIiwiY2lkIjoiWW1WbFlUYzBZVEF0T0RobE5pMDBNMkkzTFRsak16RXRaVEEzT0dVNVl6QTBaV0l3IiwidmVyIjoiaGFzaGVkcHJvb2Z0b2tlbiIsInNpdGVpZCI6IlpUUmxZamt4TlRFdE0yUTVNeTAwWXpNM0xXRmhZVEV0TlRBeFpUTmtNMlpoTjJKayIsImFwcF9kaXNwbGF5bmFtZSI6ImNsb3VkcmV2ZSIsImdpdmVuX25hbWUiOiJyeCIsImZhbWlseV9uYW1lIjoibSIsImFwcGlkIjoiNjQ0NWJkNWItZjI1OS00YmY2LTgxMTItZGFjODA2N2RmZjM5IiwidGlkIjoiNjc4NDU2MWEtZTUyZC00ZWRlLWJmOGMtZWYwZjY5NGRlOWYyIiwidXBuIjoibXJ4QGxvc3RrbmlmZS53aW4iLCJwdWlkIjoiMTAwMzIwMDBBNzQ4Q0IzOCIsImNhY2hla2V5IjoiMGguZnxtZW1iZXJzaGlwfDEwMDMyMDAwYTc0OGNiMzhAbGl2ZS5jb20iLCJzY3AiOiJhbGxmaWxlcy53cml0ZSIsInR0IjoiMiIsInVzZVBlcnNpc3RlbnRDb29raWUiOm51bGwsImlwYWRkciI6IjIwLjE5MC4xNDQuMTcwIn0.amlyRHJRbjFhcFp4cVRxa1ZXR0NadlcwT1Jrem5CRWZzWHUvUlkxZkkwRT0&ApiVersion=2.0"
	extractPartZip(u1, 1, 200)
	for {

	}

}
