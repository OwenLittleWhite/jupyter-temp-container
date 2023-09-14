package controller

import (
	"manager/api"
	"manager/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

func CreateUserSessionHandler(c *gin.Context) {
	var featureName = "temp_juptyer_server"
	_userId, exists := c.Get(CtxUserIDKey)
	userId := _userId.(int)
	if !exists {
		c.JSON(http.StatusUnauthorized, "Unauthorized")
		c.Abort()
		return
	}
	// 判断是否是vip
	isVip := api.CheckPermission(featureName, userId)
	userSessionRes, err := services.CreateUserSession(userId, isVip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
	} else {
		ResponseSuccess(c, userSessionRes)
	}
}
