package options

import (
	"github.com/astaxie/beego"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	cmnOpt "github.com/rouroumaibing/go-devops/cmd/common/options"
	"github.com/rouroumaibing/go-devops/pkg/utils/system"
)

type ServerRunOptions struct {
	cmnOpt.CommonOptions
	BackendServerOptions
}

func (s *ServerRunOptions) AddFlags(fs *pflag.FlagSet) {
	s.CommonOptions.AddFlags(fs)
	s.BackendServerOptions.AddFlags(fs)
}

func NewServerRunOptions() *ServerRunOptions {
	s := ServerRunOptions{}
	ip, err := system.GetLocalIP()
	if err != nil {
		klog.Error("get ip add failed")
		panic(err)
	}
	s.Listen.HTTPSAddr = ip
	s.Listen.HTTPSPort = 2356
	s.Listen.EnableHTTPS = true
	s.Listen.EnableHTTP = false
	s.Listen.EnableAdmin = true
	return &s
}

type BackendServerOptions struct {
	beego.Config
}

func (b *BackendServerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&b.Listen.HTTPSAddr, "https-addr", b.Listen.HTTPSAddr,
		"Https address for backend-service.")
	fs.IntVar(&b.Listen.HTTPSPort, "https-port", b.Listen.HTTPSPort,
		"Https port for backend-service.")
	fs.BoolVar(&b.Listen.EnableAdmin, "enable-admin", b.Listen.EnableAdmin,
		"open pprof or not. deafult is false")
}
