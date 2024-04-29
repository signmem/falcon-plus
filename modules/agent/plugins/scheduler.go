package plugins

import (
	"bytes"
	"encoding/json"
	"github.com/signmem/falcon-plus/common/model"
	"github.com/signmem/falcon-plus/modules/agent/g"
	"github.com/toolkits/file"
	"github.com/toolkits/sys"
	"log"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

type PluginScheduler struct {
	Ticker *time.Ticker
	Plugin *Plugin
	Quit   chan struct{}
}

func NewPluginScheduler(p *Plugin) *PluginScheduler {
	scheduler := PluginScheduler{Plugin: p}
	scheduler.Ticker = time.NewTicker(time.Duration(p.Cycle) * time.Second)
	scheduler.Quit = make(chan struct{})
	return &scheduler
}

func (this *PluginScheduler) Schedule() {
	go func() {
		for {
			select {
			case <-this.Ticker.C:
				PluginRun(this.Plugin)
			case <-this.Quit:
				this.Ticker.Stop()
				return
			}
		}
	}()
}

func (this *PluginScheduler) Stop() {
	close(this.Quit)
}

func PluginRun(plugin *Plugin) {

	timeout := plugin.Cycle*1000 - 500
	fpath := filepath.Join(g.Config().Plugin.Dir, plugin.FilePath)

	if !file.IsExist(fpath) {
		log.Println("no such plugin:", fpath)
		return
	}

	debug := g.Config().Debug
	if debug {
		log.Println(fpath, "running...")
	}

	cmd := exec.Command(fpath)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	err := cmd.Start()
	
	if err != nil {
		log.Println("[ERROR] plugin script running error, file path:  ", fpath)
		return
	}

	if debug {
		log.Println("plugin started:", fpath)
	}

	err, isTimeout := sys.CmdRunWithTimeout(cmd, time.Duration(timeout)*time.Millisecond)
	if err != nil {
		return
	}

	errStr := stderr.String()
	if errStr != "" {
		logFile := filepath.Join(g.Config().Plugin.LogDir, plugin.FilePath+".stderr.log")
		if _, err = file.WriteString(logFile, errStr); err != nil {
			log.Printf("[ERROR] write log to %s fail, error: %s\n", logFile, err)
		}
	}

	if isTimeout {
		// has be killed
		if debug {
			log.Println("[INFO] timeout and kill process", fpath, "successfully")
		}

		return
	}

	// exec successfully
	data := stdout.Bytes()
	if len(data) == 0 {
		// skip log empty data
		return
	}

	var metrics []*model.MetricValue
	err = json.Unmarshal(data, &metrics)
	if err != nil {
		// 习惯上，如果直接通过脚本上报至 127.0.0.1:22230
        // 而上报 /v1/push 后，reponse 不可以满足 json.Unmarshal 返回，因此直接忽略就可以
		return
	}

	g.SendToTransfer(metrics)
}
