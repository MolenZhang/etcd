package config

// Config 配置相关
type Config struct {
	Etcd  EtcdConfig
	Mongo MongoConfig
	Mysql MysqlConfig
}

// EtcdConfig .
type EtcdConfig struct {
	Addrs []string
}

// MongoConfig .
type MongoConfig struct {
	Addrs []string
	User  string
	Pwd   string
}

// MysqlConfig .
type MysqlConfig struct {
	User     string
	Pwd      string
	Addr     string
	Database string
}

// Cfg .
var Cfg Config

// NewConfig .
func NewConfig() *Config {
	return &Config{
		Etcd: EtcdConfig{
			Addrs: []string{},
		},
		Mongo: MongoConfig{
			Addrs: []string{},
		},
	}
}

// InitConfig .
func InitConfig() {
	cfg := NewConfig()
	cfg.Mongo = MongoConfig{
		Addrs: []string{"10.99.91.62:8378"},
		User:  "hm_w",
		Pwd:   "hm4mongo",
	}
	cfg.Mongo.Addrs = []string{"10.12.5.164:27017", "10.12.5.164:27017"}
	cfg.Mysql = MysqlConfig{
		User:     "rdswr",
		Pwd:      "rdswr2018",
		Addr:     "10.14.122.12",
		Database: "greedy",
	}
	Cfg = *cfg
}
