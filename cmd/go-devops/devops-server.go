package main

import (
	"github.com/rouroumaibing/go-devops/cmd/go-devops/app"
	_ "github.com/rouroumaibing/go-devops/cmd/go-devops/app"
	_ "github.com/rouroumaibing/go-devops/pkg/routers"
)

func main() {
	app.RunInit()
}
