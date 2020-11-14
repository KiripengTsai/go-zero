package clientinterceptors

import (
	"context"
	"path"

	"github.com/tal-tech/go-zero/core/breaker"
	"github.com/tal-tech/go-zero/zrpc/internal/codes"
	"google.golang.org/grpc"
)

//熔断器引入
func BreakerInterceptor(ctx context.Context, method string, req, reply interface{},
	cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	//基于请求办法进行熔断
	breakerName := path.Join(cc.Target(), method)
	return breaker.DoWithAcceptable(breakerName, func() error {
		//调用函数
		return invoker(ctx, method, req, reply, cc, opts...)
		//判断熔断错误的类型
	}, codes.Acceptable)
}
