package main

import (
	_ "github.com/rouroumaibing/go-devops/cmd/go-devops/app"
	_ "github.com/rouroumaibing/go-devops/pkg/routers"
	"github.com/rouroumaibing/go-devops/pkg/utils/db"
	"k8s.io/klog/v2"

	beego "github.com/astaxie/beego"
)

func main() {
	beego.Run()
	defer klog.Flush()
	defer db.RedisClient.Close()
}
