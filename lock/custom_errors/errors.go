package custom_errors

import "fmt"

func ErrLockIsNotMine(str string) error {
	return fmt.Errorf("不是你的锁 key: %s\n", str)
}

func ErrLockPreemptFailed(str string) error {
	return fmt.Errorf("加锁失败 key:%s\n", str)
}
