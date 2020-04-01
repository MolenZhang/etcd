package common

// .
const (
	JobSavePrefix = "/greedy/jobs/save"
	JobKillPrefix = "/greedy/jobs/kill"
	JobLockPrefix = "/greedy/jobs/lock"
)

// .
const (
	CrawlTypeSimulator  = "1"
	CrawlTypeCloudPhone = "2"
	CrawlTypeAgent      = "3"
)

// JobStatus 定义任务状态
type JobStatus int

// job status
const (
	Pendding JobStatus = iota + 1 // 未执行
	Running                       // 执行中
	Done                          // 执行完成
)

// JobResult 任务执行结果
type JobResult int

// job result
const (
	Successful JobResult = iota + 1 // 成功
	Failed                          // 失败
)

// CrawlType 执行方式
type CrawlType int

// crawl type
const (
	Simulator  CrawlType = iota + 1 // 模拟器
	CloudPhone                      // 云手机
	Agent                           // 破解api
)

// 转换抓取来源
var (
	ParseCrawlSource = map[int]string{
		1: "360手机卫士",
		2: "搜狗号码通",
		3: "电话邦",
	}
)

// 转换抓取方式
var (
	ParseCrawlType = map[int]string{
		1: "模拟器",
		2: "云手机",
		3: "Api",
	}
)
