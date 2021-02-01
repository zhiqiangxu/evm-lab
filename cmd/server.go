package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/urfave/cli"
	"github.com/zhiqiangxu/evm-lab/cmd/flag"
	"github.com/zhiqiangxu/evm-lab/config"
	"github.com/zhiqiangxu/evm-lab/server"
)

// ServerCmd ...
var ServerCmd = cli.Command{
	Name:  "server",
	Usage: "server actions",
	Subcommands: []cli.Command{
		serverStartCmd,
	},
}

var serverStartCmd = cli.Command{
	Name:   "start",
	Usage:  "trigger start action",
	Action: serverStart,
	Flags:  []cli.Flag{flag.ConfigFlag},
}

func serverStart(ctx *cli.Context) (err error) {

	file := ctx.String(flag.ConfigFlag.Name)
	confBytes, err := ioutil.ReadFile(file)
	if err != nil {
		err = fmt.Errorf("config file not found:%s", file)
		return
	}

	var conf config.Config
	err = json.Unmarshal(confBytes, &conf)
	if err != nil {
		return
	}

	confBytes, _ = json.Marshal(conf)
	fmt.Println("conf", string(confBytes))

	svr := server.New(conf)

	return svr.Start()
}
