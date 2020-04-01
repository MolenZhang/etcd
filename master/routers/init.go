package routers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler .
var Handler *gin.Engine

func init() {
	Handler = gin.New()
	Handler.Use(gin.Recovery())

	Handler.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "STATUS OK")
	})

	greedy := Handler.Group("/greedy/")

	// 任务相关
	job := greedy.Group("/job/")
	{
		job.POST("/", JobSave)        // 增加任务
		job.GET("/:id", JobGet)       // 任务列表
		job.DELETE("/:id", JobDelete) // 删除任务
		job.GET("/", JobLists)        // 任务列表
		job.PUT("/")                  // 理论上不需要 任务一旦处于可执行状态是无法修改的 如果要修改任务 建议重新创建任务
		job.POST("/kill")             // TODO 强杀任务
	}

	// 数据相关
	data := greedy.Group("/data/")
	{
		data.GET("/phones", Phones)  // 下发给云手机的号码数据
		data.POST("/report", Report) // 号码结果上报
		data.GET("/", Show)          // 库里所有已抓取数据结果展示
	}

	// TODO 策略相关
	policy := greedy.Group("/policy")
	{
		policy.POST("/save")     // 增加策略
		policy.DELETE("/delete") // 删除策略
		policy.PUT("/tweak")     // 调整策略
		policy.GET("/fetch")     // 获取单个策略
		policy.GET("/list")      // 获取所有策略
	}
}
