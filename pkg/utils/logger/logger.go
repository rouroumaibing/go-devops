package logger

import (
    "os"
    "gopkg.in/natefinch/lumberjack.v2"
    "k8s.io/klog/v2"
)


// 配合pflag.CommandLine.AddGoFlagSet(flag.CommandLine)追加参数到pflag使用

// klog代码中初始化日志级别为0-3（klog.Info/klog.Warning/klog.Error/klog.Fatal）

// flag.Set("logtostderr", "false") 这条语句会创建日志文件，并设置标准错误不打印到标准错误输出.在pflag.Parse()前设置,紧跟klog.InitFlags(nil)之后

func SetLogratate(){
    // flag.Set("log_file", logPath)
    logPath := os.Getenv("LOG_FILE")
    lumberJackLogger := &lumberjack.Logger{
        Filename:   logPath,
        MaxSize:    50,
        MaxBackups: 20,
        MaxAge:     28,  // days
        Compress:   true, //压缩
    } 
    klog.SetOutput(lumberJackLogger)
}
