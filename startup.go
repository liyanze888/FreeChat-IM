package main

import (
	"context"
	"github.com/liyanze888/funny-core/fn_factory"
	"github.com/liyanze888/funny-core/fn_utils"
	"github.com/lqs/sqlingo"
	"log"
	"reflect"
	"time"
)

func init() {
	db, err := sqlingo.Open("mysql", "root:liyanze@3.1415926@tcp(152.136.28.100:3306)/freechat-im")
	if err != nil {
		panic(err)
	}
	db.GetDB().SetMaxOpenConns(5)
	db.GetDB().SetMaxIdleConns(5)
	db.EnableCallerInfo(true)
	db.SetInterceptor(func(ctx context.Context, sql string, invoker sqlingo.InvokerFunc) error {
		start := time.Now()
		defer func() {
			log.Printf("%v %s", time.Since(start), sql)
		}()
		return invoker(ctx, sql)
	})

	t := reflect.TypeOf(db)
	for i := 0; i < t.Elem().NumField(); i++ {

	}
	fn_factory.BeanFactory.RegisterBean(db)
	fn_factory.BeanFactory.RegisterBean(fn_utils.NewIdUtils())
}
