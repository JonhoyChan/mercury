package ecode

import (
	"context"
	"regexp"
	"strconv"
	"strings"

	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"
	"github.com/pkg/errors"
)

var microErrorRegexp = regexp.MustCompile(`"code":(.*?),"detail":"(.*?)"`)

func ParseMicroError(err error) error {
	match := microErrorRegexp.FindStringSubmatch(err.Error())
	if len(match) != 3 {
		return err
	}

	microCode, _ := strconv.Atoi(match[1])
	if microCode == 408 {
		return ErrRequestTimeout.ResetMessage("服务请求超时")
	}

	microError := match[2]
	sep := ": "
	if strings.Contains(microError, sep) {
		e := strings.Split(microError, sep)
		code, iErr := strconv.Atoi(e[0])
		if iErr != nil {
			return err
		}

		return Code{code: code, message: e[1]}
	}

	return err
}

func FormatMicroError(err error) error {
	coder, ok := errors.Cause(err).(Coder)
	if ok {
		return Wrap(coder, strconv.Itoa(coder.Code()))
	} else {
		coder = ErrInternalServer
		return Wrap(coder.ResetMessage(err.Error()), strconv.Itoa(coder.Code()))
	}
}

// 自定义请求错误重试
func RetryOnMicroError(ctx context.Context, req client.Request, retryCount int, err error) (bool, error) {
	if err == nil {
		return false, nil
	}

	code, ok := errors.Cause(err).(Coder)
	if !ok {
		return true, nil
	}

	switch code.Code() {
	// retry on timeout or internal server error
	case 408, 500:
		return true, nil
	default:
		return false, nil
	}
}

// 服务端 - 格式化成 ecode 错误
func MicroHandlerFunc(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		if err := fn(ctx, req, rsp); err != nil {
			return FormatMicroError(err)
		}
		return nil
	}
}

// 客户端 - 解析成 ecode 错误
func MicroCallFunc(fn client.CallFunc) client.CallFunc {
	return func(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) error {
		err := fn(ctx, node, req, rsp, opts)
		if err != nil {
			return ParseMicroError(err)
		}
		return nil
	}
}
