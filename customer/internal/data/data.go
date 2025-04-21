package data

import (
	"customer/internal/biz"
	"customer/internal/conf"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
// 增加 NewCustomerData 的 provider
var ProviderSet = wire.NewSet(NewData, NewGreeterRepo, NewCustomerData)

// Data .
type Data struct {
	// TODO wrapped database client
	// 操作Redis的客户端
	Rdb *redis.Client
	// 操作MySQL的客户端
	Mdb *gorm.DB
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	data := &Data{}
	// 一、初始化 Rdb
	// 连接redis，使用服务的配置，c就是解析后的配置信息(来自于config.yaml)
	redisURL := fmt.Sprintf("redis://%s/1?dial_timeout=%d", c.Redis.Addr, 1)
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		data.Rdb = nil
		log.Fatal(err)
	}
	// new client 不会立即连接，建立客户端，需要执行命令时才会连接
	data.Rdb = redis.NewClient(options)

	// 二、初始化Mdb
	// 连接mysql，使用配置
	dsn := c.Database.Source
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		data.Mdb = nil
		log.Fatal(err)
	}
	data.Mdb = db

	// 三、开发阶段，自动迁移表。发布阶段，表结构稳定，不需要migrate
	migrateTable(db)
	cleanup := func() {
		// 清理了 Redis 连接
		_ = data.Rdb.Close()
		log.NewHelper(logger).Info("closing the data resources")
	}
	return data, cleanup, nil
}

func migrateTable(db *gorm.DB) {
	if err := db.AutoMigrate(&biz.Customer{}); err != nil {
		log.Info("customer table migrate error,err:", err)
	}
}
