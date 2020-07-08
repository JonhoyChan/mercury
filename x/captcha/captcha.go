package captcha

import (
	"math/rand"
	"time"
)

const (
	numberCaptcha         = `0123456789`
	alphaCaptcha          = `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz`
	alphaAndNumberCaptcha = `0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz`
)

// 生成指定位数的数字验证码
func GenerateNumber(n int) string {
	return generate(n, numberCaptcha)
}

// 生成指定位数的字母验证码
func GenerateAlpha(n int) string {
	return generate(n, alphaCaptcha)
}

// 生成指定位数的字母和数字混合验证码
func GenerateAlphaAndNumber(n int) string {
	return generate(n, alphaAndNumberCaptcha)
}

// 生成指定位数的验证码
func generate(n int, captchaType string) string {
	var captchaBytes = []byte(captchaType)

	if n <= 0 {
		return ""
	}

	var bytes = make([]byte, n)
	var randBy bool
	if num, err := rand.Read(bytes); num != n || err != nil {
		rand.Seed(time.Now().UnixNano())
		randBy = true
	}
	for i, b := range bytes {
		if randBy {
			bytes[i] = captchaBytes[rand.Intn(len(captchaBytes))]
		} else {
			bytes[i] = captchaBytes[b%byte(len(captchaBytes))]
		}
	}

	return string(bytes)
}
