package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"io/ioutil"
	"math/big"
	"net/http"
	"path/filepath"
	"regexp"
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
		clientDeployCmd,
		clientCallCmd,
		clientModSolcVersionCmd,
	},
}

var clientDeployCmd = cli.Command{
	Name:   "deploy",
	Usage:  "trigger deploy action",
	Action: clientDeploy,
	Flags: []cli.Flag{
		flag.SenderFlag,
		flag.ContractPathFlag,
		flag.GasFlag,
		flag.GasPriceFlag,
		flag.ValueFlag,
		flag.ConfigFlag,
	},
}

var clientCallCmd = cli.Command{
	Name:   "call",
	Usage:  "trigger call action",
	Action: clientCall,
	Flags: []cli.Flag{
		flag.SenderFlag,
		flag.ReceiverFlag,
		flag.ContractPathFlag,
		flag.MethodFlag,
		flag.GasFlag,
		flag.GasPriceFlag,
		flag.ValueFlag,
		flag.ConfigFlag,
	},
}

var clientModSolcVersionCmd = cli.Command{
	Name:   "msv",
	Usage:  "modify solc version",
	Action: clientModSolcVersion,
	Flags: []cli.Flag{
		flag.DirFlag,
		flag.VersionFlag,
	},
}

func clientDeploy(ctx *cli.Context) (err error) {

	// type DeployInput struct {
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
		if strings.ToLower(nameParts[len(nameParts)-1])+".sol" != strings.ToLower(contractFileName) {
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

	if contract == nil {
		err = fmt.Errorf("contract not found")
		return
	}
	codeAndInput := append(common.FromHex(contract.Code), inputBin...)

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

	input := server.DeployInput{
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
	resp, err := http.Post(fmt.Sprintf("http://localhost:%d%s", conf.Port, server.DeployEndpoint), "application/json", bytes.NewBuffer(inputBytes))
	if err != nil {
		err = fmt.Errorf("API err:%v", err)
		return
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var output server.DeployOutput
	err = json.Unmarshal(respBytes, &output)
	if err != nil {
		return
	}

	outputBytes, _ := json.Marshal(output)
	fmt.Println("output", string(outputBytes))
	return
}

func clientCall(ctx *cli.Context) (err error) {

	// type CallInput struct {
	// 	Input    []byte
	// 	Gas      uint64
	// 	Sender   common.Address
	// 	Receiver common.Address
	// 	GasPrice *big.Int
	// 	Value    *big.Int
	// }

	sender := common.HexToAddress(ctx.String(flag.SenderFlag.Name))
	receiver := common.HexToAddress(ctx.String(flag.ReceiverFlag.Name))
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
		if strings.ToLower(nameParts[len(nameParts)-1])+".sol" != strings.ToLower(contractFileName) {
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
		if strings.HasPrefix(arg, "0x") {
			args = append(args, common.HexToAddress(arg))
		} else {
			args = append(args, arg)
		}

	}

	inputBin, err := contractABI.Pack(ctx.String(flag.MethodFlag.Name), args...)
	if err != nil {
		err = fmt.Errorf("abi.Pack err:%v", err)
		return
	}

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

	input := server.CallInput{
		Sender:   sender,
		Receiver: receiver,
		Input:    inputBin,
		Gas:      gas,
		GasPrice: gasPrice,
		Value:    value,
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
	resp, err := http.Post(fmt.Sprintf("http://localhost:%d%s", conf.Port, server.CallEndpoint), "application/json", bytes.NewBuffer(inputBytes))
	if err != nil {
		err = fmt.Errorf("API err:%v", err)
		return
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	var output server.CallOutput
	err = json.Unmarshal(respBytes, &output)
	if err != nil {
		return
	}

	outputBytes, _ := json.Marshal(output)
	fmt.Println("output", string(outputBytes))
	return
}

func clientModSolcVersion(ctx *cli.Context) (err error) {
	re := regexp.MustCompile(`pragma\s+solidity\s+([^;])*`)

	fmt.Println("root dir", ctx.String(flag.DirFlag.Name))
	err = filepath.WalkDir(ctx.String(flag.DirFlag.Name), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if filepath.Ext(d.Name()) == ".sol" {
			fmt.Println("handling", path)
			// modify solc version
			code, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}

			replacedCode := re.ReplaceAllString(string(code), fmt.Sprintf("pragma solidity %s", ctx.String(flag.VersionFlag.Name)))
			return ioutil.WriteFile(path, []byte(replacedCode), 0777)
		}
		return nil
	})
	return
}
