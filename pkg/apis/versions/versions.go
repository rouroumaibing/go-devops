package versions

import (
	"fmt"
	"runtime"
)

// 变量通过-ldflags -X importpath.name=value在编译的时候传入
var (
	version   = ""
	gitTag    = ""
	gitBranch = ""
	gitCommit = "$Format:%H$"
	buildDate = "2006-01-02 15:04:05"
)

type Version struct {
	Version   string `json:"version"`
	GitTag    string `json:"gitTag"`
	GitBranch string `json:"gitBranch"`
	GitCommit string `json:"gitCommit"` // git rev-parse --short HEAD
	BuildDate string `json:"buildDate"`
	GoVersion string `json:"goVersion"`
	Compiler  string `json:"compiler"`
	Platform  string `json:"platform"`
}

func Info() Version {
	return Version{
		Version:   version,
		GitTag:    gitTag,
		GitBranch: gitBranch,
		GitCommit: gitCommit,
		BuildDate: buildDate,
		GoVersion: runtime.Version(),
		Compiler:  runtime.Compiler,
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
