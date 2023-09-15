package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"manager/dao/mysql"
	redisDao "manager/dao/redis"
	"manager/models"
	"manager/setting"
	"net/http"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type UserSessionRes struct {
	UserSession *models.UserSession `json:"userSession"`
	Rank        int64               `json:"rank"`
}

type TokenRes struct {
	Token string `json:"token"`
	Id    string `json:"id"`
}

var vipQueueName = "VIP_user_session_queue"
var commonQueueName = "COMMON_user_session_queue"

func CreateUserSession(userId int, isVip bool) (userSessionRes UserSessionRes, err error) {
	var userSession models.UserSession
	client := redisDao.RedisClient()
	result := mysql.Db.Where("user_id=? and status<>2", userId).Take(&userSession)
	if result.Error != nil && result.Error == gorm.ErrRecordNotFound {
		userSession = models.UserSession{
			UserId: userId,
			Status: 0,
		}
		mysql.Db.Create(&userSession)
	}
	if userSession.Status == 1 {
		return UserSessionRes{
			UserSession: &userSession,
			Rank:        0,
		}, nil
	}
	ctx := context.Background()
	timestamp := float64(time.Now().UnixNano()) / float64(time.Second)
	var queue string
	if isVip {
		queue = vipQueueName
	} else {
		queue = commonQueueName
	}
	client.ZAdd(ctx, queue, &redis.Z{
		Member: userId,
		Score:  timestamp,
	}).Result()
	rank, err := client.ZRank(ctx, queue, strconv.Itoa(userId)).Result()
	if err != nil {
		return
	}
	userSessionRes = UserSessionRes{
		UserSession: &userSession,
		Rank:        rank,
	}
	return
}

func SequProcessUserSession() {
	var protocol = setting.Conf.Protocol
	client := redisDao.RedisClient()
	ctx := context.Background()

	redisDao.RunWithLock("SequProcessUserSession", 30*time.Second, func() (rtn interface{}, err error) {
		// 先判断有没有可用的jupyterHub
		// 先查找是否有服务中的Hub
		var hub models.JupyterHub
		mysql.Db.Preload("ServerNode").Where("status = 1 and num < ?", setting.Conf.JupyterHubConfig.MaxSessionSize).Take(&hub)
		if hub.ID == 0 {
			return
		}
		var userId int
		// 先从VIP队列取数据
		result, err := client.ZPopMin(ctx, vipQueueName, 1).Result()
		if err != nil {
			fmt.Printf("SequProcessUserSession: Error: %v\n", err)
			return
		}

		if len(result) == 0 {
			// 从普通队列取数据
			result, err = client.ZPopMin(ctx, commonQueueName, 1).Result()
			if err != nil {
				fmt.Printf("SequProcessUserSession: Error: %v\n", err)
				return
			}
		}
		if len(result) == 0 {
			return
		}
		processData := result[0]
		userId, err = strconv.Atoi(processData.Member.(string))
		if err != nil {
			return
		}
		// 创建token
		tokenRes, err := postUserToken(hub.ServerNode.InternalIp, hub.Port, hub.Username)
		if err != nil {
			return
		}
		var userSession models.UserSession
		mysql.Db.Where("user_id = ? and status = 0", userId).Take(&userSession)
		if userSession.ID == 0 {
			return
		}
		mysql.Db.Model(&userSession).Updates(map[string]interface{}{"jupyter_hub_id": hub.ID, "token": tokenRes.Token, "token_id": tokenRes.Id, "url": fmt.Sprintf("%s://%s:%v/user/%s", protocol, hub.ServerNode.ExternalIp, hub.Port, hub.Username), "connected_at": time.Now(), "status": 1})

		if hub.FirstConnectedAt.IsZero() {
			mysql.Db.Model(&hub).Update("first_connected_at", time.Now()).Update("num", gorm.Expr("num+1"))
		}
		return nil, nil
	})

}

func postUserToken(internalIp string, port int, username string) (tokenRes TokenRes, err error) {
	var protocol = setting.Conf.Protocol
	// 创建一个 HTTP 请求
	url := fmt.Sprintf("%s://%s:%v/hub/api/users/%s/tokens", protocol, internalIp, port, username)
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
		err = json.Unmarshal(responseData, &tokenRes)
		if err != nil {
			fmt.Println("解析token数据出错", err)
			return
		}

	} else {
		err = fmt.Errorf("请求失败，状态码: %d, 链接：%s", resp.StatusCode, url)
		// 在此处处理失败的响应
		return
	}
	return
}

func deleteUserToken(internalIp string, port int, username string, tokenId string) (err error) {
	var protocol = setting.Conf.Protocol
	// 创建一个 HTTP 请求
	url := fmt.Sprintf("%s://%s:%v/hub/api/users/%s/tokens/%s", protocol, internalIp, port, username, tokenId)
	req, err := http.NewRequest("DELETE", url, nil)
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
	if resp.StatusCode == http.StatusNoContent {
		fmt.Println("请求成功")
		fmt.Println("响应数据:", string(responseData))
	} else {
		err = fmt.Errorf("请求失败，状态码: %d, 链接：%s", resp.StatusCode, url)
		// 在此处处理失败的响应
		return
	}
	return
}
func DestroyUserSession() (err error) {
	_, err = redisDao.RunWithLock("DestroyUserSession", 10*time.Second, func() (interface{}, error) {
		// 一小时自动销毁
		now := time.Now()
		oneHourAgo := now.Add(-time.Hour)
		var userSessions []*models.UserSession
		mysql.Db.Preload("JupyterHub.ServerNode").Where("status = ? AND connected_at <= ? AND connected_at is not null", 1, oneHourAgo).Find(&userSessions)
		for _, userSession := range userSessions {
			// 删除token
			err := deleteUserToken(userSession.JupyterHub.ServerNode.InternalIp, userSession.JupyterHub.Port, userSession.JupyterHub.Username, userSession.TokenId)
			if err != nil {
				fmt.Printf("delete token failed: %v", err)
			}
			mysql.Db.Model(&userSession).Update("status", 2)
		}
		// 更新jupyterHubs的运行数量
		mysql.Db.Exec(`
		UPDATE jupyter_hubs AS jh
		SET jh.num = (SELECT COUNT(*) FROM user_sessions AS us WHERE us.jupyter_hub_id = jh.id AND us.status = 1)`)
		return nil, nil
	})
	return
}
