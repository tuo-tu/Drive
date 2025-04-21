package data

import (
	"driver/internal/biz"
	"driver/internal/conf"
	"fmt"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewDriverInterface)

// Data .
type Data struct {
	// TODO wrapped database client
	// 操作MySQL的客户端
	Mdb *gorm.DB
	// 操作Redis的客户端
	Rdb *redis.Client
	// 中间件服务器配置
	cs *conf.Service
}

// NewData .
func NewData(c *conf.Data, cs *conf.Service, logger log.Logger) (*Data, func(), error) {
	data := &Data{
		cs: cs,
	}
	// 连接redis，使用服务的配置，c就是解析之后的配置信息,此处的redis没有配置密码
	redisURL := fmt.Sprintf("redis://%s/2?dial_timeout=%d", c.Redis.Addr, 1)
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		data.Rdb = nil
		log.Fatal(err)
	}
	// new client 不会立即连接，建立客户端，需要执行命令时才会连接
	data.Rdb = redis.NewClient(options)

	// 初始Mdb
	// 连接mysql，使用配置
	dsn := c.Database.Source
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		data.Mdb = nil
		log.Fatal(err)
	}
	data.Mdb = db
	// 三，开发阶段，自动迁移表。发布阶段，表结构稳定，不需要migrate.
	migrateTable(db)

	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return data, cleanup, nil
}

func migrateTable(db *gorm.DB) {
	// 自动迁移相关表
	if err := db.AutoMigrate(&biz.Driver{}); err != nil {
		log.Fatal(err)
	}
}
