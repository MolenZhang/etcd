package storage

// JobEtcd 保存到etcd中的任务详情
type JobEtcd struct {
	ID          string        `json:"id"`     // 任务id 唯一标识
	Name        string        `json:"name"`   // 任务名称
	Source      JobDataSource `json:"source"` // 数据来源
	CrawlType   int           // 执行方式 1-模拟器;2-云手机;3-api
	Description string        `json:"desc"`        // 备注
	Status      int           `json:"status"`      // 状态 1-未执行 2-执行中 3-执行完成
	Result      int           `json:"result"`      // 执行结果 1-成功 2-失败
	CreatedAt   string        `json:"create_time"` // 任务创建的时间
	StartedAt   string        `json:"start_time"`  // 开始执行的时间
	StopedAt    string        `json:"stop_time"`   // 结束执行的时间
}

// JobDataSource 任务数据来源
type JobDataSource struct {
	Batch string `json:"batch"` // 批次
	Count int    `json:"count"` // 条数
}
