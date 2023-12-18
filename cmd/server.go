package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/davecgh/go-spew/spew"
	"github.com/kiga-hub/arc/micro"
	basicComponent "github.com/kiga-hub/arc/micro/component"
	arcTracing "github.com/kiga-hub/arc/tracing"
	"github.com/spf13/cobra"

	"github.com/kiga-hub/arc-consumer/pkg/component"
)

func init() {
	spew.Config = *spew.NewDefaultConfig()
	spew.Config.ContinueOnMethod = true
}

// serverCmd .
var serverCmd = &cobra.Command{
	Use:   "run",
	Short: "run data arc-consumer",
	Run:   run,
}

func run(cmd *cobra.Command, args []string) {
	// recover
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered", "recover", r)
			debug.PrintStack()
			os.Exit(1)
		}
	}()

	server, err := micro.NewServer(
		AppName,
		AppVersion,
		[]micro.IComponent{
			&basicComponent.LoggingComponent{},
			&arcTracing.Component{},
			&basicComponent.GossipKVCacheComponent{
				ClusterName:   "platform-global",
				Port:          6666,
				InMachineMode: false,
			},
			&component.ArcConsumerComponent{},
		},
	)
	if err != nil {
		panic(err)
	}
	err = server.Init()
	if err != nil {
		panic(err)
	}
	err = server.Run()
	if err != nil {
		panic(err)
	}
}
