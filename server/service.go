package server

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	goruntime "runtime"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/eth/tracers/logger"
	"github.com/ethereum/go-ethereum/params"
)

func (s *Server) handleDeploy(input DeployInput) (output DeployOutput) {

	logconfig := &logger.Config{
		EnableMemory:     !s.conf.DisableMemory,
		DisableStack:     s.conf.DisableStack,
		DisableStorage:   s.conf.DisableStorage,
		EnableReturnData: !s.conf.DisableReturnData,
		Debug:            s.conf.Debug,
	}

	if s.conf.Genesis.GasLimit != 0 {
		input.Gas = s.conf.Genesis.GasLimit
	}

	var (
		tracer      vm.EVMLogger
		debugLogger *logger.StructLogger
	)

	if s.conf.Machine {
		tracer = logger.NewJSONLogger(logconfig, os.Stdout)
	} else if s.conf.Debug {
		debugLogger = logger.NewStructLogger(logconfig)
		tracer = debugLogger
	} else {
		debugLogger = logger.NewStructLogger(logconfig)
	}

	fmt.Println("sender", input.Sender.Hex(), "balance", s.statedb.GetBalance(input.Sender), "nonce", s.statedb.GetNonce(input.Sender))

	runtimeConfig := runtime.Config{
		Origin:      input.Sender,
		State:       s.statedb,
		GasLimit:    input.Gas,
		GasPrice:    input.GasPrice,
		Value:       input.Value,
		Difficulty:  s.conf.Genesis.Difficulty,
		Time:        new(big.Int).SetUint64(s.conf.Genesis.Timestamp),
		Coinbase:    s.conf.Genesis.Coinbase,
		BlockNumber: new(big.Int).SetUint64(s.conf.Genesis.Number),
		EVMConfig: vm.Config{
			Tracer: tracer,
			Debug:  s.conf.Debug || s.conf.Machine,
		},
	}

	if s.conf.Genesis.Config != nil {
		runtimeConfig.ChainConfig = s.conf.Genesis.Config
	} else {
		runtimeConfig.ChainConfig = params.AllEthashProtocolChanges
	}

	execFunc := func() ([]byte, uint64, error) {
		outputBytes, addr, gasLeft, err := runtime.Create(input.CodeAndInput, &runtimeConfig)
		output.Addr = addr
		return outputBytes, gasLeft, err
	}

	outputBytes, leftOverGas, stats, err := timedExec(s.conf.Bench, execFunc)
	if err != nil {
		output.ErrMsg = parseRevertReason(err, outputBytes)
		return
	}

	s.statedb.Commit(true)
	s.statedb.IntermediateRoot(true)

	if s.conf.Dump {
		fmt.Println(string(s.statedb.Dump(nil)))
	}

	if s.conf.Debug {
		if debugLogger != nil {
			fmt.Fprintln(os.Stderr, "#### TRACE ####")
			logger.WriteTrace(os.Stderr, debugLogger.StructLogs())
		}
		fmt.Fprintln(os.Stderr, "#### LOGS ####")
		logger.WriteLogs(os.Stderr, s.statedb.Logs())
	}

	if s.conf.Bench || s.conf.StatDump {
		fmt.Fprintf(os.Stderr, `EVM gas used:    %d
execution time:  %v
allocations:     %d
allocated bytes: %d
`, input.Gas-leftOverGas, stats.time, stats.allocs, stats.bytesAllocated)
	}
	if tracer == nil {
		fmt.Printf("0x%x\n", outputBytes)
		if err != nil {
			fmt.Printf(" error: %v\n", err)
		}
	}
	return
}

// FYI: https://coder-question.com/cq-blog/194033
func parseRevertReason(revertErr error, returnData []byte) string {
	StringTy, _ := abi.NewType("string", "", nil)

	param := abi.Arguments{
		{Type: StringTy},
	}
	args, err := param.Unpack(returnData[4:])
	var msg string
	if err != nil {
		msg = err.Error()
	} else {
		msg = args[0].(string)
	}
	return fmt.Sprintf("Error:%v Method:0x%s Msg:'%s' Raw:%v", revertErr, hex.EncodeToString(returnData[0:4]), msg, returnData)
}

type execStats struct {
	time           time.Duration // The execution time.
	allocs         int64         // The number of heap allocations during execution.
	bytesAllocated int64         // The cumulative number of bytes allocated during execution.
}

func timedExec(bench bool, execFunc func() ([]byte, uint64, error)) (output []byte, gasLeft uint64, stats execStats, err error) {
	if bench {
		result := testing.Benchmark(func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				output, gasLeft, err = execFunc()
			}
		})

		// Get the average execution time from the benchmarking result.
		// There are other useful stats here that could be reported.
		stats.time = time.Duration(result.NsPerOp())
		stats.allocs = result.AllocsPerOp()
		stats.bytesAllocated = result.AllocedBytesPerOp()
	} else {
		var memStatsBefore, memStatsAfter goruntime.MemStats
		goruntime.ReadMemStats(&memStatsBefore)
		startTime := time.Now()
		output, gasLeft, err = execFunc()
		stats.time = time.Since(startTime)
		goruntime.ReadMemStats(&memStatsAfter)
		stats.allocs = int64(memStatsAfter.Mallocs - memStatsBefore.Mallocs)
		stats.bytesAllocated = int64(memStatsAfter.TotalAlloc - memStatsBefore.TotalAlloc)
	}

	return output, gasLeft, stats, err
}

func (s *Server) handleCall(input CallInput) (output CallOutput) {
	logconfig := &logger.Config{
		EnableMemory:     !s.conf.DisableMemory,
		DisableStack:     s.conf.DisableStack,
		DisableStorage:   s.conf.DisableStorage,
		EnableReturnData: !s.conf.DisableReturnData,
		Debug:            s.conf.Debug,
	}

	if s.conf.Genesis.GasLimit != 0 {
		input.Gas = s.conf.Genesis.GasLimit
	}

	var (
		tracer      vm.EVMLogger
		debugLogger *logger.StructLogger
	)

	if s.conf.Machine {
		tracer = logger.NewJSONLogger(logconfig, os.Stdout)
	} else if s.conf.Debug {
		debugLogger = logger.NewStructLogger(logconfig)
		tracer = debugLogger
	} else {
		debugLogger = logger.NewStructLogger(logconfig)
	}

	// s.statedb.CreateAccount(input.Sender)

	runtimeConfig := runtime.Config{
		Origin:      input.Sender,
		State:       s.statedb,
		GasLimit:    input.Gas,
		GasPrice:    input.GasPrice,
		Value:       input.Value,
		Difficulty:  s.conf.Genesis.Difficulty,
		Time:        new(big.Int).SetUint64(s.conf.Genesis.Timestamp),
		Coinbase:    s.conf.Genesis.Coinbase,
		BlockNumber: new(big.Int).SetUint64(s.conf.Genesis.Number),
		EVMConfig: vm.Config{
			Tracer: tracer,
			Debug:  s.conf.Debug || s.conf.Machine,
		},
	}

	if s.conf.Genesis.Config != nil {
		runtimeConfig.ChainConfig = s.conf.Genesis.Config
	} else {
		runtimeConfig.ChainConfig = params.AllEthashProtocolChanges
	}

	execFunc := func() ([]byte, uint64, error) {
		return runtime.Call(input.Receiver, input.Input, &runtimeConfig)
	}

	outputBytes, leftOverGas, stats, err := timedExec(s.conf.Bench, execFunc)
	output.Result = outputBytes
	if err != nil {
		output.ErrMsg = parseRevertReason(err, outputBytes)
		return
	}

	s.statedb.Commit(true)
	s.statedb.IntermediateRoot(true)
	if s.conf.Dump {
		fmt.Println(string(s.statedb.Dump(nil)))
	}

	if s.conf.Debug {
		if debugLogger != nil {
			fmt.Fprintln(os.Stderr, "#### TRACE ####")
			logger.WriteTrace(os.Stderr, debugLogger.StructLogs())
		}
		fmt.Fprintln(os.Stderr, "#### LOGS ####")
		logger.WriteLogs(os.Stderr, s.statedb.Logs())
	}

	if s.conf.Bench || s.conf.StatDump {
		fmt.Fprintf(os.Stderr, `EVM gas used:    %d
execution time:  %v
allocations:     %d
allocated bytes: %d
`, input.Gas-leftOverGas, stats.time, stats.allocs, stats.bytesAllocated)
	}
	if tracer == nil {
		fmt.Printf("0x%x\n", outputBytes)
		if err != nil {
			fmt.Printf(" error: %v\n", err)
		}
	}
	return
}
