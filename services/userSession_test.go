package services

import (
	"encoding/json"
	"fmt"
	"manager/test"
	"testing"
)

func TestCreateUserSession(t *testing.T) {
	// 在这里编写您的测试逻辑
	test.Setup(t)
	// defer test.Teardown(t)
	res, err := CreateUserSession(13, false)
	if err != nil {
		t.Errorf("CreateUserSession function returned an error: %v", err)
	}
	str, err := json.Marshal(res)
	if err != nil {
		t.Errorf("CreateUserSession json marshal function returned an error: %v", err)
	}
	fmt.Println(string(str))
	// 在这里添加其他测试逻辑
}

func TestSequProcessUserSession(t *testing.T) {
	test.Setup(t)
	SequProcessUserSession()
}

func TestDestroyUserSession(t *testing.T) {
	test.Setup(t)
	DestroyUserSession()
}
