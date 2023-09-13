package models

import (
	"time"

	"gorm.io/gorm"
)

type JupyterHub struct {
	gorm.Model
	ServerNodeId     uint       `gorm:"not null;"`
	Status           int        `gorm:"size:2;default:1;comment:状态,1服务中,2半关闭,3已关闭"`
	Type             int        `gorm:"size:2;default:0;comment:类型，0共享，1独享"`
	Username         string     `gorm:"size:32;default:share;comment:jupyter用户名，默认是share"`
	DockerId         string     `gorm:"size:255;comment:docker container id"`
	Port             int        `gorm:"size:10;not null;comment:jupyterHub的启动端口"`
	Num              int        `gorm:"size:10;default:0;comment:正在运行的用户数"`
	StartAt          time.Time  `gorm:"comment:启动时间点"`
	FirstConnectedAt time.Time  `gorm:"default:null;comment:首次连接时间，超过一小时则变为半关闭，超过三小时自动关闭"`
	ServerNode       ServerNode `gorm:"foreignKey:ServerNodeId"`
}
