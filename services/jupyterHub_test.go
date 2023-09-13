package services

import (
	"fmt"
	"manager/test"
	"testing"
)

func TestStartJupyterHubIfNotAvailable(t *testing.T) {
	// 在这里编写您的测试逻辑
	test.Setup(t)
	// defer test.Teardown(t)
	id, err := StartJupyterHubIfNotAvailable()
	if err != nil {
		t.Errorf("StartJupyterHubIfNotAvailable function returned an error: %v", err)
	}
	fmt.Println(id)
	// 在这里添加其他测试逻辑
}

func TestProcessHub(t *testing.T) {
	test.Setup(t)
	err := ProcessHub()
	if err != nil {
		t.Errorf("TestProcessHub function returned an error: %v", err)
	}
}
