package biz

import (
	"context"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"
	"gorm.io/gorm"
	"strconv"
	"valuation/api/mapService"
)

type PriceRuleWork struct {
	CityID      uint  `gorm:"" json:"city_id"`
	StartFee    int64 `gorm:"" json:"start_fee"`        // 起步费
	DistanceFee int64 `gorm:"" json:"distance_fee"`     // 里程费
	DurationFee int64 `gorm:"" json:"duration_fee"`     // 时长费
	StartAt     int   `gorm:"type:int" json:"start_at"` // 0 [0
	EndAt       int   `gorm:"type:int" json:"end_at"`   // 7 0)
}

type PriceRule struct {
	gorm.Model
	PriceRuleWork
}

// 定义操作priceRule的接口
type PriceRuleInterface interface {
	// 获取规则
	GetRule(cityid uint, curr int) (*PriceRule, error)
}

type ValuationBiz struct {
	pri PriceRuleInterface
}

func NewValuationBiz(pri PriceRuleInterface) *ValuationBiz {
	return &ValuationBiz{
		pri: pri,
	}
}

// 获取价格
func (vb *ValuationBiz) GetPrice(ctx context.Context, distance, duration string, cityId uint, curr int) (int64, error) {
	// 一，获取规则
	rule, err := vb.pri.GetRule(cityId, curr)
	if err != nil {
		return 0, err
	}
	// 二，将距离和时长，转换为int64
	distancInt64, err := strconv.ParseInt(distance, 10, 64)
	if err != nil {
		return 0, err
	}

	durationInt64, err := strconv.ParseInt(duration, 10, 64)
	if err != nil {
		return 0, err
	}

	// 三，基于rule计算
	distancInt64 /= 1000        // 公里
	durationInt64 /= 60         // 分钟？
	var startDistance int64 = 5 // 起始距离 5公里？
	total := rule.StartFee +
		rule.DistanceFee*(distancInt64-startDistance) +
		rule.DurationFee*durationInt64
	return total, nil
}

// 获取距离和时长
func (*ValuationBiz) GetDrivingInfo(ctx context.Context, origin, destination string) (distance string, duration string, err error) {
	// 一，发出GRPC请求
	// 使用服务发现
	// 1.获取consul客户端
	consulConfig := api.DefaultConfig()
	consulConfig.Address = "192.168.43.144:8500"
	consulClient, err := api.NewClient(consulConfig)
	if err != nil {
		return
	}
	// 2.获取服务发现管理器
	dis := consul.New(consulClient)
	// 2.1,连接目标grpc服务器
	endpoint := "discovery:///Map"
	conn, err := grpc.DialInsecure(
		context.Background(),
		//ctx,
		grpc.WithEndpoint(endpoint), // 目标服务的名字
		grpc.WithDiscovery(dis),     // 使用服务发现
		// 中间件
		grpc.WithMiddleware(
			// tracing 的客户端中间件
			tracing.Client(),
		),
	)

	if err != nil {
		return
	}
	//关闭
	defer func() {
		_ = conn.Close()
	}()

	// 2.2,发送获取驾驶距离和时长请求，RPC调用
	client := mapService.NewMapServiceClient(conn)
	reply, err := client.GetDrivingInfo(context.Background(), &mapService.GetDrivingInfoReq{
		Origin:      origin,
		Destination: destination,
	})

	if err != nil {
		return
	}
	distance, duration = reply.Distance, reply.Duration
	// 返回正确信息
	return
}
