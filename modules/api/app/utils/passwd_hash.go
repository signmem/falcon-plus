package utils

import (
	"github.com/signmem/falcon-plus/modules/api/config"
	"github.com/spf13/viper"
	"github.com/toolkits/str"
)

func HashIt(passwd string) (hashed string) {
	salt := viper.GetString("salt")
	if salt == "" {
		config.Logger.Error("salt is empty, please check your conf")
	}
	hashed = str.Md5Encode(salt + passwd)
	return
}
