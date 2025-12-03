package app

import (
	goflag "flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
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
	"github.com/rouroumaibing/go-devops/pkg/utils/system"
	"github.com/rouroumaibing/go-devops/pkg/workers"
)

const (
	ComponentName = "go-devops"
)

func init() {
	jwtauth.Init()
	redisInit()
	mysqlInit()
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

	node, producer, consumer := WokersInit()

	// 将defer语句放在前面，确保资源一定会被清理
	defer klog.Flush()
	defer db.RedisClient.Close()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 在goroutine中运行beego服务器，避免阻塞主协程
	go beego.Run()

	// 等待中断信号
	<-sigChan
	klog.Info("Received shutdown signal, starting graceful shutdown...")

	if err := node.Shutdown(); err != nil {
		klog.Errorf("node.Shutdown failed: %v", err)
	}
	consumer.Stop()
	producer.Stop()

	klog.Info("Graceful shutdown completed")
	os.Exit(0)
}

func WokersInit() (node *workers.DistributedNode, producer *workers.TaskProducer, consumer *workers.TaskConsumer) {
	raftClusterID := os.Getenv("RAFT_CLUSTER_ID")
	clusterID := os.Getenv("CLUSTER_ID")
	raftPort := os.Getenv("RAFT_PORT")
	if raftPort == "" {
		raftPort = workers.DefaultRaftPort
	}

	dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		dataDir = workers.DefaultDataDir
	}

	localIP, err := system.GetLocalIP()
	if err != nil {
		log.Fatalf("Failed to get local IP address: %v", err)
	}
	raftAddr := fmt.Sprintf("%s:%s", localIP, raftPort)

	cfg := workers.NodeConfig{
		RaftClusterID: raftClusterID,
		NodeIP:        localIP,
		ClusterID:     clusterID,
		RaftAddr:      raftAddr,
		DataDir:       dataDir,
	}

	node, err = workers.NewDistributedNode(cfg)
	if err != nil {
		log.Fatalf("Failed to create node: %v", err)
	}

	producer = workers.NewTaskProducer(node)
	producer.Start()

	consumer = workers.NewTaskConsumer(node)
	consumer.Start()

	return node, producer, consumer
}
