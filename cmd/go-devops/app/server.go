package app

import (
	goflag "flag"
	"os"
	"time"

	beego "github.com/astaxie/beego"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	cmnUtil "github.com/rouroumaibing/go-devops/cmd/common/utils"
	"github.com/rouroumaibing/go-devops/cmd/go-devops/app/options"
	_ "github.com/rouroumaibing/go-devops/pkg/models"
	jwtauth "github.com/rouroumaibing/go-devops/pkg/utils/auth/jwt"
	"github.com/rouroumaibing/go-devops/pkg/utils/customstring"
	"github.com/rouroumaibing/go-devops/pkg/utils/db"
	"github.com/rouroumaibing/go-devops/pkg/utils/logger"
)

const (
	ComponentName = "go-devops"
)

func init() {
	jwtauth.Init()
	redisInit()
	mysqlInit()
	RunInit()
}

func mysqlInit() {
	driverName := os.Getenv("MYSQL_DRIVERNAME")
	user := os.Getenv("MYSQL_USER")
	pwd := os.Getenv("MYSQL_PWD")
	host := os.Getenv("MYSQL_HOST")
	portStr := os.Getenv("MYSQL_PORT")
	dbname := os.Getenv("MYSQL_DBNAME")
	maxidleStr := os.Getenv("MYSQL_MAXIDLE")
	maxconnStr := os.Getenv("MYSQL_MAXCONN")

	port, err := customstring.EnvToInt(portStr)
	if err != nil {
		port = 3306
	}
	maxidle, err := customstring.EnvToInt(maxidleStr)
	if err != nil {
		maxidle = 1000
	}
	maxconn, err := customstring.EnvToInt(maxconnStr)
	if err != nil {
		maxconn = 2000
	}

	mysqlcnn := &db.MysqlConfig{
		DriverName: driverName,
		User:       user,
		Pwd:        pwd,
		Host:       host,
		Port:       port,
		DBName:     dbname,
		MaxIdle:    maxidle,
		MaxConn:    maxconn,
	}
	mysqlcnn.Init()
	os.Unsetenv("MYSQL_PWD")
}

func redisInit() {
	address := os.Getenv("REDIS_ADDRESS")
	pwd := os.Getenv("REDIS_PWD")
	poolsizeStr := os.Getenv("REDIS_POOLSIZE")
	maxidleStr := os.Getenv("REDIS_MAXIDLE")
	idletimeoutStr := os.Getenv("REDIS_IDLETIMEOUT")

	poolsize, err := customstring.EnvToInt(poolsizeStr)
	if err != nil {
		poolsize = 1000
	}
	maxidle, err := customstring.EnvToInt(maxidleStr)
	if err != nil {
		maxidle = 64
	}
	idletimeout, err := customstring.EnvToInt64(idletimeoutStr)
	if err != nil {
		idletimeout = 100
	}

	rediscnn := &db.RedisConfig{
		Address:     address,
		Password:    pwd,
		PoolSize:    poolsize,
		MaxIdle:     maxidle,
		IdleTimeOut: time.Duration(idletimeout) * time.Minute,
	}
	rediscnn.Init()
	os.Unsetenv("REDIS_PWD")
}

func RunInit() {
	klog.InitFlags(nil)
	if err := goflag.Set("logtostderr", "false"); err != nil {
		klog.Error(err)
	}

	s := options.NewServerRunOptions()
	s.AddFlags(pflag.CommandLine)

	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	pflag.Parse()

	logger.SetLogratate()

	err := cmnUtil.InitServerSecureConfig(s.BackendServerOptions.Config, s.CommonOptions)
	if err != nil {
		klog.Error(err)
	}

	beego.BConfig.MaxMemory = int64(s.MaxRequestBodyBytes + 1)
	// 是否允许在 HTTP 请求时，返回原始请求体数据字节，默认为 false （GET or HEAD or 上传文件请求除外）。
	beego.BConfig.CopyRequestBody = true

}
