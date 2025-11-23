package common

// RequestWithBody 有 body 的请求
type RequestWithBody interface {
	// Body 返回 body 的指针
	Body() interface{}
}
