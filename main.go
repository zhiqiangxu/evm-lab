package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli"
	"github.com/zhiqiangxu/evm-lab/cmd"
)

func setupAPP() *cli.App {
	app := cli.NewApp()
	app.Usage = "EvmLab Cli"
	app.Copyright = "Copyright in 2021"
	app.Commands = []cli.Command{
		cmd.ServerCmd,
		cmd.ClientCmd,
	}
	return app
}

func main() {
	if err := setupAPP().Run(os.Args); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
