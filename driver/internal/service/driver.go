package service

import (
	"context"
	"driver/internal/biz"
	"log"
	"time"

	pb "driver/api/driver"
)

type DriverService struct {
	pb.UnimplementedDriverServer
	Bz *biz.DriverBiz
}

func NewDriverService(bz *biz.DriverBiz) *DriverService {
	return &DriverService{
		Bz: bz,
	}
}

// IDNoCheck 校验身份证号码
func (s *DriverService) IDNoCheck(ctx context.Context, req *pb.IDNoCheckReq) (*pb.IDNoCheckResp, error) {
	return &pb.IDNoCheckResp{}, nil
}

func (s *DriverService) GetVerifyCode(ctx context.Context, req *pb.GetVerifyCodeReq) (*pb.GetVerifyCodeResp, error) {
	// 获取验证码
	code, err := s.Bz.GetVerifyCode(ctx, req.Telephone)
	if err != nil {
		return &pb.GetVerifyCodeResp{
			Code:    1,
			Message: err.Error(),
		}, nil
	}
	// 响应
	return &pb.GetVerifyCodeResp{
		Code:           0,
		Message:        "SUCCESS",
		VerifyCode:     code,
		VerifyCodeTime: time.Now().Unix(),
		VerifyCodeLife: 60,
	}, nil
}

// 提交电话号码
func (s *DriverService) SubmitPhone(ctx context.Context, req *pb.SubmitPhoneReq) (*pb.SubmitPhoneResp, error) {
	// 将司机信息入库，并设置状态为stop暂时停用(核心逻辑）
	driver, err := s.Bz.InitDriverInfo(ctx, req.Telephone)
	if err != nil {
		return &pb.SubmitPhoneResp{
			Code:    1,
			Message: "司机号码提交失败",
		}, nil
	}

	return &pb.SubmitPhoneResp{
		Code:    0,
		Message: "司机号码提交成功",
		Status:  driver.Status.String,
	}, nil
}

// 登录 service
func (s *DriverService) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginResp, error) {
	// 由 biz 层完成业务逻辑处理
	token, err := s.Bz.CheckLogin(ctx, req.Telephone, req.VerifyCode)
	if err != nil {
		log.Println(err)
		return &pb.LoginResp{
			Code:    1,
			Message: "司机登录失败",
		}, nil
	}
	return &pb.LoginResp{
		Code:          0,
		Message:       "司机登录成功",
		Token:         token,
		TokenCreateAt: time.Now().Unix(),
		TokenLife:     biz.DriverTokenLife,
	}, nil
}
