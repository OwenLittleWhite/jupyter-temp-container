package controller

import (
	"manager/dao/mysql"
	"manager/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
)

func CreateServerNodeHandler(c *gin.Context) {
	p := new(models.ParamPostServerNode)
	if err := c.ShouldBindJSON(p); err != nil {
		errs, ok := err.(validator.ValidationErrors) // 类型断言
		if !ok {
			ResponseError(c, CodeInvalidParam)
			return
		} // 翻译并去除掉错误提示中的结构体标识
		ResponseErrorWithMsg(c, CodeInvalidParam, errs)
		return
	}

	var node = models.ServerNode{Url: p.Url}
	result := mysql.Db.Create(&node)
	if result.Error != nil {
		ResponseError(c, CodeServerBusy)
	}
	ResponseSuccess(c, node)
}

func GetServerNodeHandler(c *gin.Context) {
	var nodes []models.ServerNode
	mysql.Db.Find(&nodes).Limit(20)
	ResponseSuccess(c, nodes)
}
