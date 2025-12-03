package models

import (
	"github.com/astaxie/beego/orm"
	"github.com/rouroumaibing/go-devops/pkg/apis/leaders"
)

func init() {
	orm.RegisterModel(new(leaders.LeaderElection))
}
