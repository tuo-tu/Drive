package data

import (
	"context"
	"database/sql"
	"driver/api/verifyCode"
	"driver/internal/biz"
	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/hashicorp/consul/api"
	"time"
)

type DriverData struct {
	data *Data
}

func NewDriverInterface(data *Data) biz.DriverInterface {
	return &DriverData{data: data}
}

// 获取token的实现
func (dt *DriverData) GetToken(ctx context.Context, tel string) (string, error) {
	// 1.数据表查询
	driver := biz.Driver{}
	if err := dt.data.Mdb.Where("telephone=?", tel).First(&driver).Error; err != nil {
		return "", err
	}
	// 2.返回token
	return driver.Token.String, nil
}

// 获取已存储的验证码
func (dt *DriverData) GetSavedVerifyCode(ctx context.Context, tel string) (string, error) {
	return dt.data.Rdb.Get(ctx, "DVC:"+tel).Result()
}

// 存储token到数据表
func (dt *DriverData) SaveToken(ctx context.Context, tel, token string) error {
	//先获取司机信息
	driver := biz.Driver{}
	if err := dt.data.Mdb.Where("telephone=?", tel).First(&driver).Error; err != nil {
		return err
	}
	//再更新司机信息
	driver.Token = sql.NullString{
		String: token,
		Valid:  true,
	}
	if err := dt.data.Mdb.Save(&driver).Error; err != nil {
		return err
	}
	return nil
}

// 初始化司机信息
func (dt *DriverData) InitDriverInfo(ctx context.Context, tel string) (*biz.Driver, error) {
	// 入库，设置状态为stop
	driver := biz.Driver{}
	driver.Telephone = tel
	driver.Status = sql.NullString{
		String: "stop",
		Valid:  true,
	}
	if err := dt.data.Mdb.Create(&driver).Error; err != nil {
		return nil, err
	}
	return &driver, nil
}

func (dt *DriverData) GetVerifyCode(ctx context.Context, tel string) (string, error) {
	// grpc 请求
	consulConfig := api.DefaultConfig()
	consulConfig.Address = dt.data.cs.Consul.Address
	consulClient, err := api.NewClient(consulConfig)
	// 注册服务管理器
	dis := consul.New(consulClient)
	if err != nil {
		return "", err
	}
	endpoint := "discovery:///VerifyCode"
	conn, err := grpc.DialInsecure(
		//ctx,
		context.Background(),
		grpc.WithEndpoint(endpoint), // 目标服务的名字
		grpc.WithDiscovery(dis),     // 使用服务发现
	)
	if err != nil {
		return "", err
	}
	//关闭
	defer conn.Close()
	// 2.2,发送获取验证码请求
	client := verifyCode.NewVerifyCodeClient(conn)
	// 作为客户端在进行远程调用
	reply, err := client.GetVerifyCode(ctx, &verifyCode.GetVerifyCodeRequest{
		Length: 6,
		Type:   1,
	})
	if err != nil {
		return "", err
	}
	// 三，redis的临时存储
	// 设置key, customer-verify-code
	status := dt.data.Rdb.Set(ctx, "DVC:"+tel, reply.Code, 60*time.Second)
	if _, err := status.Result(); err != nil {
		return "", err
	}
	return reply.Code, nil
}

// 获取号码对应的验证码
func (dt *DriverData) FetchVerifyCode(ctx context.Context, telephone string) (string, error) {
	status := dt.data.Rdb.Get(context.Background(), "DVC:"+telephone)
	code, err := status.Result() // status.String()
	if err != nil {
		return "", err
	}
	return code, nil
}

func (dt *DriverData) FetchInfoByTel(ctx context.Context, tel string) (*biz.Driver, error) {
	driver := &biz.Driver{}
	if err := dt.data.Mdb.Where("telephone=?", tel).First(driver).Error; err != nil {
		return nil, err
	}
	return driver, nil
}
