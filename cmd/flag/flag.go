package flag

import (
	"github.com/urfave/cli"
	"github.com/zhiqiangxu/evm-lab/config"
)

// ConfigFlag ...
var ConfigFlag = cli.StringFlag{
	Name:  "cfg",
	Usage: "specify config file",
	Value: config.ConfigJSON,
}

// GasFlag ...
var GasFlag = cli.Uint64Flag{
	Name:  "gas",
	Usage: "gas limit for tx",
	Value: 10000000000,
}

// GasPriceFlag ...
var GasPriceFlag = cli.StringFlag{
	Name:  "gas_price",
	Usage: "gas price for tx",
	Value: "0",
}

// ValueFlag ...
var ValueFlag = cli.StringFlag{
	Name:  "value",
	Usage: "value for tx",
	Value: "0",
}

// ContractPathFlag ...
var ContractPathFlag = cli.StringFlag{
	Name:     "contract_path",
	Usage:    "specify contract path",
	Required: true,
}

// MethodFlag ...
var MethodFlag = cli.StringFlag{
	Name:     "method",
	Usage:    "specify method name",
	Required: true,
}

// SenderFlag ...
var SenderFlag = cli.StringFlag{
	Name:     "sender",
	Usage:    "sender of tx",
	Required: true,
}

// ReceiverFlag ...
var ReceiverFlag = cli.StringFlag{
	Name:  "receiver",
	Usage: "receiver of tx",
}

// DirFlag ...
var DirFlag = cli.StringFlag{
	Name:  "dir",
	Usage: "directory of solidity",
}

// VersionFlag ...
var VersionFlag = cli.StringFlag{
	Name:  "version",
	Usage: "version of solidity",
}
