package server

import (
	"context"
	"customer/internal/service"
	"fmt"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/transport"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"strings"
)

// customerJWT 生成中间件的方法，将ctx的请求头中的token与数据库中的token进行比较，不符合则不给通过；
func customerJWT(customerService *service.CustomerService) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// 一，获取存储在jwt中的用户（顾客）id
			claims, ok := jwt.FromContext(ctx)
			fmt.Println(claims, ok)
			if !ok {
				// 没有获取到 claims
				return nil, errors.Unauthorized("UNAUTHORIZED", "claims not found")
			}
			// 1.2 断言使用
			claimsMap := claims.(jwtv5.MapClaims)
			id := claimsMap["jti"] //获取JWT的ID

			// 二，获取id对应的顾客的token
			token, err := customerService.CD.GetToken(id)
			if err != nil {
				return nil, errors.Unauthorized("UNAUTHORIZED", "customer not found")
			}

			// 三，比对数据表中的token与请求的token是否一致
			// 获取请求头
			header, _ := transport.FromServerContext(ctx)
			// 从header获取token
			auths := strings.SplitN(header.RequestHeader().Get("Authorization"), " ", 2)
			jwtToken := auths[1]
			// 比较请求中的token与数据表中获取的token是否一致
			if jwtToken != token {
				return nil, errors.Unauthorized("UNAUTHORIZED", "token was updated")
			}
			// 四，校验通过，发行，继续执行
			// 交由下个中间件（handler）处理
			return handler(ctx, req)
		}
	}
}
