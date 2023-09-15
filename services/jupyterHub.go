package services

import (
	"fmt"
	"io"
	"manager/dao/mysql"
	"manager/dao/redis"
	"manager/models"
	"manager/setting"
	"math/rand"
	"net/http"
	"time"

	"gorm.io/gorm"
)

func contains(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
func getRandomPort(exclude []int) (randomValue int) {
	for {
		// 生成随机数，范围在17000到18000之间
		randomValue = rand.Intn(1001) + 17000
		// 检查随机数是否在排除列表中
		if !contains(exclude, randomValue) {
			return randomValue
		}
	}
}

func createJupyterServer(internalIp string, port int, username string) (err error) {
	var protocol = setting.Conf.Protocol
	// 创建一个 HTTP 请求
	url := fmt.Sprintf("%s://%s:%v/hub/api/users/%s/server", protocol, internalIp, port, username)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		fmt.Println("创建请求时出错:", err)
		return
	}
	// 添加自定义标头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", setting.Conf.JupyterHubConfig.DefaultToken))
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
	if resp.StatusCode == http.StatusCreated {
		fmt.Println("请求成功")
		fmt.Println("响应数据:", string(responseData))
		// 在此处处理成功的响应
	} else {
		err = fmt.Errorf("请求失败，状态码: %d, 链接：%s", resp.StatusCode, url)
		// 在此处处理失败的响应
		return
	}
	return
}

func waitForServiceAvailability(url string) error {
	maxAttempts := 5
	for i := 0; i < maxAttempts; i++ {
		resp, err := http.Get(url)
		if err == nil && resp.StatusCode == http.StatusOK {
			return nil // 服务就绪
		}
		fmt.Printf("等待服务就绪，尝试 #%d\n", i+1)
		time.Sleep(2 * time.Second) // 等待 2 秒后重试
	}
	return fmt.Errorf("服务未能就绪")
}
func StartJupyterHubIfNotAvailable() (jupyterHubId uint, err error) {
	id, err := redis.RunWithLock("StartJupyterHubIfNotAvailable", time.Second*30, func() (id interface{}, err error) {
		var protocol = setting.Conf.Protocol
		// 先查找是否有服务中的Hub
		var hub models.JupyterHub
		mysql.Db.Where("status = 1 and num < ?", setting.Conf.JupyterHubConfig.MaxSessionSize).Take(&hub)
		if hub.ID != 0 {
			return hub.ID, nil
		}
		// 超过总hub数则返回
		var total int64
		mysql.Db.Model(&models.JupyterHub{}).Where("status=1").Count(&total)
		if total >= int64(setting.Conf.JupyterHubConfig.MaxSize) {
			return 0, nil
		}
		// 无运行中的hub则创建
		var node models.ServerNode
		mysql.Db.Where("status=1").Order("num asc").Take(&node)
		if node.ID == 0 {
			return 0, fmt.Errorf("no available server")

		}
		// 查出来已经运行的hub占用的端口，避免重复
		var alreadyHubs []models.JupyterHub
		mysql.Db.Where("(status=1 or status=2) and server_node_id = ?", node.ID).Find(&alreadyHubs)
		var ports = make([]int, 0)
		for _, aHub := range alreadyHubs {
			ports = append(ports, aHub.Port)
		}
		port := getRandomPort(ports)
		containerId, err := CreateDocker(port, fmt.Sprintf("tcp://%s:2375", node.InternalIp))
		if err != nil {
			return
		}
		err = waitForServiceAvailability(fmt.Sprintf("%s://%s:%v/hub/api", protocol, node.InternalIp, port))
		if err != nil {
			return
		}
		err = createJupyterServer(node.InternalIp, port, "share")
		if err != nil {
			return
		}
		mysql.Db.Model(&node).Update("num", gorm.Expr("num+1"))
		hub = models.JupyterHub{
			ServerNodeId: node.ID,
			DockerId:     containerId,
			Port:         port,
			StartAt:      time.Now(),
			Status:       1,
		}
		result := mysql.Db.Create(&hub) // 通过数据的指针来创建
		if err := result.Error; err != nil {
			return 0, err
		}
		return hub.ID, nil
	})

	if err != nil {
		return
	}
	val, ok := id.(uint)
	if !ok {
		return 0, nil
	}
	jupyterHubId = val
	return jupyterHubId, nil
}

func ProcessHub() (err error) {
	_, err = redis.RunWithLock("ProcessHub", 10*time.Second, func() (interface{}, error) {
		// 首次连接时间超过一小时的变为半关闭
		now := time.Now()
		oneHourAgo := now.Add(-time.Hour)
		result := mysql.Db.Model(&models.JupyterHub{}).
			Where("status = ? AND first_connected_at <= ? AND first_connected_at is not null", 1, oneHourAgo).
			Updates(map[string]interface{}{"status": 2})
		if result.Error != nil {
			return nil, result.Error
		}
		// 首次连接时间超过三小时的要删除容器，状态改为已关闭
		threeHourAgo := now.Add(-time.Hour * 3)
		var hubs []*models.JupyterHub
		mysql.Db.Preload("ServerNode").Where("status=? AND first_connected_at <= ? AND first_connected_at is not null", 2, threeHourAgo).Find(&hubs)
		for _, hub := range hubs {
			// 删除容器
			err := DestroyDocker(hub.DockerId, fmt.Sprintf("tcp://%s:2375", hub.ServerNode.InternalIp))
			if err != nil {
				fmt.Printf("destroy docker failed: %v", err)
			}
			mysql.Db.Model(&hub).Update("status", 3)
		}
		// 更新serverNodes的运行数量
		mysql.Db.Exec(`
		UPDATE server_nodes AS sn
		SET sn.num = (SELECT COUNT(*) FROM jupyter_hubs AS jh WHERE jh.server_node_id = sn.id AND jh.status = 1)`)
		return nil, nil

	})
	return
}
