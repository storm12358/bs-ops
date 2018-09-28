package routers

import (
	"github.com/astaxie/beego"
	"golang.garena.com/cow/bs-ops/controllers"
)

func init() {
	beego.Router("/", &controllers.DeployController{})
	beego.Router("/deploy", &controllers.DeployController{})
	beego.Router("/deploy/action", &controllers.DeployController{}, "*:Action")
	beego.Router("/deploy/downloadLog", &controllers.DeployController{}, "*:DownloadLog")
}
