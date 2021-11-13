package main

import (
	"fmt"
	_ "freechat/im/fc-im-grpc-server"
	recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/liyanze888/funny-core/fn_factory"
	_ "github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_grpc"
	_ "github.com/liyanze888/funny-core/fn_grpc"
	"github.com/liyanze888/funny-core/fn_grpc/fn_grpc_config"
	"github.com/liyanze888/funny-core/fn_grpc/grpc_interceptor/grpc_interceptor_server"
	"github.com/liyanze888/funny-core/fn_log"
	"log"
	"runtime/debug"
)

func init() {
	recoveryHandlerOption := recovery.WithRecoveryHandler(func(p interface{}) (err error) {
		debug.PrintStack()
		err = fmt.Errorf("panic: %v", p)
		return
	})

	grpc_interceptor_server.CreateUnaryInterceptors(
		recovery.UnaryServerInterceptor(recoveryHandlerOption),
		grpc_interceptor_server.UnaryLogServerInterceptor(),
	)

	grpc_interceptor_server.CreateStreamInterceptors(
		recovery.StreamServerInterceptor(recoveryHandlerOption),
		grpc_interceptor_server.StreamLogServerInterceptor(),
	)

	fn_factory.BeanFactory.RegisterBean(&fn_grpc_config.FnGrpcConfig{
		Port: 50051,
	})

}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile)
	fn_log.Printf("%v", "start main")
	fn_factory.BeanFactory.StartUp()
	fn_grpc.GrpcBeanFactory.StartUp()
}
