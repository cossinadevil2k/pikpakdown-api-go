package httpHandler

import (
	"fmt"
	"github.com/mrxtryagin/pikpakdown-api-go/utils"
	"io"
	"net/http"
	"path"
	"sync"
)

// GeneralClient 通用 HTTP Client
var GeneralClient Client = NewClient()

// Client 请求客户端 共用接口
type Client interface {
	Request(method, target string, body io.Reader, opts ...Option) *Response
}

// HTTPClient 实现 Client 接口
type HTTPClient struct {
	// 暴力互斥锁类似于 java的lock
	mu      sync.Mutex
	options *options
}

//NewClient 初始化client
func NewClient(opts ...Option) *HTTPClient {
	client := &HTTPClient{
		options: newDefaultOption(),
	}

	for _, o := range opts {
		//修改options
		o.apply(client.options)
	}

	return client
}

// Request 发送HTTP请求
func (c *HTTPClient) Request(method, target string, body io.Reader, opts ...Option) *Response {

	// 应用额外设置
	c.mu.Lock()
	options := *c.options
	c.mu.Unlock()
	for _, o := range opts {
		//修改options
		o.apply(&options)
	}
	if options.endpoint != nil {
		targetURL := *options.endpoint
		targetURL.Path = path.Join(targetURL.Path, target)
		target = targetURL.String()
	}
	if options.queryString != "" {
		target += "?" + options.queryString
	}
	//utils.Log().Info(target)
	defer HttpTimer(fmt.Sprintf("[target: %s method: %s]", target, method))()

	// 创建请求客户端(提供timeout)
	client := &http.Client{Timeout: options.timeout}
	if options.proxyUrl != nil {
		utils.Log().Debug("使用代理:%s", options.proxyUrl.String())
		client.Transport = &http.Transport{
			Proxy: http.ProxyURL(options.proxyUrl),
		}
	}
	//判断重定向
	if !options.isRedirect {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			//使用上一次的response
			return http.ErrUseLastResponse
		}
	}

	// size为0时将body设为nil
	if options.contentLength == 0 {
		body = nil
	}

	// 创建请求
	var (
		req *http.Request
		err error
	)
	if options.ctx != nil {
		//使用上下文的request 这样的话 request 可以通过context 进行进一步的管理
		req, err = http.NewRequestWithContext(options.ctx, method, target, body)
	} else {
		req, err = http.NewRequest(method, target, body)
	}
	if err != nil {
		//返回体
		return &Response{Err: err}
	}

	// 添加请求相关设置
	if options.header != nil {
		for k, v := range options.header {
			//todo: 为什么使用strings.Join 而不是一个个加
			//req.Header.Add(k, strings.Join(v, " "))
			for _, val := range v {
				req.Header.Add(k, val)
			}
		}
	}

	// -1 content长度
	if options.contentLength != -1 {
		req.ContentLength = options.contentLength
	}

	//打印头部
	//util.Log().Info("请求头为:%v", req.Header)
	//util.Log().Info("url为:%s", req.URL.String())

	// 发送请求
	resp, err := client.Do(req)

	//req.Close = true
	if err != nil {
		return &Response{Err: err}
	}

	return &Response{Err: nil, Response: resp}
}

//Get
func (c *HTTPClient) Get(target string, body io.Reader, opts ...Option) *Response {

	return c.Request(http.MethodGet, target, body, opts...)
}

//Head
func (c *HTTPClient) Head(target string, body io.Reader, opts ...Option) *Response {
	return c.Request(http.MethodHead, target, body, opts...)
}

//Post
func (c *HTTPClient) Post(target string, body io.Reader, opts ...Option) *Response {
	return c.Request(http.MethodPost, target, body, opts...)
}

//Put
func (c *HTTPClient) Put(target string, body io.Reader, opts ...Option) *Response {
	return c.Request(http.MethodPut, target, body, opts...)
}

//Patch
func (c *HTTPClient) Patch(target string, body io.Reader, opts ...Option) *Response {
	return c.Request(http.MethodPatch, target, body, opts...)
}

//Delete
func (c *HTTPClient) Delete(target string, body io.Reader, opts ...Option) *Response {
	return c.Request(http.MethodDelete, target, body, opts...)
}

//Connect
func (c *HTTPClient) Connect(target string, body io.Reader, opts ...Option) *Response {
	return c.Request(http.MethodConnect, target, body, opts...)
}

//Connect
func (c *HTTPClient) Options(target string, body io.Reader, opts ...Option) *Response {
	return c.Request(http.MethodOptions, target, body, opts...)
}

//Trace
func (c *HTTPClient) Trace(target string, body io.Reader, opts ...Option) *Response {
	return c.Request(http.MethodTrace, target, body, opts...)
}
