package models

import (
	"time"

	"gorm.io/gorm"
)

type UserSession struct {
	gorm.Model
	UserId       int       `gorm:"not null;comment:用户id"`
	Status       int       `gorm:"size:2;default:0;comment: 状态,0排队中,1连接中,2已连接,3已销毁"`
	JupyterHubId int       `gorm:"default:0;comment:连接的jupyterHub"`
	Token        string    `gorm:"size:255;default:'';comment:jupyter token"`
	TokenId      string    `gorm:"size:32;default:'';comment: the token id used to delete"`
	Url          string    `gorm:"size:255;default:'';comment:docker server链接"`
	ConnectedAt  time.Time `gorm:"comment:连接时间，超过一小时自动销毁"`
}
