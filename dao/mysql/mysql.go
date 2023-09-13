package mysql

import (
	"database/sql"
	"fmt"
	"log"
	"manager/models"
	"manager/setting"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var Db *gorm.DB
var sqlDB *sql.DB

// Init 初始化MySQL连接
func Init(cfg *setting.MySQLConfig) (err error) {

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&loc=Local", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB)
	Db, err = gorm.Open(mysql.New(mysql.Config{
		DriverName: "mysql",
		DSN:        dsn, // data source name, 详情参考：https://github.com/go-sql-driver/mysql#dsn-data-source-name
	}), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{LogLevel: logger.Info}),
	})

	if err != nil {
		return
	}
	sqlDB, err = Db.DB()
	if err != nil {
		return
	}
	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)
	syncToDb()
	return
}

func syncToDb() {
	Db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&models.ServerNode{})
	Db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&models.JupyterHub{})
	Db.Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(&models.UserSession{})
}

// Close 关闭MySQL连接
func Close() {
	_ = sqlDB.Close()
}
