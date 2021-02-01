package server

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

// CreateContractInput ...
type CreateContractInput struct {
	Sender       common.Address
	CodeAndInput []byte
	Gas          uint64
	GasPrice     *big.Int
	Value        *big.Int
}

// CreateContractOutput ...
type CreateContractOutput struct {
	Addr   common.Address
	Result []byte
	ErrMsg string
}

// CallContractInput ...
type CallContractInput struct {
	Input    []byte
	Gas      uint64
	Sender   common.Address
	Receiver common.Address
	GasPrice *big.Int
	Value    *big.Int
}

// CallContractOutput ...
type CallContractOutput struct {
	Result []byte
	ErrMsg string
}
