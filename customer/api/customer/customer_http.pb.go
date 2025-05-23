// Code generated by protoc-gen-go-http. DO NOT EDIT.
// versions:
// - protoc-gen-go-http v2.7.3
// - protoc             v5.26.1
// source: customer.proto

package customer

import (
	context "context"
	http "github.com/go-kratos/kratos/v2/transport/http"
	binding "github.com/go-kratos/kratos/v2/transport/http/binding"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the kratos package it is being compiled against.
var _ = new(context.Context)
var _ = binding.EncodeURL

const _ = http.SupportPackageIsVersion1

const OperationCustomerEstimatePrice = "/api.customer.Customer/EstimatePrice"
const OperationCustomerGetVerifyCode = "/api.customer.Customer/GetVerifyCode"
const OperationCustomerLogin = "/api.customer.Customer/Login"
const OperationCustomerLogout = "/api.customer.Customer/Logout"

type CustomerHTTPServer interface {
	// EstimatePrice 价格预估
	EstimatePrice(context.Context, *EstimatePriceReq) (*EstimatePriceResp, error)
	// GetVerifyCode 获取验证码
	GetVerifyCode(context.Context, *GetVerifyCodeReq) (*GetVerifyCodeResp, error)
	// Login 登录
	Login(context.Context, *LoginReq) (*LoginResp, error)
	// Logout 退出登陆
	Logout(context.Context, *LogoutReq) (*LogoutResp, error)
}

func RegisterCustomerHTTPServer(s *http.Server, srv CustomerHTTPServer) {
	r := s.Route("/")
	r.GET("/customer/get-verify-code/{telephone}", _Customer_GetVerifyCode0_HTTP_Handler(srv))
	r.POST("/customer/login", _Customer_Login0_HTTP_Handler(srv))
	r.GET("/customer/logout", _Customer_Logout0_HTTP_Handler(srv))
	r.GET("/customer/estimate-price/{origin}/{destination}", _Customer_EstimatePrice0_HTTP_Handler(srv))
}

func _Customer_GetVerifyCode0_HTTP_Handler(srv CustomerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in GetVerifyCodeReq
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationCustomerGetVerifyCode)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.GetVerifyCode(ctx, req.(*GetVerifyCodeReq))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*GetVerifyCodeResp)
		return ctx.Result(200, reply)
	}
}

func _Customer_Login0_HTTP_Handler(srv CustomerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in LoginReq
		if err := ctx.Bind(&in); err != nil {
			return err
		}
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationCustomerLogin)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.Login(ctx, req.(*LoginReq))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*LoginResp)
		return ctx.Result(200, reply)
	}
}

func _Customer_Logout0_HTTP_Handler(srv CustomerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in LogoutReq
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationCustomerLogout)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.Logout(ctx, req.(*LogoutReq))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*LogoutResp)
		return ctx.Result(200, reply)
	}
}

func _Customer_EstimatePrice0_HTTP_Handler(srv CustomerHTTPServer) func(ctx http.Context) error {
	return func(ctx http.Context) error {
		var in EstimatePriceReq
		if err := ctx.BindQuery(&in); err != nil {
			return err
		}
		if err := ctx.BindVars(&in); err != nil {
			return err
		}
		http.SetOperation(ctx, OperationCustomerEstimatePrice)
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			return srv.EstimatePrice(ctx, req.(*EstimatePriceReq))
		})
		out, err := h(ctx, &in)
		if err != nil {
			return err
		}
		reply := out.(*EstimatePriceResp)
		return ctx.Result(200, reply)
	}
}

type CustomerHTTPClient interface {
	EstimatePrice(ctx context.Context, req *EstimatePriceReq, opts ...http.CallOption) (rsp *EstimatePriceResp, err error)
	GetVerifyCode(ctx context.Context, req *GetVerifyCodeReq, opts ...http.CallOption) (rsp *GetVerifyCodeResp, err error)
	Login(ctx context.Context, req *LoginReq, opts ...http.CallOption) (rsp *LoginResp, err error)
	Logout(ctx context.Context, req *LogoutReq, opts ...http.CallOption) (rsp *LogoutResp, err error)
}

type CustomerHTTPClientImpl struct {
	cc *http.Client
}

func NewCustomerHTTPClient(client *http.Client) CustomerHTTPClient {
	return &CustomerHTTPClientImpl{client}
}

func (c *CustomerHTTPClientImpl) EstimatePrice(ctx context.Context, in *EstimatePriceReq, opts ...http.CallOption) (*EstimatePriceResp, error) {
	var out EstimatePriceResp
	pattern := "/customer/estimate-price/{origin}/{destination}"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation(OperationCustomerEstimatePrice))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "GET", path, nil, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *CustomerHTTPClientImpl) GetVerifyCode(ctx context.Context, in *GetVerifyCodeReq, opts ...http.CallOption) (*GetVerifyCodeResp, error) {
	var out GetVerifyCodeResp
	pattern := "/customer/get-verify-code/{telephone}"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation(OperationCustomerGetVerifyCode))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "GET", path, nil, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *CustomerHTTPClientImpl) Login(ctx context.Context, in *LoginReq, opts ...http.CallOption) (*LoginResp, error) {
	var out LoginResp
	pattern := "/customer/login"
	path := binding.EncodeURL(pattern, in, false)
	opts = append(opts, http.Operation(OperationCustomerLogin))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "POST", path, in, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *CustomerHTTPClientImpl) Logout(ctx context.Context, in *LogoutReq, opts ...http.CallOption) (*LogoutResp, error) {
	var out LogoutResp
	pattern := "/customer/logout"
	path := binding.EncodeURL(pattern, in, true)
	opts = append(opts, http.Operation(OperationCustomerLogout))
	opts = append(opts, http.PathTemplate(pattern))
	err := c.cc.Invoke(ctx, "GET", path, nil, &out, opts...)
	if err != nil {
		return nil, err
	}
	return &out, nil
}
