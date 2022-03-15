package server

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// DeployInput ...
type DeployInput struct {
	Sender       common.Address
	CodeAndInput []byte
	Gas          uint64
	GasPrice     *big.Int
	Value        *big.Int
}

// DeployOutput ...
type DeployOutput struct {
	Addr   common.Address
	ErrMsg string
}

// CallInput ...
type CallInput struct {
	Input    []byte
	Gas      uint64
	Sender   common.Address
	Receiver common.Address
	GasPrice *big.Int
	Value    *big.Int
}

// CallOutput ...
type CallOutput struct {
	Result []byte
	ErrMsg string
}
