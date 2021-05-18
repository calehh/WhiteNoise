package main

import (
	"fmt"
	"github.com/urfave/cli"
	"os"
	"os/signal"
	"syscall"
	"whitenoise/network"
	"whitenoise/sdk"
)

package main

import (
"bufio"
"context"
"fmt"
"github.com/golang/protobuf/proto"
"github.com/urfave/cli"
"io"
"math/rand"
"os"
"os/signal"
"syscall"
"time"
"whitenoise/cmd/chat"
"whitenoise/common/account"
"whitenoise/common/config"
"whitenoise/common/log"
"whitenoise/internal/pb"
"whitenoise/network"
"whitenoise/sdk"
"whitenoise/secure"
)

var node *network.Node
var wnSDK *sdk.WhiteNoiseClient
var (
	WSPortFlag = cli.IntFlag{
		Name:  "port, p",
		Value: 8001,
	}

	BootStrapFlag = cli.StringFlag{
		Name:  "bootstrap, b",
		Usage: "PeerId of the node to bootstrap from.",
		Value: "",
	}
	NodeFlag = cli.StringFlag{
		Name:  "node, n",
		Usage: "PeerId of the node to connect to.",
		Value: "",
	}
	ModeFlag = cli.BoolFlag{
		Name:     "client, c",
		Usage:    "Build a client node if flag is on and MainNet Node by default.",
		Required: false,
		Hidden:   false,
	}
	BootFlag = cli.BoolFlag{
		Name:     "boot",
		Usage:    "Build a boot node if flag is on and MainNet Node by default.",
		Required: false,
		Hidden:   false,
	}
	LogLevelFlag = cli.IntFlag{
		Name:  "log, l",
		Value: 2,
	}
	NickFlag = cli.StringFlag{
		Name:  "nick",
		Usage: "Set nick name for chat example client",
		Value: "",
	}
)

func main() {
	if err := initApp().Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func initApp() *cli.App {
	app := cli.NewApp()
	app.Usage = " whitenoise protocol implement"
	app.Action = func() {}
	app.Flags = []cli.Flag{
		WSPortFlag,
	}
	return app

}

func Start(ctx *cli.Context) {
	port := ctx.String("port")
	ws := NewWsService(port)

}

func waitToExit() {
	exit := make(chan bool, 0)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sc {
			fmt.Printf("received exit signal:%v", sig.String())
			close(exit)
			break
		}
	}()
	<-exit
}
