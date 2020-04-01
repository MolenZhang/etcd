package routers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"code.safe.molen.com/molen/haoma/greedy/master/common"
	"code.safe.molen.com/molen/haoma/greedy/master/core"
	etcd "code.safe.molen.com/molen/haoma/greedy/master/storage/etcd"
	"code.safe.molen.com/molen/haoma/greedy/master/storage/mongo"
	mysql "code.safe.molen.com/molen/haoma/greedy/master/storage/mysql"

	"github.com/MolenZhang/log"
	"github.com/gin-gonic/gin"
	"github.com/volatiletech/null"
	"go.uber.org/zap"
)

// WebJob 定义平台任务相关字段
type WebJob struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Batch       string `json:"batch"`
	Count       int    `json:"count"`
	CreatedAt   string `json:"create_time"`
	Status      string `json:"status"` // 1-未执行 2-执行中 3-执行完成
	Description string `json:"desc"`
	CrawlType   string `json:"crawl_type"`   // 执行方式 1-模拟器;2-云手机;3-api
	Result      string `json:"crawl_result"` // 执行结果 1-成功 2-失败
}

// Source 数据源
type Source struct {
	Batch string `json:"batch"` // 批次
	Count int    `json:"count"` // 条数
}

// JobSave 保存任务
func JobSave(c *gin.Context) {
	res := common.ResultJob{
		Status: http.StatusOK,
	}
	defer res.Done(c)

	req := &WebJob{}
	if err := c.BindJSON(req); err != nil {
		log.Errorf("[JobSave] Parse Request Body Error: %v", err)
		res.SetStatus(http.StatusBadRequest).ErrMsg.SetMsg(err.Error())
		return
	}
	ct, err := strconv.Atoi(req.CrawlType)
	if err != nil {
		log.Errorf("[JobSave] Parse CrawlType Error: %v", err)
		res.SetStatus(http.StatusBadRequest).ErrMsg.SetMsg(err.Error())
		return
	}
	job := etcd.JobEtcd{
		ID:          common.GenerateID(),
		Name:        req.Name,
		CrawlType:   ct,
		Description: req.Description,
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
		Status:      int(common.Pendding), /*未执行*/
		Source: etcd.JobDataSource{
			Batch: req.Batch,
			Count: req.Count,
		},
	}
	if _, err := core.JobMgr.JobSave(job); err != nil {
		log.Errorf("[JobSave] JobSave Error: %v", err)
		res.SetStatus(http.StatusInternalServerError).ErrMsg.SetMsg(err.Error())
		return
	}
	// 此时将任务相关的数据从mongo中 导入到mysql中
	// TODO 此处待优化
	go func(job etcd.JobEtcd) {
		ins := []mysql.JobDatum{}
		dsrc, err := mongo.CrawlSourceGet(job.Source.Batch, job.Source.Count)
		if err != nil {
			log.Error("[JobSave] Get Source Data Error", zap.Error(err),
				zap.String("JobName", job.Name), zap.String("JobID", job.ID))
			return
		}
		for _, d := range dsrc {
			jd := mysql.JobDatum{
				Phone:     null.StringFrom(d.Phone),
				Batch:     null.StringFrom(d.Batch),
				JobID:     null.StringFrom(job.ID),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			ins = append(ins, jd)
		}
		mysql.BatchInsert(ins)
	}(job)
	res.SetData(job)
}

// JobDelete 删除任务
func JobDelete(c *gin.Context) {
	res := common.ResultJob{
		Status: http.StatusOK,
	}
	defer res.Done(c)
	id := c.Param("id")
	if len(id) == 0 {
		res.SetStatus(http.StatusBadRequest).ErrMsg.SetMsg("id is required")
		log.Error("[JobDelete] Parameter id is required")
		return
	}

	job, err := core.JobMgr.JobGetWithID(id)
	if err != nil {
		res.SetStatus(http.StatusBadRequest).ErrMsg.SetMsg(err.Error())
		log.Errorf("[JobDelete] No Job Matched With ID: %v", id)
		return
	}

	if _, err := core.JobMgr.JobDelete(job.Name); err != nil {
		log.Error("[JobDelete] Delete Job Error", zap.Error(err))
		res.SetStatus(http.StatusInternalServerError).ErrMsg.SetMsg(err.Error())
		return
	}
	res.SetData(gin.H{"ok": true})
}

// JobLists 任务列表
func JobLists(c *gin.Context) {
	res := common.ResultJob{
		Status: http.StatusOK,
	}
	defer res.Done(c)
	l := Limit{Start: 0, End: 999999}

	rge := []int{}
	pRange := c.Query("range")
	if len(pRange) != 0 {
		if err := json.Unmarshal([]byte(pRange), &rge); err != nil {
			res.SetStatus(http.StatusBadRequest).ErrMsg.SetMsg("range is invalid")
			log.Error("[JobLists] Parameter range is invalid")
			return
		}
		if len(rge) == 2 {
			l = Limit{Start: rge[0], End: rge[1]}
		}
	}

	jobs, err := core.JobMgr.JobLists()
	if err != nil {
		res.SetStatus(http.StatusInternalServerError).ErrMsg.SetMsg(err.Error())
		log.Errorf("[JobLists] Get Job Lists Error: %v", err)
		return
	}

	resp := []WebJob{}
	if len(jobs) != 0 {
		for i, job := range jobs {
			if (i <= l.End) && (i >= l.Start) {
				tmp := WebJob{
					ID:        job.ID,
					Name:      job.Name,
					Batch:     job.Source.Batch,
					Count:     job.Source.Count,
					CreatedAt: job.CreatedAt,
					Status:    strconv.Itoa(job.Status),
					CrawlType: strconv.Itoa(job.CrawlType),
					Result:    strconv.Itoa(job.Result),
				}
				resp = append(resp, tmp)
			}
		}
	}
	c.Header("Content-Range", fmt.Sprintf("posts %v-%v/%v", l.Start, l.End, len(jobs)))
	res.SetData(resp)
}

// JobKill 强杀任务
func JobKill(c *gin.Context) {}

// JobGet 获取单个任务
func JobGet(c *gin.Context) {
	res := common.ResultJob{
		Status: http.StatusOK,
	}
	defer res.Done(c)

	id := c.Param("id")
	if len(id) == 0 {
		res.SetStatus(http.StatusBadRequest).ErrMsg.SetMsg("id is required")
		log.Error("[JobGet] Parameter id is required")
		return
	}

	job, err := core.JobMgr.JobGetWithID(id)
	if err != nil {
		res.SetStatus(http.StatusBadRequest).ErrMsg.SetMsg(err.Error())
		log.Errorf("[JobGet] No Job Matched With ID: %v", id)
		return
	}

	res.SetData(job)
}
