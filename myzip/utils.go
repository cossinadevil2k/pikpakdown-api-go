package myzip

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/imroc/req/v3"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// 约定量
const (
	EocdRecordSize       = 22 // ecod大小
	Zip64EocdRecordSize  = 56 // zip64EocdRecord大小
	Zip64EocdLocatorSize = 20 // zip64EcodLocator大小

	MaxStandardZipSize = 4_294_967_295 //标准压缩包的大小

	TimeOut = 60 * time.Second

	RetryFixedInterval = 2 * time.Second

	RetryCount = 5

	// 开发环境使用
	ProxyUrl = "http://127.0.0.1:7890"
)

var (
	client *req.Client
)

var (
	UrlIsInvalid         = errors.New("url is invalid")
	NoUnZipNumbersChoice = errors.New("no unzip numbers choice")
)

type UnzipProps struct {
	Url string
	// 优先级  UnzipAll > Range > numbers
	UnzipAll    bool
	RangeStart  int
	RangeEnd    int
	Numbers     []int
	CharsetName string
	TargetPath  string
}

type ResultProps struct {
	IsSuccess         bool
	SuccessHandleByte int64
	Err               error
	File              *File
	FileRangeStart    int64
	FileRangeEnd      int64
	CDStart           uint64
	ZipReader         *Reader
	FilePath          string
	FileIndex         int //文件在压缩文件目录的位置
}

//RetryConditionForContentLength 对于ContentLength的重试条件
func RetryConditionForContentLength(resp *req.Response, err error) bool {
	return err != nil || resp.ContentLength == -1
}

func init() {
	// 初始化client
	client = req.C().
		SetTimeout(TimeOut).
		SetCommonRetryFixedInterval(RetryFixedInterval).
		SetCommonRetryCount(RetryCount).
		AddCommonRetryCondition(RetryConditionForContentLength).
		SetProxyURL(ProxyUrl).
		DisableAutoReadResponse(). // 禁用自动读取
		EnableDebugLog()
}

func getFileSize(url string) (int64, error) {
	response, err := client.R().
		Get(url)
	if err != nil {
		return 0, err
	}
	if response.ContentLength == -1 {
		return 0, UrlIsInvalid
	}
	return response.ContentLength, nil

}

func getRangeBytes(url string, start, end int64) (*[]byte, error) {
	headers := map[string]string{
		"range": fmt.Sprintf("bytes=%d-%d", start, end),
	}
	response, err := client.R().
		SetHeaders(headers).
		SetRetryCount(10).
		SetRetryInterval(func(resp *req.Response, attempt int) time.Duration {
			// Sleep seconds from "Retry-After" response header if it is present and correct.
			// https://www.w3.org/Protocols/rfc2616/rfc2616-sec14.html
			if resp.Response != nil {
				if ra := resp.Header.Get("Retry-After"); ra != "" {
					after, err := strconv.Atoi(ra)
					if err == nil {
						return time.Duration(after) * time.Second
					}
				}
			}
			return 2 * time.Second // Otherwise, sleep 2 seconds
		}).
		Get(url)
	if err != nil {
		return nil, err
	}
	// 获取响应体
	toBytes, err := response.ToBytes()
	if err != nil {
		return nil, err
	}

	return &toBytes, nil

}

/**
combin bytes
*/
func BytesCombine(pBytes ...[]byte) []byte {
	var buffer bytes.Buffer
	for _, pByte := range pBytes {
		buffer.Write(pByte)
	}
	return buffer.Bytes()
}

/**
parse standardZip
Offset | Bytes | Description
12     | 4     | Size of central directory
16     | 4     | Offset of start of CD, relative to start of archive
*/
func parseToInt(eocd *[]byte) (int64, int64) {
	input := *eocd
	cdSize := binary.LittleEndian.Uint32(input[12:16])
	cdStart := binary.LittleEndian.Uint32(input[16:20])
	return int64(cdSize), int64(cdStart)
}

/**
parse zip64
Offset | Bytes | Description
40     | 8     | Size of central directory
48     | 8     | Offset of start of CD, relative to start of archive
*/
func parseToInt64(eocd64 *[]byte) (int64, int64) {
	input := *eocd64
	// 读8位
	cdSize := binary.LittleEndian.Uint64(input[40:48])
	cdStart := binary.LittleEndian.Uint64(input[48:56])
	return int64(cdSize), int64(cdStart)
}

// GetZipReaderFromUrl 从url获得 zipReader
func GetZipReaderFromUrl(url string) (*Reader, error) {

	//获得文件总大小
	totalSize, err := getFileSize(url)
	if err != nil {
		return nil, err
	}
	//获得eocd的数据
	eocdRecord, err := getRangeBytes(url, totalSize-EocdRecordSize, totalSize)
	if err != nil {
		return nil, err
	}

	//判断是普通的zip还是zip64
	if totalSize <= MaxStandardZipSize {
		// 如果是普通zip
		/*
		   totalMeta: central_directory + eocd_record
		*/
		//获得cd的信息
		cdSize, cdStart := parseToInt(eocdRecord)

		//获得cd的数据
		centralDirectory, err := getRangeBytes(url, cdStart, cdStart+cdSize-1)
		if err != nil {
			return nil, err
		}
		//组合cd + eocd 获得文件源信息
		totalMeta := BytesCombine(*centralDirectory, *eocdRecord)

		// 额外大小 = len(totalMeta)
		//totalExtraSize := cdSize + EocdRecordSize

		args := &InitArgs{
			IsZip64:              false,
			TotalSize:            totalSize,
			EOCDSize:             EocdRecordSize,
			CDSize:               cdSize,
			Zip64EocdRecordSize:  Zip64EocdRecordSize,
			Zip64EocdLocatorSize: Zip64EocdLocatorSize,
			ExtraSize:            int64(len(totalMeta)),
		}
		reader, err := NewReaderFromArgs(bytes.NewReader(totalMeta), args)
		return reader, err
	} else {
		/*
				As already mentioned,
				the ZIP and ZIP64 structure are a bit different.
				For the latter one,
				the algorithm looks also the same.
				The only difference is that you need to
				fetch an extra ZIP64 CD record and a ZIP64 CD locator.
				Then the four bytes blocks (CD+EOCD64 record+EOCD64 locator+CD) can be read and open as a ZIP file.
			    totalFetch: central_directory + zip64_eocd_record + zip64_eocd_locator + eocd_record
		*/
		//如果是zip64 超过4G的zip,还需要请求剩余的eocd

		// 获得 zip64EocdRecord的数据
		zip64EocdRecordStart := totalSize - (EocdRecordSize + Zip64EocdLocatorSize + Zip64EocdRecordSize)
		zip64EocdRecord, err := getRangeBytes(url,
			zip64EocdRecordStart,
			zip64EocdRecordStart+Zip64EocdRecordSize-1,
		)
		if err != nil {
			return nil, err
		}

		// 获得 zip64Eocdlocator的数据
		zip64EocdlocatorStart := totalSize - (EocdRecordSize + Zip64EocdLocatorSize)
		zip64Eocdlocator, err := getRangeBytes(url,
			zip64EocdlocatorStart,
			zip64EocdlocatorStart+Zip64EocdLocatorSize-1,
		)
		if err != nil {
			return nil, err
		}

		// 从  zip64EocdRecord 获得 zip64的 cd
		cdSize, cdStart := parseToInt64(zip64EocdRecord)
		centralDirectory, err := getRangeBytes(url, cdStart, cdStart+cdSize-1)
		if err != nil {
			return nil, err
		}
		//组合central_directory + zip64_eocd_record + zip64_eocd_locator + eocd_record 获得文件源信息
		totalMeta := BytesCombine(*centralDirectory, *zip64EocdRecord, *zip64Eocdlocator, *eocdRecord)

		args := &InitArgs{
			IsZip64:              true,
			TotalSize:            totalSize,
			EOCDSize:             EocdRecordSize,
			CDSize:               cdSize,
			Zip64EocdRecordSize:  Zip64EocdRecordSize,
			Zip64EocdLocatorSize: Zip64EocdLocatorSize,
			ExtraSize:            int64(len(totalMeta)),
		}

		reader, err := NewReaderFromArgs(bytes.NewReader(totalMeta), args)
		return reader, err
	}

}

func PrintZipFiles(reader *Reader) {
	fmt.Println("Files Tree:")

	fmt.Printf("%s | %s | %s | %s\n", "No.", "Name", "CompressedSize", "UncompressedSize")
	for index, file := range reader.File {
		name, _ := getDecodeFileName(file, "")
		fmt.Printf("%d | %s | %d | %d\n", index, name, file.CompressedSize64, file.UncompressedSize64)
	}
	fmt.Printf("TotalCount:%d, TotalCompressedSize: %d TotalUncompressedSize: %d\n", len(reader.File), reader.FileCompressedSize64, reader.FileUncompressedSize64)
}

func UnZipFilesFromNumbers(url string, reader *Reader, targetPath string, nos ...int) (int64, error) {
	if len(nos) == 0 {
		return 0, NoUnZipNumbersChoice
	}
	lastIndex := len(reader.File) - 1
	token := make(chan int, 64)
	var wg sync.WaitGroup
	var err error
	// ecod的偏移
	eocdOffest := reader.CD.DirectoryOffset
	var total int64
	for _, no := range nos {
		wg.Add(1)
		// 从no获得下标
		noIndex := no - 1
		if noIndex < 0 || noIndex > lastIndex {
			return 0, errors.New(fmt.Sprintf("noIndex < 0 or noIndex > lastIndex,noIndex is invalid,index=%d", noIndex))
		}
		go func(noIndex int) {
			defer wg.Done()
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
			f, err := useOldFileGetFullNewFile(url, nowFile, start, end)
			if err != nil {
				return
			}
			_, file, err := unzipFile(f, targetPath, "")
			if err != nil {
				return
			}
			total += file
			<-token
		}(noIndex)
	}
	wg.Wait()
	close(token)
	return total, err
}

func UnZipFilesFromRange(url string, reader *Reader, targetPath string, startNo, endNo int) (int64, error) {
	var input []int
	for i := startNo; i <= endNo; i++ {
		input = append(input, i)
	}
	return UnZipFilesFromNumbers(url, reader, targetPath, input...)

}

func UnZipAllFiles(url string, reader *Reader, targetPath string) (int64, error) {
	return UnZipFilesFromRange(url, reader, targetPath, 1, len(reader.File))
}

//useOldFileGetFullNewFile 用旧的文件 + bytes 获得新的文件
func useOldFileGetFullNewFile(url string, f *File, firstOffset, secondOffset int64) (*File, error) {
	fileBytes, err := getRangeBytes(url, firstOffset, secondOffset)
	if err != nil {
		return nil, err
	}
	firstFileReader := bytes.NewReader(*fileBytes)
	// 设置file 从 0开始读 给的字节也是请求的字节
	// 拷贝一个f
	newf := &File{
		FileHeader:   f.FileHeader,
		Zip:          f.Zip,
		Zipr:         f.Zipr,
		Zip64:        f.Zip64,
		DescErr:      f.DescErr,
		HeaderOffset: f.HeaderOffset,
	}
	// 修改File对应的数据区域
	newf.Zip.R = firstFileReader
	newf.Zipr = firstFileReader
	//注意偏移一定要设置为0
	newf.HeaderOffset = newFirstOffest
	return newf, nil
}

func unzipFile(f *File, dst string, charsetName string) (string, int64, error) {
	//可能要进行编码
	decodeName, err := getDecodeFileName(f, charsetName)
	if err != nil {
		return "", 0, err
	}
	destination := filepath.Join(dst, decodeName)
	if f.FileInfo().IsDir() {
		//如果这个文件是文件夹 直接创建文件夹即可
		//fmt.Printf("成功创建文件夹%s", destination)
		os.MkdirAll(destination, os.ModePerm)
		return destination, 0, nil
	} else {
		//如果是文件夹套的文件 先建文件夹
		if err := os.MkdirAll(filepath.Dir(destination), os.ModePerm); err != nil {
			return "", 0, err
		}
		destinationFile, err := os.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return "", 0, err
		}
		defer destinationFile.Close()
		sourceFile, err := f.Open()
		if err != nil {
			return "", 0, err
		}
		defer sourceFile.Close()

		n, err := io.Copy(destinationFile, sourceFile)
		return destination, n, err
	}

}

//getDecodeFileName 获得decode的Name
func getDecodeFileName(f *File, charSetName string) (string, error) {
	var (
		decodeName string
		err        error
	)
	//fmt.Printf("%+v\n", f)
	if f.Flags == 0 {
		//fmt.Printf("采用日语编码...\n")
		////如果标致位是0  则是默认的本地编码   默认为gbk
		//i := bytes.NewReader([]byte(f.Name))
		//decoder := transform.NewReader(i, japanese.ISO2022JP.NewDecoder())
		//content, _ := ioutil.ReadAll(decoder)
		//decodeName = string(content)
		decodeName, err = FromXXToUTF8(f.Name, charSetName)

	} else {
		//如果标志为是 1 << 11也就是 2048  则是utf-8编码
		decodeName = f.Name
	}
	return decodeName, err
}

func (args *UnzipProps) GetZipReader() (*Reader, error) {
	// get Numbers
	return GetZipReaderFromUrl(args.Url)

}

func (args *UnzipProps) InfoPrint(reader *Reader) error {
	fmt.Println("Files Tree:")

	fmt.Printf("%s | %s | %s | %s\n", "No.", "Name", "CompressedSize", "UncompressedSize")
	for index, file := range reader.File {
		name, err := getDecodeFileName(file, args.CharsetName)
		if err != nil {
			return err
		}
		fmt.Printf("%d | %s | %d | %d\n", index, name, file.CompressedSize64, file.UncompressedSize64)
	}
	fmt.Printf("TotalCount:%d, TotalCompressedSize: %d TotalUncompressedSize: %d\n", len(reader.File), reader.FileCompressedSize64, reader.FileUncompressedSize64)
	return nil
}

func (args *UnzipProps) Unzip(reader *Reader) (int64, *[]*ResultProps, error) {
	lastIndex := len(reader.File) - 1
	var numbers []int
	//获得 index
	if args.UnzipAll {
		//for all
		for i := 1; i <= len(reader.File); i++ {
			numbers = append(numbers, i)
		}
	} else if args.RangeStart > 0 && args.RangeEnd > 0 {
		if args.RangeStart > args.RangeEnd {
			return 0, nil, errors.New(fmt.Sprintf("args.RangeEnd=%d > args.RangeStart=%d ", args.RangeEnd, args.RangeStart))
		}
		for i := args.RangeStart; i <= args.RangeEnd; i++ {
			numbers = append(numbers, i)
		}
	} else if len(args.Numbers) != 0 {
		numbers = append(numbers, args.Numbers...)
	}
	if len(numbers) <= 0 {
		return 0, nil, NoUnZipNumbersChoice

	}
	url := args.Url
	targetPath := args.TargetPath
	charsetName := args.CharsetName
	token := make(chan int, getSuitableTokenSize(int64(len(reader.File))))

	var wg sync.WaitGroup
	var wgResult sync.WaitGroup
	var err error
	results := make(chan *ResultProps, len(numbers))
	// ecod的偏移
	eocdOffest := reader.CD.DirectoryOffset
	var total int64
	var result []*ResultProps
	//处理结果
	go func() {
		wgResult.Add(1)
		for res := range results {
			result = append(result, res)
		}
		wgResult.Done()
	}()
	for _, no := range numbers {
		wg.Add(1)
		// 从no获得下标
		noIndex := no - 1
		if noIndex < 0 || noIndex > lastIndex {
			return 0, nil, errors.New(fmt.Sprintf("noIndex < 0 or noIndex > lastIndex,noIndex is invalid,index=%d", noIndex))
		}
		go func(noIndex int) {
			defer wg.Done()
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
			f, err := useOldFileGetFullNewFile(url, nowFile, start, end)
			result := &ResultProps{
				IsSuccess:         false,
				SuccessHandleByte: 0,
				Err:               err,
				File:              f,
				FileRangeStart:    start,
				FileRangeEnd:      end,
				CDStart:           eocdOffest,
				ZipReader:         reader,
				FileIndex:         noIndex,
			}
			if err != nil {
				results <- result
				return
			}

			unZipedFilePath, fileZipByteTotal, err := unzipFile(f, targetPath, charsetName)
			if err != nil {
				result.Err = err
				results <- result
				return
			}
			// 回收bytes 防止内存溢出
			recycleBytes(f)
			total += fileZipByteTotal
			result.IsSuccess = true
			result.SuccessHandleByte = fileZipByteTotal
			result.FilePath = unZipedFilePath
			results <- result
			<-token
		}(noIndex)
	}
	wg.Wait()
	close(results) //关闭 并不影响接收遍历
	wgResult.Wait()
	close(token)

	return total, &result, err
}

func getSuitableTokenSize(total int64) int {
	if total <= 100 {
		return 20
	} else if total > 100 && total < 800 {
		return 32
	} else {
		return 64
	}

}

func ResultPrint(results *[]*ResultProps) {
	fmt.Println("Results Tree:")
	if len(*results) <= 0 {
		return
	}

	fmt.Printf("%s | %s | %s | %s | %s | %s | %s | %s | %s | %s \n", "No.", "OriginFileName", "FilePath", "FileSize", "FileIndex", "FileRangeStart", "FileRangeEnd", "CDStart", "isSuccess", "Error")
	index := 0
	errorCount := 0
	successCount := 0
	var totalUnzipedSize int64
	for _, result := range *results {
		//fmt.Printf("%+v", result)
		errorInfo := ""
		if result.Err != nil {
			errorInfo = result.Err.Error()
		}
		originName := ""
		if result.File != nil {
			originName = result.File.Name
		}
		fmt.Printf("%d | %s | %s | %d | %d | %d | %d | %d | %v | %s \n", index, originName, result.FilePath, result.SuccessHandleByte, result.FileIndex, result.FileRangeStart, result.FileRangeEnd, result.CDStart, result.IsSuccess, errorInfo)
		index++
		totalUnzipedSize += result.SuccessHandleByte
		if result.IsSuccess {
			successCount++
		} else {
			errorCount++
		}
	}
	fmt.Printf("TotalCount:%d, UnZipedSize: %d SuccessCount: %d\n ErrorCount: %d\n", index+1, totalUnzipedSize, successCount, errorCount)

}
func recycleBytes(f *File) {
	f.Zipr = nil
	f.Zip = nil
}
