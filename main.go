package main

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/kiga-hub/arc-consumer/cmd"
	_ "go.uber.org/automaxprocs"
)

var (
	// AppName - 应用名称
	AppName string
	// AppVersion - 应用版本
	AppVersion string
	// BuildVersion - 编译版本
	BuildVersion string
	// BuildTime - 编译时间
	BuildTime string
	// GitRevision - Git版本
	GitRevision string
	// GitBranch - Git分支
	GitBranch string
	// GoVersion - Golang信息
	GoVersion string
)

func main() {
	spew.Config = *spew.NewDefaultConfig()
	spew.Config.ContinueOnMethod = true
	cmd.AppName = AppName
	cmd.AppVersion = AppVersion
	Version()

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}

// Version prints version info of the program
func Version() {
	fmt.Printf(
		"App Name:\t%s\nApp Version:\t%s\nBuild version:\t%s\nBuild time:\t%s\nGit revision:\t%s\nGit branch:\t%s\nGolang Version: %s\n",
		AppName,
		AppVersion,
		BuildVersion,
		BuildTime,
		GitRevision,
		GitBranch,
		GoVersion,
	)
}
