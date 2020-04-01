package routers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"code.safe.molen.com/molen/haoma/greedy/master/common"
	"code.safe.molen.com/molen/haoma/greedy/master/core"
	"code.safe.molen.com/molen/haoma/greedy/master/storage/mongo"

	"github.com/MolenZhang/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
)

// ReportBody 上报数据内容
type ReportBody struct {
	JobID     string      `json:"job_id"`     // 任务id
	JobResult int         `json:"job_result"` // 任务结果 1-成功 2-失败
	JobDetail []JobDetail `json:"job_detail"` // 任务详情
}

// JobDetail .
type JobDetail struct {
	Phone  string `json:"phone"`
	Result string `json:"crawl_result"`
	Type   int    `json:"crawl_type"`   // 1-模拟器;2-云手机;3-api
	Source int    `json:"crawl_source"` // 1-360手机卫士;2-搜狗号码通;3-电话邦
	Batch  string `json:"crawl_batch"`  // 当前批次
}

// Report 数据上报
func Report(c *gin.Context) {
	res := common.Result{}
	h := &common.RespHeader{
		Version: "1.0.0",
		Time:    time.Now().Unix(),
		Status:  http.StatusOK,
	}
	defer res.SetHeader(h).Done(c)

	body := ReportBody{}
	if err := c.BindJSON(&body); err != nil {
		log.Error("[Report] json unmarshal requst body error:%v", zap.Error(err))
		h.SetStatus(http.StatusBadRequest).SetMsg("Invalid request body")
		return
	}

	job, err := core.JobMgr.JobGetWithID(body.JobID)
	if err != nil {
		log.Errorf("[Report] Get Job With ID: %v, Error:%v", body.JobID, err)
		h.SetStatus(http.StatusBadRequest).SetMsg(err.Error())
		return
	}

	job.Status = int(common.Done)
	job.StopedAt = time.Now().Format("2006-01-02 15:04:05")
	switch body.JobResult {
	case 1:
		job.Result = int(common.Successful)
		// 插入数据库, 如果出错 打印到日志文件
		go func(datas []JobDetail) {
			var docs = []interface{}{}
			for _, data := range datas {
				doc := mongo.CrawlResult{
					ID:        bson.NewObjectId(),
					Phone:     data.Phone,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Source:    data.Source,
					Result:    data.Result,
					Batch:     data.Batch,
					JobID:     job.ID,
					Type:      data.Type,
				}
				docs = append(docs, doc)
			}
			if err := mongo.CrawlResultInsert(docs); err != nil {
				records, _ := json.Marshal(datas)
				log.Error("[Report] data report error", zap.Error(err),
					zap.String("phones", string(records)))
			}
			return
		}(body.JobDetail)
	default:
		// 任务失败时 status=3/执行完成 result=2/失败
		job.Result = int(common.Failed)
	}

	oldJob, err := core.JobMgr.JobSave(job)
	if err != nil {
		log.Errorf("[Report] Update Status Of Job: %#v, With Error: %v", job, err)
		h.SetStatus(http.StatusInternalServerError).SetMsg(err.Error())
		return
	}
	log.Debugf("[Report] Before Update, Job is: %v", oldJob)
	res.SetData(gin.H{"result": "success"})
}

// FilterPool 条件筛选池
// 适配通用平台 所有的条件均在filter参数中 以key-value形式存在
type FilterPool struct {
	JobID string `json:"job_id"` // 任务id
	Batch string `json:"batch"`  // 批次
	Start int    `json:"start"`  // 数据开始的位置
	End   int    `json:"end"`    // 数据结束的位置
}

// Limit 数据起始位置
type Limit struct {
	Start int
	End   int
}

// 处理条件参数
func parseParams(param string, params map[string]string) string {
	if v, ok := params[param]; ok {
		return v
	}
	return ""
}

// DetailShow 结果数据详情
type DetailShow struct {
	ID          string `json:"id"`
	Phone       string `json:"phone"`        // 号码
	CrawlResult string `json:"crawl_result"` // 结果
	CrawlSource string `json:"crawl_source"` // 结果来源 1-360手机卫士;2-搜狗号码通;3-电话邦
	CrawlType   string `json:"crawl_type"`   // 执行方式 1-模拟器;2-云手机;3-api
}

// Show 数据展示
// 根据任务ID 或者 批次进行查看
func Show(c *gin.Context) {
	res := common.ResultJob{
		Status: http.StatusOK,
	}
	defer res.Done(c)

	fMap := map[string]string{}
	l := Limit{Start: 0, End: 999999}
	rge := []int{}
	pRange := c.Query("range")
	if len(rge) != 0 {
		if err := json.Unmarshal([]byte(pRange), &rge); err != nil {
			log.Error("[Show] Json Unmarshal range Error", zap.Error(err))
			res.SetStatus(http.StatusBadRequest).ErrMsg.SetMsg("range is invalid")
			return
		}
		if len(rge) == 2 {
			l = Limit{Start: rge[0], End: rge[1]}
		}
	}

	fs := c.Query("filter")
	if len(fs) == 0 {
		log.Error("[Show] Parameter of filter is required")
		res.SetStatus(http.StatusBadRequest).ErrMsg.SetMsg("filter is required")
		return
	}

	if err := json.Unmarshal([]byte(fs), &fMap); err != nil {
		log.Error("[Show] Json Unmarshal filter Error", zap.Error(err))
		res.SetStatus(http.StatusBadRequest).ErrMsg.SetMsg("filter is invalid")
		return
	}

	filter := FilterPool{
		JobID: parseParams("job_id", fMap),
		Batch: parseParams("batch", fMap),
		Start: l.Start,
		End:   l.End,
	}

	log.Debugf("[Show] Data Show With JobID: %v", filter.JobID)
	if len(filter.JobID) == 0 {
		log.Error("[Show] JobID is required")
		res.SetStatus(http.StatusBadRequest).ErrMsg.SetMsg("job_id is required")
		return
	}

	resp := []DetailShow{}
	// mongo中获取结果
	selecttor := map[string]string{
		"job_id": filter.JobID,
	}

	ds, err := mongo.CrawlResultGet(selecttor)
	if err != nil {
		log.Error("[Show] Get Result From Mongo Error", zap.Error(err))
		res.SetStatus(http.StatusInternalServerError).ErrMsg.SetMsg("Get Data Error")
		return
	}

	for i, d := range ds {
		if (i <= filter.End) && (i >= filter.Start) {
			phone := DetailShow{
				ID:          d.ID.Hex(),
				CrawlResult: d.Result,
				CrawlSource: parseCrawlSource(d.Source),
				CrawlType:   parseCrawlType(d.Type),
				Phone:       d.Phone,
			}
			resp = append(resp, phone)
		}
	}
	c.Header("Content-Range", fmt.Sprintf("posts %v-%v/%v", filter.Start, filter.End, len(ds)))
	res.SetData(resp)
}

// 1-模拟器;2-云手机;3-api
func parseCrawlType(t int) string {
	if v, ok := common.ParseCrawlType[t]; ok {
		return v
	}
	return ""
}

// 1-360手机卫士;2-搜狗号码通;3-电话邦
func parseCrawlSource(s int) string {
	if v, ok := common.ParseCrawlSource[s]; ok {
		return v
	}
	return ""
}
