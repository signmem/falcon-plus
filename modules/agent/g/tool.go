package g

import (
	"bytes"
	"fmt"
	"github.com/toolkits/file"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

var (
	VersionDir string
	VersionFile string
)

func ForceResetVersion(versionFile string) string {

	hour :=  JobTime.Hour
	minite := JobTime.Minite
	now := GetNow()
	timestring := []byte(now.Hour + now.Minite)

	ioutil.WriteFile(versionFile, timestring, 0644)
	versionString := hour + "-" + minite + "@" + now.Hour + now.Minite
	return versionString

}

func IsDir(path string) bool  {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func IsFile(path string) bool {

	return !IsDir(path)
}

func CreateVersionFile(dir string, file string) bool {
	if IsExist(file) == false {
		if IsExist(dir) == false {
			os.Mkdir(dir, 0755)
			os.Create(file)
		} else {
			if IsDir(dir) == true {
				os.Create(file)
			} else if IsFile(dir) {
				err := os.Remove(dir)
				if err != nil {
					log.Printf("[ERROR] can not remove dir %s", dir)
				}
				os.Mkdir(dir, 0755)
				os.Create(file)
			}
		}
	}

	if IsDir(file) == true {
		os.Remove(file)
		os.Create(file)
	}
	return true
}

func GetPluginVersion() string {

	hour :=  JobTime.Hour
	minite := JobTime.Minite

	if ! Config().Plugin.Enabled {
		return "plugin_config_not_enabled"
	}


	if ! IsExist(VersionFile) {
		_ = CreateVersionFile(VersionDir, VersionFile)
		versionString := ForceResetVersion(VersionFile)
		return versionString
	}

	versionF, _ := os.Stat(VersionFile)
	if versionF.Size() == 0 {
		versionString := ForceResetVersion(VersionFile)
		return versionString
	}

	version, err := ioutil.ReadFile(VersionFile)
	if err != nil {
		versionString := ForceResetVersion(VersionFile)
		return versionString
	}

	versionFileString := strings.TrimSpace(string(version))
	versionString := hour + "-" + minite + "@" + versionFileString

	return versionString
}

func UpdatePluginVersion( versionString string ) bool {


	if ! IsExist(VersionFile) {
		_ = CreateVersionFile(VersionDir, VersionFile)
	}

	versionByte := []byte(versionString)

	err := ioutil.WriteFile(VersionFile, versionByte ,0644)
	if err != nil {
		log.Println("UpdatePluginVersion() error can not write into version file")
		return false
	}

	return true


}

func GetVersionFileInfo() string {

	if ! IsExist(VersionFile) {
		_ = CreateVersionFile(VersionDir, VersionFile)
		return ""
	}

	version, err := ioutil.ReadFile(VersionFile)
	if err != nil {
		versionInfo := "can not read version file info from " +  VersionFile + "\n"
		return versionInfo
	}

	versionData := strings.TrimSpace(string(version))
	if len(versionData) == 0 {
		versionInfo := "plugin version file " + VersionFile + " is empty\n"
		return versionInfo
	}
	versionInfo := "plugin version is " + versionData + "\n"

	return versionInfo
}

func GetCurrPluginVersion() string {
	if !Config().Plugin.Enabled {
		return "plugin not enabled"
	}

	pluginDir := Config().Plugin.Dir
	if !file.IsExist(pluginDir) {
		return "plugin dir not existent"
	}

	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = pluginDir

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return fmt.Sprintf("Error:%s", err.Error())
	}

	return strings.TrimSpace(out.String())
}

func ChownPluginDir(dirString string) bool {
	cmd := exec.Command("sudo", "chown", "apps:apps", "-R", dirString)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true

}

func CheckOSVersion() ( version string ) {
	cmd := exec.Command("rpm", "--eval","%{centos_ver}")
	var out bytes.Buffer
	cmd.Stdout =  &out
	err := cmd.Run()
	if err != nil {
		version = "not_known"
		return
	}

	version = strings.Replace(out.String(),"\n","", -1)
	return version
}

func CheckUrlTcpHealth(hostUrl string) bool {
	port := "80"
	timeout := time.Second

	urlReg := regexp.MustCompile("http://")
	hostPost := urlReg.ReplaceAllString(hostUrl, "")
	host := strings.Split(string(hostPost), "/")
	_, err := net.DialTimeout("tcp", host[0] + ":" + port, timeout)
	if err != nil {
		return false
	}
	return true
}