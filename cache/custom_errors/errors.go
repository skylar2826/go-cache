package custom_errors

import (
	"fmt"
)

func ErrFieldNotFound(str string) error {
	return fmt.Errorf("字段不存在：%s\n", str)
}

func ErrFieldSetFailed(str string) error {
	return fmt.Errorf("字段设置失败：%s\n", str)
}

func ErrFieldOverMaxSize(str string) error {
	return fmt.Errorf("缓存已达上限：%s\n", str)
}
