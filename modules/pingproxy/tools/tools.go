package tools

import (
	"io/ioutil"
	"os/exec"
	"strings"
)

func ToString(filePath string) (string, error) {
	b, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func ToTrimString(filePath string) (string, error) {
	str, err := ToString(filePath)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(str), nil
}

func PingStatus(ip string) bool {
	cmd := exec.Command("ping", ip, "-c", "2", "-W", "2")
	err := cmd.Run()
	if err != nil {
		return false
	}
	return true
}

