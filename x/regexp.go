package x

import (
	"regexp"
)

// 校验手机格式
var mobilePattern = `^((13[0-9])|(14[1,4,5,6,7,8])|(15[^4])|(16[5-7])|(17[0-8])|(18[0-9])|(19[1,8,9]))\d{8}$`

// 校验IP地址格式
var ipPattern = `^(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)\.(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)\.(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)\.(25[0-5]|2[0-4]\d|[0-1]\d{2}|[1-9]?\d)$`

func MatchMobile(mobile string) bool {
	matched, err := regexp.MatchString(mobilePattern, mobile)
	if err != nil {
		return false
	}
	return matched
}

func MatchIP(ip string) bool {
	matched, err := regexp.MatchString(ipPattern, ip)
	if err != nil {
		return false
	}
	return matched
}

// 显示前三后四
func ReplaceMobile(mobile string) string {
	var re *regexp.Regexp
	// 国内手机号固定长度为11
	if len(mobile) != 11 {
		return mobile
	} else {
		re = regexp.MustCompile("(\\d{3})\\d{4}(\\d{4})")
	}
	return re.ReplaceAllString(mobile, "$1****$2")
}

func ReplaceHttpOrHttps(url string) string {
	re := regexp.MustCompile("http[s]*://")
	return re.ReplaceAllString(url, "")
}
