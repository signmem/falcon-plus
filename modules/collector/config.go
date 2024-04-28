package main

import (
	"log"
	"github.com/spf13/viper"
	"gitlab.tools.vipshop.com/vip-ops-sh/falcon-stats-collector/g"
)

var Config *viper.Viper

func InitConfig(env string) {
	v := viper.New()
	v.SetConfigType(g.Config().Env.Type)
	v.SetConfigName(g.Config().Env.Name)
	v.AddConfigPath(g.Config().Env.Path)

	err := v.ReadInConfig()
	if err != nil {
		log.Fatal("error on parsing configuration file", err)
	}
	Config = v
}
