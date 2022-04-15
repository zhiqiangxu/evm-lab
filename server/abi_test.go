package server

import (
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"gotest.tools/assert"
)

func TestABI(t *testing.T) {
	StringTy, _ := abi.NewType("string", "", nil)

	param := abi.Arguments{
		{Type: StringTy},
	}

	value := "test"
	packedValue, err := param.Pack(value)
	assert.NilError(t, err, "Pack failed")

	args, err := param.Unpack(packedValue)
	assert.NilError(t, err, "Unpack failed")

	assert.Equal(t, value, args[0].(string))
}
