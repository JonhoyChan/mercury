package x

import "fmt"

func Sprintf(format string, v ...interface{}) string {
	return fmt.Sprintf(format, v...)
}
