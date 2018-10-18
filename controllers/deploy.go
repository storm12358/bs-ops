package controllers

import (
	"bufio"
	"bytes"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/astaxie/beego/logs"
	"github.com/fsnotify/fsnotify"

	"github.com/astaxie/beego"
	"golang.garena.com/cow/bs-ops/models"
)

type DeployController struct {
	beego.Controller
}

type RespJson struct {
	Message string `json:"message"`
}

func (this *DeployController) Get() {
	this.TplName = "deploy/main.html"
}

func (this *DeployController) Action() {

	action_type := this.GetString("type", "")
	var exec_result RespJson
	switch action_type {
	case "show_stats":
		exec_result = this.showStats()
	case "source_sync":
		exec_result = this.sourceSync()
	case "rebuild_gs":
		exec_result = this.rebuildGS()
	case "restart_gs":
		exec_result = this.restartGS()
	}

	this.Data["json"] = exec_result
	this.ServeJSON()
}

func (this *DeployController) showStats() RespJson {
	shell := "ps -ef|grep gameserver|grep -v grep"
	out := exec_shell(shell)
	if out == "" {
		out = "Server is not running"
	}
	var responseJson RespJson
	responseJson = RespJson{
		Message: out,
	}
	return responseJson
}
func (this *DeployController) sourceSync() RespJson {
	shell := "echo 'Jingle@100'|p4 login && p4 sync //msgame/dev/Server/..."
	out := exec_shell(shell)
	var responseJson RespJson
	responseJson = RespJson{
		Message: out,
	}
	return responseJson
}
func (this *DeployController) rebuildGS() RespJson {
	dist := beego.AppConfig.String("gameserverdir") + "/gameserver"
	cmd := exec.Command("go", "build", "-o", dist, "golang.garena.com/cow/gameserver/cmd/gameserver")
	// cmd := exec.Command("go", "build", "golang.garena.com/cow/gameserver/cmd/gameserver")

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	var str string
	if err != nil {
		str = out.String()
	} else {
		str = "Build success"
	}
	var responseJson RespJson
	responseJson = RespJson{
		Message: str,
	}
	return responseJson
}

func (this *DeployController) restartGS() RespJson {
	var cmd *exec.Cmd
	cmd = models.GetGloableCmdList()
	if cmd != nil {
		cmd.Process.Kill()
	}
	shell := beego.AppConfig.String("gameserverdir") + "/gameserver"
	args := []string{"--config-path", beego.AppConfig.String("gameservercfg")}
	cmd = exec.Command(shell, args...)
	cmd.Dir = beego.AppConfig.String("gameserverdir")

	//
	cmdReader, outErr := cmd.StdoutPipe()
	if outErr != nil {
		logs.Error(os.Stderr, "Error cmd StdoutPipe", outErr)
	}
	scanner := bufio.NewScanner(cmdReader)

	logs.Info("Running shell ", shell, args)
	err := cmd.Start()
	if err != nil {
		logs.Error(os.Stderr, "Error starting Cmd", err)
		os.Exit(1)
	}

	go func() {
		for scanner.Scan() {
			logs.Info("GameServer:", scanner.Text())
		}
	}()

	go func() {
		cmd.Wait()
	}()
	models.SetGloableCmdList(cmd)

	return this.showStats()
}

func (this *DeployController) DownloadLog() {
	this.Ctx.Output.Download(beego.AppConfig.String("gameserverdir")+"/logs/ms-gameserver.log", "server.log")
}

func exec_shell(s string) string {
	cmd := exec.Command("/bin/bash", "-c", s)
	var out bytes.Buffer

	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		// return err.Error()
	}
	return out.String()
}
func (this *DeployController) FolderWatcher() {
	logs.Info("####Watching")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// check gameserver changed
				fileName := filepath.Base(event.Name)
				if fileName != "gameserver" {
					continue
				}
				logs.Info("#### Change Event", event.Op)
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					logs.Info("#### modified file:", event.Name)
					this.sourceSync()
					this.restartGS()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logs.Info("error:", err)
			}
		}
	}()

	dir := beego.AppConfig.String("gameserverdir")
	err = watcher.Add(dir)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}
