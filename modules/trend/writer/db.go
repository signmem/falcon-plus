package writer

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/signmem/falcon-plus/modules/trend/g"
)

var DB *sql.DB

func initDB() {
	var err error
	DB, err = sql.Open("mysql", g.Config().DB.Dsn)
	if err != nil {
		g.Logger.Errorf("open db fail:", err)
	}

	DB.SetMaxIdleConns(g.Config().DB.MaxIdle)

	err = DB.Ping()
	if err != nil {
		g.Logger.Errorf("ping db fail:", err)
	}
}

func closeDB() {
	if DB != nil {
		DB.Close()
	}
}
