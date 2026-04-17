package db

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/signmem/falcon-plus/modules/hbs/g"
	"log"
	"time"
)

var DB *sql.DB

func Init() {
	var err error
	DB, err = sql.Open("mysql", g.Config().Database)
	if err != nil {
		log.Fatalln("open db fail:", err)
	}

	DB.SetMaxIdleConns(g.Config().MaxIdle)
	DB.SetConnMaxIdleTime(time.Duration(g.Config().IdleTime) * time.Second)

	err = DB.Ping()
	if err != nil {
		log.Fatalln("ping db fail:", err)
	}
}
