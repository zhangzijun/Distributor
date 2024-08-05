package main

import (
	"Distributor/Api"
	"Distributor/Background"
	"Distributor/Cron"
	"Distributor/model"
	"time"
)

func main() {
	println("hello")

	// 添加日志，确认代码运行到此处
	println("Initializing database...")

	// 初始化数据库连接
	model.InitDB()
	println("Database initialized")

	// 添加日志，确认代码运行到此处
	println("Initializing cron jobs...")

	// 启动定时任务
	go Cron.InitJobs()
	println("Cron jobs initialized")

	// 添加日志，确认代码运行到此处
	println("Setting up API server...")

	// 启动API服务
	go func() {
		Api.ApiServer()
		println("API server started")
	}()

	go func() {
		Background.BackgroundServer()
		println("BackGround server started")
	}()

	// 保持主进程运行，等待退出信号
	for {
		time.Sleep(10 * time.Second) // 间隔一定时间进行检查或其他操作
	}
}
