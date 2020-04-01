package routers

import (
	"fmt"
	"net/http"
	"time"

	"code.safe.molen.com/molen/haoma/greedy/master/common"
	"code.safe.molen.com/molen/haoma/greedy/master/core"
	etcd "code.safe.molen.com/molen/haoma/greedy/master/storage/etcd"
	"code.safe.molen.com/molen/haoma/greedy/master/storage/mongo"
	mysql "code.safe.molen.com/molen/haoma/greedy/master/storage/mysql"

	"github.com/MolenZhang/log"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PhonesResp 号码列表返回
type PhonesResp struct {
	Phones []string `json:"phone"`
	Batch  string   `json:"batch"`
	JobID  string   `json:"job_id"`
	Policy struct {
		Num      int `json:"num"`
		Interval int `json:"interval"`
	} `json:"policy"`
}

// SourceResp 返回待执行的数据
type SourceResp struct {
	Resp PhonesResp
}

// Phones 获取号码列表
func Phones(c *gin.Context) {
	res := common.Result{}
	h := &common.RespHeader{
		Version: "1.0.0",
		Time:    time.Now().Unix(),
		Status:  http.StatusOK,
	}
	resp := &PhonesResp{
		Phones: []string{},
	}
	defer res.SetHeader(h).SetData(resp).Done(c)

	// 执行方式
	crawlType := c.Query("act_type")
	if len(crawlType) == 0 {
		h.SetStatus(http.StatusBadRequest).SetMsg("Invalid Parameter of act_type")
		return
	}
	// 运行环境
	envType := c.Query("env_type")
	if len(envType) == 0 {
		h.SetStatus(http.StatusBadRequest).SetMsg("Invalid Parameter of env_type")
		return
	}
	log.Debug("[Phones] Parameters",
		zap.String("act_type", crawlType),
		zap.String("env_type", envType))

	switch envType {
	case common.CrawlTypeSimulator:
	case common.CrawlTypeCloudPhone:
		if err := resp.SuitCloudPhone(); err != nil {
			h.SetStatus(http.StatusBadRequest).SetMsg(err.Error())
		}
	case common.CrawlTypeAgent:
	default:
		h.SetStatus(http.StatusBadRequest).SetMsg("Invalid Parameter of act_type")
	}
}

// SuitCloudPhone 适配云手机
func (s *PhonesResp) SuitCloudPhone() (err error) {
	// Mock TODO delete
	/*if true {
		s.Phones = append(s.Phones, "15330091234", "15757121234")
		s.Batch = "20200218"
		s.JobID = "1234-asdf-4567"
		return
	}*/

	// 获取任务 任务加锁
	job, err := core.JobMgr.JobMatch(2)
	if err != nil {
		return
	}
	log.Debugf("[SuitCloudPhone] JobMatched Is: %v", job)
	// 抢锁
	jobLock := core.JobMgr.JobLockMade(job.Name)
	err = jobLock.TryLock()
	if err != nil {
		log.Error("[SuitCloudPhone] JobLock Made Error", zap.Error(err))
		return
	}
	defer jobLock.UnLock()
	log.Debugf("[SuitCloudPhone] TryLock SUCC")

	// 加锁成功后获取数据
	// 当数据获取成功后 修改任务执行状态
	err = s.CrawlSource(job)
	if err != nil {
		log.Errorf("[SuitCloudPhone] CrawlSource Error: %v", err)
		return
	}
	s.Batch = job.Source.Batch
	s.JobID = job.ID

	job.Status = int(common.Running)
	job.StartedAt = time.Now().Format("2006-01-02 15:04:05")
	curJob, err := core.JobMgr.JobSave(job)
	if err != nil {
		log.Error("[SuitCloudPhone] Update Job Status Error", zap.Error(err))
		return
	}
	log.Infof("[SuitCloudPhone] Job Matched is: %#v", curJob)

	return
}

// CrawlSource .
// 根据任务的状态选取数据来源
// 如果是首次 先从mongo中获取(此处从mongo中是因为担心 任务执行时 mysql中还未抓取到待执行数据)
// 如果是非首次 从mysql中获取(此时 mysql中应该有该任务待执行的数据)
func (s *PhonesResp) CrawlSource(job etcd.JobEtcd) (err error) {
	switch common.JobResult(job.Result) {
	case common.Failed:
		dsMysql := []mysql.JobDatum{}
		// 读取mysql 非首次次如果能命中该任务则说明该任务曾经执行失败过一次 再次命中时mysql中已经有任务相关的数据
		dsMysql, err = mysql.BatchFetch(job.ID)
		if err != nil {
			log.Error("[CrawlSource] Get CrawlSource Data Error", zap.Error(err))
			return
		}
		if len(dsMysql) == 0 {
			err = fmt.Errorf("No Data Matched")
			log.Errorf("[CrawlSource] No Data Matched From Mongo With Job: %v", job)
			return
		}
		for _, d := range dsMysql {
			s.Phones = append(s.Phones, d.Phone.String)
		}
	default:
		// 读取mongo 首次命中 先从mongo中获取数据
		docs := []mongo.CrawlSource{}
		docs, err = mongo.CrawlSourceGet(job.Source.Batch, job.Source.Count)
		if err != nil {
			log.Error("[CrawlSource] Get CrawlSource Data Error", zap.Error(err))
			return
		}
		if len(docs) == 0 {
			err = fmt.Errorf("No Data Matched")
			log.Errorf("[CrawlSource] No Data Matched From Mongo With Job: %v", job)
			return
		}
		for _, doc := range docs {
			s.Phones = append(s.Phones, doc.Phone)
		}
	}
	return
}
