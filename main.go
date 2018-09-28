package main

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	_ "golang.garena.com/cow/bs-ops/routers"
)

func main() {
	logs.SetLogger(logs.AdapterFile, `{"filename":"logs/running.log"}`)
	beego.Run()
}
