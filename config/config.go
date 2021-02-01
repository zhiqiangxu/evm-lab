package config

import "github.com/ethereum/go-ethereum/core"

// Config ...
type Config struct {
	Port              int
	Genesis           *core.Genesis
	Verbosity         int
	DisableMemory     bool
	DisableStack      bool
	DisableStorage    bool
	DisableReturnData bool
	Bench             bool
	EVMInterpreter    string
	Machine           bool
	Debug             bool
	Dump              bool
	StatDump          bool
}
