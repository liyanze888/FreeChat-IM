package grpc_common

import (
	"errors"
	gatewaypb "freechat/im/generated/grpc/im/gateway"
	"google.golang.org/grpc/metadata"
	"reflect"
	"strconv"
)

const (
	UserIdDescribe = "UserId"
)

type GrpcContextInfo struct {
	Login     bool   //是否登陆
	UserId    int64  `desc:"userId"`      //用户类型
	BizType   int64  `desc:"bizType"`     //业务类型
	UniqueKey string `desc:"instance-id"` //唯一标识 多端登录
	Server    gatewaypb.ImService_ConnectServer
}

func Tranfer2ContextInfo(server gatewaypb.ImService_ConnectServer) (*GrpcContextInfo, error) {
	md, b := metadata.FromIncomingContext(server.Context())

	if !b {
		return nil, errors.New("Parse Headers Failed")
	}
	context := &GrpcContextInfo{
		Server: server,
	}
	t := reflect.TypeOf(context)
	v := reflect.ValueOf(context)

	fieldSize := t.Elem().NumField()
	for i := 0; i < fieldSize; i++ {
		field := t.Elem().Field(i)
		if desc, ok := field.Tag.Lookup("desc"); ok {
			values := md.Get(desc)
			if len(values) > 0 {
				switch field.Type.Kind() {
				case reflect.Int:
					parseInt, err := strconv.ParseInt(values[0], 10, 10)
					if err != nil {
						return nil, err
					}
					v.Elem().Field(i).Set(reflect.ValueOf(parseInt))
					break
				case reflect.Int32:
					parseInt, err := strconv.ParseInt(values[0], 10, 32)
					if err != nil {
						return nil, err
					}
					v.Elem().Field(i).Set(reflect.ValueOf(parseInt))
					break
				case reflect.Int64:
					parseInt, err := strconv.ParseInt(values[0], 10, 64)
					if err != nil {
						return nil, err
					}
					v.Elem().Field(i).Set(reflect.ValueOf(parseInt))
					break
				case reflect.String:
					v.Elem().Field(i).Set(reflect.ValueOf(values[0]))
					break
				}

			}
		}
	}
	return context, nil
}
