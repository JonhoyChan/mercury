package ecode

import (
	"context"
	"outgoing/x/log"
	"regexp"
	"strconv"
	"strings"

	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/registry"
	"github.com/micro/go-micro/v2/server"
	"github.com/pkg/errors"
)

var (
	microErrorRegexp = regexp.MustCompile(`"code":(.*?),"detail":"(.*?)"`)
	errInternal      = ErrInternalServer.ResetMessage("服务器内部错误")
)

func ParseMicroError(err error) error {
	match := microErrorRegexp.FindStringSubmatch(err.Error())
	if len(match) != 3 {
		return errInternal
	}

	microCode, _ := strconv.Atoi(match[1])
	if microCode == 408 {
		return ErrRequestTimeout.ResetMessage("服务请求超时")
	}

	microError := match[2]
	sep := ": "
	if strings.Contains(microError, sep) {
		e := strings.Split(microError, sep)
		code, err := strconv.Atoi(e[1])
		if err != nil {
			return errInternal
		}

		return Code{code: code, message: e[0]}
	}

	return errInternal
}

func FormatMicroError(err error) error {
	_, ok := errors.Cause(err).(Coder)
	if ok {
		return err
	} else {
		return Wrap(ErrInternalServer, "服务器内部错误")
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
		err := fn(ctx, req, rsp)
		if err != nil {
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
			log.Error(err.Error())
			return ParseMicroError(err)
		}
		return nil
	}
}
