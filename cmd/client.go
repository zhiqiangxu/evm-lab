package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/cmd/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/compiler"
	"github.com/urfave/cli"
	"github.com/zhiqiangxu/evm-lab/cmd/flag"
	"github.com/zhiqiangxu/evm-lab/config"
	"github.com/zhiqiangxu/evm-lab/server"
)

// ClientCmd ...
var ClientCmd = cli.Command{
	Name:  "client",
	Usage: "client actions",
	Subcommands: []cli.Command{
		clientCreateContractCmd,
	},
}

var clientCreateContractCmd = cli.Command{
	Name:   "create_contract",
	Usage:  "trigger create_contract action",
	Action: clientCreateContract,
	Flags: []cli.Flag{
		flag.SenderFlag,
		flag.ContractPathFlag,
		flag.GasFlag,
		flag.GasPriceFlag,
		flag.ValueFlag,
		flag.ConfigFlag,
	},
}

func clientCreateContract(ctx *cli.Context) (err error) {

	// type CreateContractInput struct {
	// 	Sender       common.Address
	// 	CodeAndInput []byte
	// 	Gas          uint64
	// 	GasPrice     *big.Int
	// 	Value        *big.Int
	// }

	sender := common.HexToAddress(ctx.String(flag.SenderFlag.Name))
	contracts, err := compiler.CompileSolidity("", ctx.String(flag.ContractPathFlag.Name))
	if err != nil {
		utils.Fatalf("CompileSolidity err: %v", err)
	}
	contractFileName := filepath.Base(ctx.String(flag.ContractPathFlag.Name))

	var (
		contract    *compiler.Contract
		contractABI abi.ABI
	)
	for name, c := range contracts {
		nameParts := strings.Split(name, ":")
		if nameParts[len(nameParts)-1]+".sol" != contractFileName {
			continue
		}

		if contract != nil {
			utils.Fatalf("Multiple contracts filtered.")
		}
		contract = c
		abiBytes, _ := json.Marshal(contract.Info.AbiDefinition)
		contractABI, err = abi.JSON(strings.NewReader(string(abiBytes)))
		if err != nil {
			err = fmt.Errorf("abi.JSON err:%v", err)
			return
		}
	}

	var args []interface{}
	for _, arg := range ctx.Args() {
		args = append(args, arg)
	}
	var inputBin []byte
	if len(args) > 0 {
		inputBin, err = contractABI.Pack("", args...)
		if err != nil {
			err = fmt.Errorf("abi.Pack err:%v", err)
			return
		}
	}

	codeAndInput := append([]byte(contract.Code), inputBin...)
	gas := ctx.Uint64(flag.GasFlag.Name)
	gasPrice, ok := big.NewInt(0).SetString(ctx.String(flag.GasPriceFlag.Name), 10)
	if !ok {
		err = fmt.Errorf("invalid gas price:%s", ctx.String(flag.GasPriceFlag.Name))
		return
	}
	value, ok := big.NewInt(0).SetString(ctx.String(flag.ValueFlag.Name), 10)
	if !ok {
		err = fmt.Errorf("invalid value:%s", ctx.String(flag.ValueFlag.Name))
		return
	}

	input := server.CreateContractInput{
		Sender:       sender,
		CodeAndInput: codeAndInput,
		Gas:          gas,
		GasPrice:     gasPrice,
		Value:        value,
	}

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

	inputBytes, _ := json.Marshal(input)
	resp, err := http.Post(fmt.Sprintf("http://localhost:%d%s", conf.Port, server.CreateContractEndpoint), "application/json", bytes.NewBuffer(inputBytes))
	if err != nil {
		err = fmt.Errorf("API err:%v", err)
		return
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var output server.CallContractOutput
	err = json.Unmarshal(respBytes, &output)
	if err != nil {
		return
	}

	outputBytes, _ := json.Marshal(output)
	fmt.Println("output", string(outputBytes))
	return
}
