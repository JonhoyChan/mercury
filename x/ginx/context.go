package ginx

import (
	jsoniter "github.com/json-iterator/go"
	"net/http"
	"outgoing/x/ecode"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type Context struct {
	*gin.Context
}

// API返回格式
type response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// 成功返回
func (c *Context) Success(obj interface{}) {
	c.Response(nil, obj)
}

// 错误返回
func (c *Context) Error(err error) {
	c.Response(err, nil)
}

// API返回
func (c *Context) Response(err error, obj interface{}) {
	code := ecode.Cause(err)
	c.Set(ContextEcodeKey, code)

	if obj == nil {
		obj = struct{}{}
	}

	c.JSON(http.StatusOK, &response{
		Code:    code.Code(),
		Message: code.Message(),
		Data:    obj,
	})
}

func JSON(c *Context, status int, obj interface{}) {
	buf, _ := jsoniter.Marshal(obj)
	c.Set(JSONBodyKey, buf)
	c.Data(status, "application/json; charset=utf-8", buf)
	c.Abort()
}

// 绑定请求信息 GET请求绑定Query POST绑定Body
func (c *Context) BindRequest(obj interface{}) (err error) {
	switch c.Request.Method {
	case http.MethodGet:
		err = c.BindQuery(obj)
	case http.MethodPost, http.MethodPut:
		err = c.ShouldBindBodyWith(obj, binding.JSON)
	}

	return err
}

func (c *Context) ParseIntParam(key string) (int64, error) {
	return strconv.ParseInt(c.Param(key), 10, 64)
}

func (c *Context) ParseUintParam(key string) (uint64, error) {
	return strconv.ParseUint(c.Param(key), 10, 64)
}

// 获取分页页数和分页条数
func (c *Context) GetPageAndSize() (uint32, uint32) {
	var err error
	// 获取分页页数,默认第一页
	page, err := strconv.Atoi(c.Query("page"))
	if err != nil {
		page = 1
	}
	if page < 1 {
		page = 1
	}
	// 获取分页条数, 默认一页十条
	size, err := strconv.Atoi(c.Query("size"))
	if err != nil {
		size = 10
	}
	if size < 1 {
		size = 10
	}

	return uint32(page), uint32(size)
}

func (c *Context) GetAPIKey() string {
	// Check header.
	apiKey := c.GetHeader(APIKeyHeader)

	// Check URL query parameters.
	if apiKey == "" {
		apiKey = c.Query("api_key")
	}

	return apiKey
}
