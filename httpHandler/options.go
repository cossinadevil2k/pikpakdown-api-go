package httpHandler

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

// Option 发送请求的额外设置, 接口
type Option interface {
	apply(*options)
}

type options struct {
	timeout time.Duration //timeout
	header  http.Header   //header
	//sign          auth.Auth     //签名
	//signTTL       int64         //签名时间
	ctx           context.Context //ctx上下文
	contentLength int64           //内容长度
	isRedirect    bool            //是否重定向
	//masterMeta    bool
	endpoint *url.URL
	//slaveNodeID   string
	queryString string
	proxyUrl    *url.URL
	retryCount  int
	retryFunc   func(repose *http.Response, otherError error) bool
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) {
	//调用传入的函数
	f(o)
}

//newDefaultOption 默认Option
func newDefaultOption() *options {
	return &options{
		header:        http.Header{},
		timeout:       time.Duration(30) * time.Second,
		isRedirect:    true, //允许重定向(目前就是走默认实现)
		contentLength: -1,
	}
}

// WithTimeout 设置请求超时
func WithTimeout(t time.Duration) Option {
	//强转匿名函数为optionFunc函数,走的是 optionFunc的实现(apply)
	return optionFunc(func(o *options) {
		o.timeout = t
	})
}

// WithContext 设置请求上下文
func WithContext(c context.Context) Option {
	return optionFunc(func(o *options) {
		// 在go > 1.7之后可以借助ctx 对request 进行一定的处理 比如超时(context.WithTimeout) https://www.cnblogs.com/ricklz/p/14840205.html
		o.ctx = c
	})
}

// WithCredential 对请求进行签名
//func WithCredential(instance auth.Auth, ttl int64) Option {
//	return optionFunc(func(o *options) {
//		o.sign = instance
//		o.signTTL = ttl
//	})
//}

// WithHeader 设置请求Header
func WithHeader(header http.Header) Option {
	return optionFunc(func(o *options) {
		for k, v := range header {
			o.header[k] = v
		}
	})
}

// WithoutHeader 设置清除请求Header
func WithoutHeader(header []string) Option {
	return optionFunc(func(o *options) {
		for _, v := range header {
			delete(o.header, v)
		}

	})
}

// WithEndpoint 使用同一的请求Endpoint
func WithEndpoint(endpoint string) Option {
	endpointURL, _ := url.Parse(endpoint)
	return optionFunc(func(o *options) {
		o.endpoint = endpointURL
	})
}

// WithQueryString query
func WithQueryString(queryString string) Option {
	return optionFunc(func(o *options) {
		o.queryString = queryString
	})
}

// WithContentLength 设置请求大小
func WithContentLength(s int64) Option {
	return optionFunc(func(o *options) {
		o.contentLength = s
	})
}

//WithProxy 使用代理
func WithProxy(proxyUrl string) Option {
	proxyUri, _ := url.Parse(proxyUrl)
	return optionFunc(func(o *options) {
		o.proxyUrl = proxyUri
	})
}

//WithRedirect 重定向
func WithRedirect(isRedirect bool) Option {
	return optionFunc(func(o *options) {
		o.isRedirect = isRedirect
	})
}

func WithRetry(retryCount int, retryFunc func(repose *http.Response, otherError error) bool) Option {
	return optionFunc(func(o *options) {
		o.retryCount = retryCount
		o.retryFunc = retryFunc
	})
}

func WithOnlyRetry(retryCount int) Option {
	return optionFunc(func(o *options) {
		o.retryCount = retryCount
	})
}
