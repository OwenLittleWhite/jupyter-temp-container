package services

import (
	"testing"
)

func TestCreateDocker(t *testing.T) {
	// 在这里编写您的测试逻辑

	// 调用 Connect 函数并检查是否出现错误
	_, err := CreateDocker(18555, "tcp://localhost:2375")
	if err != nil {
		t.Errorf("create function returned an error: %v", err)
	}
	// err = Destroy(containerId)
	// if err != nil {
	// 	t.Errorf("destroy function returned an error: %v", err)
	// }

	// 在这里添加其他测试逻辑
}
