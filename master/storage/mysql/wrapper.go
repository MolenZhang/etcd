package storage

import (
	"context"

	"github.com/MolenZhang/log"
	"github.com/volatiletech/sqlboiler/boil"
	. "github.com/volatiletech/sqlboiler/queries/qm"
	"go.uber.org/zap"
)

// BatchInsert 批量插入
// TODO 数据量大的时候换一种方式
func BatchInsert(ds []JobDatum) {
	for _, d := range ds {
		if err := d.Insert(context.TODO(), boil.GetContextDB(), boil.Infer()); err != nil {
			log.Error("[BatchInsert] Insert Error", zap.Any("data", d), zap.Any("error", err))
			continue
		}
	}
}

// BatchFetch 获取数据
// 获取 任务 相关的数据
func BatchFetch(jobID string) (ds []JobDatum, err error) {
	ds = []JobDatum{}
	q := NewQuery(Where("job_id = ?", jobID))
	q.Bind(context.TODO(), boil.GetContextDB(), &ds)
	return
}
