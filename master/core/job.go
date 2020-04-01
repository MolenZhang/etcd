package core

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"code.safe.molen.com/molen/haoma/greedy/master/common"
	etcd "code.safe.molen.com/molen/haoma/greedy/master/storage/etcd"

	"github.com/MolenZhang/log"
	"github.com/coreos/etcd/clientv3"
)

// JobManager 任务管理器
type JobManager struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

// JobMgr .
var JobMgr JobManager

// InitJobManager 初始化任务管理器
func InitJobManager(endpoints []string) error {
	config := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: time.Duration(10) * time.Second,
	}
	client, err := clientv3.New(config)
	if err != nil {
		return err
	}
	JobMgr = JobManager{
		client: client,
		kv:     clientv3.NewKV(client),
		lease:  clientv3.NewLease(client),
	}
	return nil
}

// JobSave .
// 任务保存至/greedy/jobs/name
func (JobMgr *JobManager) JobSave(job etcd.JobEtcd) (oldJob etcd.JobEtcd, err error) {
	key := fmt.Sprintf("%s%s", common.JobSavePrefix, job.Name)
	value, _ := json.Marshal(job)
	resp, err := JobMgr.kv.Put(context.TODO(), key, string(value), clientv3.WithPrevKV())
	if err != nil {
		return
	}
	// 更新操作的话 会返回旧的任务详情
	if resp.PrevKv != nil {
		oldJob = etcd.JobEtcd{}
		err = json.Unmarshal(resp.PrevKv.Value, &oldJob)
		if err != nil {
			return
		}
	}
	return
}

// JobDelete 任务删除
func (JobMgr *JobManager) JobDelete(name string) (oldJob etcd.JobEtcd, err error) {
	key := fmt.Sprintf("%s%s", common.JobSavePrefix, name)
	resp, err := JobMgr.kv.Delete(context.TODO(), key, clientv3.WithPrevKV())
	if err != nil {
		return
	}

	if resp.PrevKvs != nil {
		oldJob = etcd.JobEtcd{}
		err = json.Unmarshal(resp.PrevKvs[0].Value, &oldJob)
		if err != nil {
			return
		}
	}
	return
}

// JobLists 任务清单
func (JobMgr *JobManager) JobLists() (jobs []etcd.JobEtcd, err error) {
	resp, err := JobMgr.kv.Get(context.TODO(), common.JobSavePrefix, clientv3.WithPrefix())
	if err != nil {
		return
	}
	jobs = []etcd.JobEtcd{}
	for _, kv := range resp.Kvs {
		job := etcd.JobEtcd{}
		err := json.Unmarshal(kv.Value, &job)
		if err != nil {
			continue
		}
		jobs = append(jobs, job)
	}
	return
}

// JobKill 任务强杀
// TODO make sure
func (JobMgr *JobManager) JobKill(name string) (err error) {
	key := fmt.Sprintf("%s%s", common.JobKillPrefix, name)
	resp, err := JobMgr.lease.Grant(context.TODO(), 1)
	if err != nil {
		return
	}
	leaseID := resp.ID
	_, err = JobMgr.kv.Put(context.TODO(), key, "", clientv3.WithLease(leaseID))
	return
}

// ParseToString .
func ParseToString(src interface{}) (string, error) {
	drs, err := json.Marshal(src)
	return string(drs), err
}

// JobMatch 获取可执行的任务
func (JobMgr *JobManager) JobMatch(crawlType int) (job etcd.JobEtcd, err error) {
	jobs, err := JobMgr.JobLists()
	if err != nil || len(jobs) == 0 {
		log.Errorf("[JobMatch] No Job or Get Jobs error: %v", err)
		return
	}
	// 顺序随机
	jobs, err = JobMgr.FisherYates(jobs)
	for _, v := range jobs {
		// 校验是否符合当前执行方式
		if v.CrawlType != crawlType {
			continue
		}
		// 校验状态是在执行中
		if v.Status == 2 {
			continue
		}
		// 校验状态是否为执行完成且执行结果为执行成功
		if v.Status == 3 && v.Result == 1 {
			continue
		}
		job = v
		return
	}
	err = fmt.Errorf("No Job Matched")
	return
}

// JobLockMade 给任务造锁
func (JobMgr *JobManager) JobLockMade(name string) *JobLock {
	return NewJobLock(name, JobMgr.kv, JobMgr.lease)
}

// FisherYates 任务打乱
func (JobMgr *JobManager) FisherYates(src []etcd.JobEtcd) ([]etcd.JobEtcd, error) {
	if len(src) == 0 {
		return nil, fmt.Errorf("Invalid parameters")
	}
	t := len(src)
	rand.Seed(time.Now().Unix())
	for i := t - 1; i > 0; i-- {
		num := rand.Intn(i + 1)
		src[i], src[num] = src[num], src[i]
	}
	return src, nil
}

// JobGetWithID 根据id获取job详情
func (JobMgr *JobManager) JobGetWithID(id string) (job etcd.JobEtcd, err error) {
	jobs, err := JobMgr.JobLists()
	if err != nil {
		return
	}
	for _, v := range jobs {
		if v.ID == id {
			job = v
			return
		}
	}
	err = fmt.Errorf("No Job Matched")
	return
}
