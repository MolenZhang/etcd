package core

import (
	"context"
	"errors"

	"code.safe.molen.com/molen/haoma/greedy/master/common"
	"github.com/MolenZhang/log"
	"github.com/coreos/etcd/clientv3"
	"go.uber.org/zap"
)

// JobLock 构建分布式锁
type JobLock struct {
	kv         clientv3.KV
	lease      clientv3.Lease
	jobName    string
	cancelFunc context.CancelFunc
	leaseID    clientv3.LeaseID
	isLocked   bool
}

// NewJobLock 初始化锁
func NewJobLock(jobName string, kv clientv3.KV, lease clientv3.Lease) *JobLock {
	return &JobLock{
		kv:      kv,
		lease:   lease,
		jobName: jobName,
	}
}

// TryLock 加锁
func (jobLock *JobLock) TryLock() (err error) {
	resp, err := jobLock.lease.Grant(context.TODO(), 5)
	if err != nil {
		return
	}
	ctx, cancelFunc := context.WithCancel(context.TODO())
	leaseID := resp.ID

	keepRespChan, err := jobLock.lease.KeepAlive(ctx, leaseID)
	if err != nil {
		cancelFunc()
		return
	}

	go func() {
		for {
			select {
			case keepResp := <-keepRespChan:
				if keepResp == nil {
					return
				}
			}
		}
	}()

	// 创建事务txn
	txn := jobLock.kv.Txn(context.TODO())
	lockKey := common.JobLockPrefix + jobLock.jobName

	// 事务抢锁
	txn.If(clientv3.Compare(clientv3.
		CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseID))).
		Else(clientv3.OpGet(lockKey))

	txnResp, err := txn.Commit()
	if err != nil {
		cancelFunc()
		err = errors.New("Lock Failed, Transaction Commit Failed")
		return
	}

	// 成功返回，失败释放租约
	if !txnResp.Succeeded {
		cancelFunc()
		err = errors.New("Lock Failed, Lock Occupied")
		return
	}

	jobLock.leaseID = leaseID
	jobLock.cancelFunc = cancelFunc
	jobLock.isLocked = true
	return
}

// UnLock 解锁
func (jobLock *JobLock) UnLock() {
	if jobLock.isLocked {
		jobLock.cancelFunc() // 取消自动续约
		if _, err := jobLock.lease.Revoke(context.TODO(), jobLock.leaseID); err != nil {
			log.Error("解锁失败", zap.Error(err))
		} // 释放租约
	}
}
