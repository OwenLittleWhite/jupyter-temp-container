package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"manager/models"
	"manager/setting"
	"net/http"
	"strconv"
)

func ValidateUserToken(token string) (valid bool, user models.AuthUser) {
	valid = false
	url := fmt.Sprintf("%s/internal/authenticated", setting.Conf.CoreserviceUrl)
	jsonData, _ := json.Marshal(&struct {
		Token string `json:"token"`
	}{Token: token})
	body := bytes.NewBuffer(jsonData)
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		fmt.Println("创建请求时出错:", err)

		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", setting.Conf.InternalKey)
	// 创建 HTTP 客户端
	client := &http.Client{}
	// 发起请求
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("发起请求时出错:", err)
		return
	}
	defer resp.Body.Close()
	// 读取响应数据
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取响应数据时出错:", err)
		return
	}

	// 处理响应
	if resp.StatusCode == http.StatusOK {
		valid = true
		json.Unmarshal(responseData, &user)
	}
	return
}

func CheckPermission(feature string, userId int) (has bool) {
	has = false
	url := fmt.Sprintf("%s/internal/permissions/check?userId=%v&featureName=%s", setting.Conf.CoreserviceUrl, userId, feature)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("创建请求时出错:", err)
		return
	}
	req.Header.Set("x-api-key", setting.Conf.InternalKey)
	// 创建 HTTP 客户端
	client := &http.Client{}
	// 发起请求
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("发起请求时出错:", err)
		return
	}
	defer resp.Body.Close()
	// 读取响应数据
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("读取响应数据时出错:", err)
		return
	}
	// 处理响应
	if resp.StatusCode == http.StatusOK {
		// 解析响应为布尔值
		result, err := strconv.ParseBool(string(responseData))
		if err != nil {
			fmt.Println("解析响应失败:", err)
			return
		}
		has = result
	}
	return
}
