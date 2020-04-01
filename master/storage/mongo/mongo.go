package mongo

import (
	"fmt"
	"sync"
	"time"

	"code.safe.molen.com/molen/haoma/greedy/master/config"
	"github.com/MolenZhang/log"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// table
const (
	CollectCrawlSource string = "hm_greedy_source"
	CollectCrawlResult string = "hm_greedy_result"
	database           string = "hm_online"
)

var session *mgo.Session
var once sync.Once

// Client .
type Client struct {
	Conn *mgo.Session
}

// InitMongo init mongodb
func InitMongo(cfg config.MongoConfig) (err error) {
	once.Do(func() {
		mgoInfo := &mgo.DialInfo{
			Addrs:  cfg.Addrs,
			Direct: false,

			Timeout:   time.Second * 60,
			Database:  database,
			Username:  cfg.User,
			Password:  cfg.Pwd,
			PoolLimit: 1024,
		}
		s, err := mgo.DialWithInfo(mgoInfo)
		if err != nil {
			return
		}
		s.SetMode(mgo.Monotonic, true) // 分散一些读操作到其他服务器 但是未必读到最新的数据
		session = s
	})
	return
}

// PubCollection 获取collection
func PubCollection(collection string, f func(*mgo.Collection) error) error {
	c := NewClient()
	defer c.Conn.Close()
	return f(c.Conn.DB(database).C(collection))
}

// NewClient .
func NewClient() *Client {
	return &Client{
		Conn: session.Clone(),
	}
}

// CrawlSource 数据源
type CrawlSource struct {
	ID        bson.ObjectId `bson:"_id"`
	Phone     string        `bson:"phone"`
	DataType  int           `bson:"data_type"`   // 0：抓取补充数据 1：用于评估对标数据（百度有结果的数据，同时更新同步到结果表）
	CreatedAt time.Time     `bson:"create_time"` // 创建时间
	UpdatedAt time.Time     `bson:"update_time"` // 修改时间
	IsVerify  int           `bson:"is_verify"`   // 是否获取（抓取）过数据 0-否 1-是
	Batch     string        `bson:"batch"`       // 批次
}

// CrawlResult 结果存储
type CrawlResult struct {
	ID        bson.ObjectId `bson:"_id"`
	Phone     string        `bson:"phone"`
	CreatedAt time.Time     `bson:"create_time"` // 创建时间
	UpdatedAt time.Time     `bson:"update_time"` // 修改时间
	Name      string        `bson:"name"`
	Tag       string        `bson:"tag"`
	Category  string        `bson:"category"`
	Source    int           `bson:"crawl_source"` // 1-360手机卫士;2-搜狗号码通;3-电话邦
	Type      int           `bson:"crawl_type"`   // 1-模拟器;2-云手机;3-api
	Result    string        `bson:"crawl_result"` // 抓取结果 json 字符串即可
	Batch     string        `bson:"crawl_batch"`  // 数据批次 20200218
	JobID     string        `bson:"job_id"`       // 任务id
}

// CrawlResultInsert 上报的结果数据入库
func CrawlResultInsert(docs []interface{}) error {
	insert := func(c *mgo.Collection) error {
		return c.Insert(docs...)
	}
	return PubCollection(CollectCrawlResult, insert)
}

// CrawlSourceGet 获取号码列表
func CrawlSourceGet(batch string, limit int) (docs []CrawlSource, err error) {
	query := func(c *mgo.Collection) error {
		return c.Find(bson.M{"batch": batch}).Sort("_id").Limit(limit).All(&docs)
	}
	if err := PubCollection(CollectCrawlSource, query); err != nil {
		log.Error("[PhoneList] Get Phones From Mongo Error", zap.Error(err))
	}
	return
}

// CrawlSourceUpdate 批量更新已抓取的数据状态
func CrawlSourceUpdate(phone string, status int) error {
	selector := bson.M{"phone": phone}
	data := bson.M{"$set": bson.M{"is_verify": status}}
	update := func(c *mgo.Collection) error {
		return c.Update(selector, data)
	}
	return PubCollection(CollectCrawlSource, update)
}

// CrawlResultGet 结果获取
func CrawlResultGet(fs map[string]string) (docs []CrawlResult, err error) {
	if len(fs) == 0 {
		err = fmt.Errorf("Bad Params")
		return
	}

	selector := bson.M{}

	if v, ok := fs["batch"]; ok {
		selector["batch"] = v
	}
	if v, ok := fs["job_id"]; ok {
		selector["job_id"] = v
	}
	if len(selector) == 0 {
		err = fmt.Errorf("Bad Params")
		return
	}

	query := func(c *mgo.Collection) error {
		return c.Find(selector).Sort("_id").All(&docs)
	}
	if err := PubCollection(CollectCrawlResult, query); err != nil {
		log.Error("[PhoneList] Get Result From Mongo Error", zap.Error(err))
	}
	return
}

// GetCount 获取数据总数
func GetCount(collection string, fs map[string]string) (count int, err error) {
	selector := bson.M{}
	if v, ok := fs["batch"]; ok {
		selector["batch"] = v
	}
	if v, ok := fs["job_id"]; ok {
		selector["job_id"] = v
	}
	if len(selector) == 0 {
		err = fmt.Errorf("Bad Params")
		return
	}

	cnt := func(c *mgo.Collection) error {
		count, err = c.Find(selector).Count()
		return err
	}
	if err := PubCollection(collection, cnt); err != nil {
		log.Error("[CetCount] Get Count Error", zap.Error(err))
	}
	return
}
