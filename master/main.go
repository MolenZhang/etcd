package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"code.safe.molen.com/molen/haoma/greedy/master/config"
	"code.safe.molen.com/molen/haoma/greedy/master/core"
	"code.safe.molen.com/molen/haoma/greedy/master/routers"
	mgo "code.safe.molen.com/molen/haoma/greedy/master/storage/mongo"
	mysql "code.safe.molen.com/molen/haoma/greedy/master/storage/mysql"

	"github.com/MolenZhang/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	debug := log.Config{
		Level:              zapcore.DebugLevel, // 日志级别 debug级别下会打印所有类型日志
		EncodeLogsAsJSON:   false,              // 输出日志格式是为json格式
		FileLoggingEnabled: true,               // 输出日志是否保存到文件
		StdLoggingDisabled: true,               // 日志是否标准输入 与上一项保存到文件互斥
		MaxSize:            100000,             // 日志文件最大限制 单位Mb
		MaxBackups:         3,                  // 最大保留备份数
		MaxAge:             7,                  // 保存的天数
		IsAddCaller:        true,               // 是否开启调用追踪
		CallerSkip:         1,
		Directory:          "./log/",
		Filename:           "greedy-master.log",
	}

	debug.Init()

	defer func() {
		if r := recover(); r != nil {
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Debugf("%s Recovery panic recovered: %v\n%s%s\n",
				time.Now().Format("2006-01-02 15:04:05"), r, buf, string([]byte{27, 91, 48, 109}))
		}
	}()

	// 仅仅适配云手机 初始化master任务管理器
	if err := core.InitJobManager([]string{"10.99.91.62:8378"}); err != nil {
		panic(fmt.Sprintf("Init Job Manager error: %v", err))
	}

	// init config
	// TODO add config
	config.InitConfig()

	// init mongo
	if err := mgo.InitMongo(config.Cfg.Mongo); err != nil {
		panic(fmt.Sprintf("Init Mongo error: %v", err))
	}

	// init mysql
	if err := mysql.InitMysql(config.Cfg.Mysql); err != nil {
		panic(fmt.Sprintf("Init Mongo error: %v", err))
	}

	srv := &http.Server{
		Addr:              ":8585",
		Handler:           routers.Handler,
		ReadTimeout:       time.Second * 120,
		ReadHeaderTimeout: time.Second * 120,
		WriteTimeout:      time.Second * 120,
		IdleTimeout:       time.Second * 120,
		MaxHeaderBytes:    10240,
	}

	log.Debugf("listen at \x1b[95m%s\x1b[0m", ":8585")
	if err := srv.ListenAndServe(); err != nil {
		log.Info("listen %s", zap.Error(err))
	}

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	log.Info("shutting down ...")
	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("shutting down : %s", err)
	}
	log.Info("Shutdown byte")
}
