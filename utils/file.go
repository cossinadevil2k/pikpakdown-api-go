package utils

import (
	"bufio"
	"io"
	"io/ioutil"
	"os"
)

const (
	ChunkSize = 1024
	// 新建 + 追加写
	FileAppendFlag = os.O_WRONLY | os.O_CREATE | os.O_APPEND
	// 新建 + 覆盖写
	FileOverWriteFlag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
)

//ReadWholeFile 阅读整个文件
func ReadWholeFile(filePath string) (*[]byte, error) {
	content, err := ioutil.ReadFile(filePath)
	return &content, err
}

//ReadFileByChunk 按照chunk 阅读文件
func ReadFileByChunk(filePath string) (*[]byte, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	chunks := make([]byte, 0)
	buf := make([]byte, ChunkSize)
	for {
		//也可以用bufio.Read
		n, err := f.Read(buf)
		if err != nil {
			//是否到达文件末尾
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		chunks = append(chunks, buf[:n]...)
	}
	err = f.Close()
	if err != nil {
		return nil, err
	}
	return &chunks, nil
}

//ReadFileByLine 按照行来读取文件(字符串文本)
func ReadFileByLine(filePath string) (*[]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	buf := bufio.NewReader(file)
	var results []string
	for {
		//读成字符串 也可以用 buf.ReadLine()
		line, err := buf.ReadString('\n')
		results = append(results, line)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
	}
	err = file.Close()
	if err != nil {
		return nil, err
	}
	return &results, nil
}

//WriteFileCustom 写入content 使用 自定义的flag
func WriteFileCustom(content *[]byte, filePath string, flag int) error {
	file, err := os.OpenFile(filePath, flag, 0666)
	if err != nil {
		return err
	}
	//也可以用File.Write
	writer := bufio.NewWriter(file)
	_, err = writer.Write(*content)
	if err != nil {
		return err
	}
	err = writer.Flush()
	if err != nil {
		return err

	}
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}

//WriteFileOverWrite 直接覆盖写 不存在创建,存在清空写入
func WriteFileOverWrite(content *[]byte, filePath string) error {
	return os.WriteFile(filePath, *content, 0666)
}
