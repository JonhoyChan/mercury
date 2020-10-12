package orm

import (
	"mercury/x/config"
	"mercury/x/log"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// NewORM new db and retry connection when has error.
func NewORM(c config.DatabaseProvider) (db *gorm.DB, err error) {
	db, err = gorm.Open("postgres", c.Database().DSN)
	if err != nil {
		log.Error("db dsn(%s) error: %v", c.Database().DSN, err)
		return
	}
	db.DB().SetMaxIdleConns(c.Database().Idle)
	db.DB().SetMaxOpenConns(c.Database().Active)
	db.DB().SetConnMaxLifetime(c.Database().IdleTimeout / time.Second)
	// 全局禁用表名复数
	db.SingularTable(true)
	db.LogMode(true)
	return
}
