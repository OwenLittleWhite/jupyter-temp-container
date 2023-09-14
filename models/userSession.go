package models

import (
	"time"

	"gorm.io/gorm"
)

type UserSession struct {
	gorm.Model
	UserId       int        `gorm:"not null;comment:用户id" json:"userId"`
	Status       int        `gorm:"size:2;default:0;comment: 状态,0排队中,1已连接,2已销毁" json:"status"`
	JupyterHubId int        `gorm:"default:0;comment:连接的jupyterHub" json:"jupyterHubId"`
	Token        string     `gorm:"size:255;default:'';comment:jupyter token" json:"token"`
	TokenId      string     `gorm:"size:32;default:'';comment: the token id used to delete" json:"tokenId"`
	Url          string     `gorm:"size:255;default:'';comment:docker server链接" json:"url"`
	ConnectedAt  time.Time  `gorm:"default:null;comment:连接时间，超过一小时自动销毁" json:"connectedAt"`
	JupyterHub   JupyterHub `gorm:"foreignKey:JupyterHubId"`
}
