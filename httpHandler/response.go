package httpHandler

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

// Response 请求的响应或错误信息
type Response struct {
	Err      error
	Response *http.Response
}

// GetResponse 检查响应并获取响应正文([]byte]
func (resp *Response) GetResponse() ([]byte, error) {
	if resp.Err != nil {
		return nil, resp.Err
	}
	var (
		respBody []byte
		err      error
	)
	retry := 3
	errorCount := 1
	for retry > 0 {
		respBody, err = ioutil.ReadAll(resp.Response.Body)
		if err == nil {
			break
		} else {
			retry--
			fmt.Printf("errorCount %d retry_left %d,error:%s\n", errorCount, retry, err.Error())
			errorCount++
			time.Sleep(1 * time.Second)
		}

	}

	_ = resp.Response.Body.Close()

	return respBody, err
}

// CheckHTTPResponseCode 检查请求响应HTTP状态码
func (resp *Response) CheckHTTPResponseCode(status int) *Response {
	if resp.Err != nil {
		return resp
	}

	// 检查HTTP状态码
	if resp.Response.StatusCode != status {

		resp.Err = fmt.Errorf("服务器返回非正常HTTP状态, 需要 %d,却发现是 %d", status, resp.Response.StatusCode)
	}
	return resp
}

//CheckHttpStatusOk 检查返回值是否是200
func (resp *Response) CheckHttpStatusOk() *Response {
	return resp.CheckHTTPResponseCode(200)
}

//CheckHttpStatusByFunc 使用func 来检查
func (resp *Response) CheckHttpStatusByFunc(checkFunc func(res *http.Response) error) *Response {
	if resp.Err != nil {
		return resp
	}

	err := checkFunc(resp.Response)
	if err != nil {
		resp.Err = err
	}

	return resp
}

// NopRSCloser 实现不完整seeker
type NopRSCloser struct {
	body   io.ReadCloser
	status *rscStatus
}

type rscStatus struct {
	// http.ServeContent 会读取一小块以决定内容类型，
	// 但是响应body无法实现seek，所以此项为真时第一个read会返回假数据
	IgnoreFirst bool

	Size int64
}

// GetRSCloser 返回带有空seeker的RSCloser，供http.ServeContent使用
func (resp *Response) GetRSCloser() (*NopRSCloser, error) {
	if resp.Err != nil {
		return nil, resp.Err
	}

	return &NopRSCloser{
		body: resp.Response.Body,
		status: &rscStatus{
			Size: resp.Response.ContentLength,
		},
	}, resp.Err
}

// SetFirstFakeChunk 开启第一次read返回空数据
// TODO 测试
func (instance NopRSCloser) SetFirstFakeChunk() {
	instance.status.IgnoreFirst = true
}

// SetContentLength 设置数据流大小
func (instance NopRSCloser) SetContentLength(size int64) {
	instance.status.Size = size
}

// Read 实现 NopRSCloser reader
func (instance NopRSCloser) Read(p []byte) (n int, err error) {
	if instance.status.IgnoreFirst && len(p) == 512 {
		return 0, io.EOF
	}
	return instance.body.Read(p)
}

// Close 实现 NopRSCloser closer
func (instance NopRSCloser) Close() error {
	return instance.body.Close()
}

// Seek 实现 NopRSCloser seeker, 只实现seek开头/结尾以便http.ServeContent用于确定正文大小
func (instance NopRSCloser) Seek(offset int64, whence int) (int64, error) {
	// 进行第一次Seek操作后，取消忽略选项
	if instance.status.IgnoreFirst {
		instance.status.IgnoreFirst = false
	}
	if offset == 0 {
		switch whence {
		case io.SeekStart:
			return 0, nil
		case io.SeekEnd:
			return instance.status.Size, nil
		}
	}
	return 0, errors.New("未实现")

}
