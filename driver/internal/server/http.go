package server

import (
	"context"
	"driver/api/driver"
	v1 "driver/api/helloworld/v1"
	"driver/internal/biz"
	"driver/internal/conf"
	"driver/internal/service"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	jwtv5 "github.com/golang-jwt/jwt/v5"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// NewHTTPServer new an HTTP server.
func NewHTTPServer(c *conf.Server, greeter *service.GreeterService, driverService *service.DriverService, logger log.Logger) *http.Server {
	var opts = []http.ServerOption{
		http.Middleware(
			recovery.Recovery(),
			// JWT 中间件
			selector.Server(
				jwt.Server(func(token *jwtv5.Token) (interface{}, error) { // 这个中间件有什么用？
					return []byte(biz.SecretKey), nil // 用于验证客户端发送的JWT令牌；
				}),
				DriverToken(driverService),
			).Match(func(ctx context.Context, operation string) bool {
				// 记录不需要的校验的
				noJWT := map[string]struct{}{
					"/api.driver.Driver/Login":         {},
					"/api.driver.Driver/GetVerifyCode": {},
					"/api.driver.Driver/SubmitPhone":   {},
				}
				if _, exists := noJWT[operation]; exists {
					return false
				}
				return true
			}).Build(),
		),
	}
	if c.Http.Network != "" {
		opts = append(opts, http.Network(c.Http.Network))
	}
	if c.Http.Addr != "" {
		opts = append(opts, http.Address(c.Http.Addr))
	}
	if c.Http.Timeout != nil {
		opts = append(opts, http.Timeout(c.Http.Timeout.AsDuration()))
	}
	srv := http.NewServer(opts...)
	v1.RegisterGreeterHTTPServer(srv, greeter)
	driver.RegisterDriverHTTPServer(srv, driverService)
	return srv
}
