package storage

import (
	"database/sql"
	"fmt"

	"code.safe.molen.com/molen/haoma/greedy/master/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/volatiletech/sqlboiler/boil"
)

// InitMysql .
func InitMysql(cfg config.MysqlConfig) (err error) {
	metadata := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		cfg.User, cfg.Pwd, cfg.Addr, cfg.Database)
	db, err := sql.Open("mysql", metadata)
	if err != nil {
		return
	}
	db.SetMaxIdleConns(100)
	db.SetMaxOpenConns(100)
	boil.SetDB(db)
	return
}
