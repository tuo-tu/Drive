package server

import (
	"context"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// 跨域资源共享的中间件
func MWCors() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				// Do something on entering
				// 在响应中增加特定的头信息
				ht := tr.(http.Transporter)
				// 允许的域
				ht.ReplyHeader().Set("Access-Control-Allow-Origin", "*")
				// 允许的请求方法
				ht.ReplyHeader().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS,PUT,PATCH,DELETE")
				//允许证书认证
				ht.ReplyHeader().Set("Access-Control-Allow-Credentials", "true")
				// 允许的头
				ht.ReplyHeader().Set("Access-Control-Allow-Headers", "Content-Type,User-Agent,Content-Length,Authorization,Accept,Referer,Host")

				defer func() {
					// Do something on exiting
				}()
			}
			return handler(ctx, req)
		}
	}
}
