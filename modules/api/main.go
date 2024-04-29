package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"log"
	"github.com/gin-gonic/gin"
	yaag_gin "github.com/masato25/yaag/gin"
	"github.com/masato25/yaag/yaag"
	"github.com/signmem/falcon-plus/modules/api/app/controller"
	"github.com/signmem/falcon-plus/modules/api/config"
	"github.com/signmem/falcon-plus/modules/api/graph"
	"github.com/signmem/falcon-plus/modules/api/data"
	"github.com/spf13/viper"
)

func initGraph() {
	graph.Start(viper.GetStringMapString("graphs.cluster"))
}

func main() {
	cfgTmp := flag.String("c", "cfg.json", "configuration file")
	version := flag.Bool("v", false, "show version")
	help := flag.Bool("h", false, "help")
	flag.Parse()
	cfg := *cfgTmp
	if *version {
		fmt.Println(config.VERSION)
		os.Exit(0)
	}

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	viper.AddConfigPath(".")
	viper.AddConfigPath("/")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("./api/config")
	cfg = strings.Replace(cfg, ".json", "", 1)
	viper.SetConfigName(cfg)

	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("[ERROR] read config error: %s", err)
	}
	// err = config.InitLog(viper.GetString("log_level"))
	//if err != nil {
	//	log.Fatal(err)
	//}
	err = config.InitDB(viper.GetBool("db.db_bug"))
	if err != nil {
		log.Printf("[ERROR] db conn failed with error %s", err.Error())
	}

	if viper.GetString("log_level") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}
	routes := gin.Default()
	if viper.GetBool("gen_doc") {
		yaag.Init(&yaag.Config{
			On:       true,
			DocTitle: "Gin",
			DocPath:  viper.GetString("gen_doc_path"),
			BaseUrls: map[string]string{"Production": "/api/v1", "Staging": "/api/v1"},
		})
		routes.Use(yaag_gin.Document())
	}

	config.Logfile = viper.GetString("logfile")
	config.LogMaxAge = viper.GetInt("logmaxage")
	config.LogRotateAge = viper.GetInt("logrotateage")
	config.Logger = config.InitLog()
	data.MetricFile  = viper.GetString("metric_list_file")

	initGraph()
	//start gin server
	config.Logger.Debugf("start with port:%v", viper.GetString("web_port"))
	go controller.StartGin(viper.GetString("web_port"), routes)
	go data.CronGenMetric()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println()
		os.Exit(0)
	}()

	select {}
}
