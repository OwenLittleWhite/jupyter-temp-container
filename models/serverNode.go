package models

import "gorm.io/gorm"

type ServerNode struct {
	gorm.Model
	InternalIp string `gorm:"size:255;default:'';comment:内网ip"`
	ExternalIp string `gorm:"size:255;default:'';"`
	Num        int    `gorm:"size:10;default:0;comment:运行的docker数"`
	Status     int    `gorm:"size:2;default:1;comment: 连接状态,0掉线,1在线,2禁用"`
}
