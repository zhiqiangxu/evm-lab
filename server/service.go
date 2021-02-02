package server

import (
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	goruntime "runtime"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/core/vm/runtime"
	"github.com/ethereum/go-ethereum/params"
)

func (s *Server) handleCreateContract(input CreateContractInput) (output CreateContractOutput) {

	logconfig := &vm.LogConfig{
		DisableMemory:     s.conf.DisableMemory,
		DisableStack:      s.conf.DisableStack,
		DisableStorage:    s.conf.DisableStorage,
		DisableReturnData: s.conf.DisableReturnData,
		Debug:             s.conf.Debug,
	}

	if s.conf.Genesis.GasLimit != 0 {
		input.Gas = s.conf.Genesis.GasLimit
	}

	var (
		tracer      vm.Tracer
		debugLogger *vm.StructLogger
	)

	if s.conf.Machine {
		tracer = vm.NewJSONLogger(logconfig, os.Stdout)
	} else if s.conf.Debug {
		debugLogger = vm.NewStructLogger(logconfig)
		tracer = debugLogger
	} else {
		debugLogger = vm.NewStructLogger(logconfig)
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
			Tracer:         tracer,
			Debug:          s.conf.Debug || s.conf.Machine,
			EVMInterpreter: s.conf.EVMInterpreter,
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
		output.ErrMsg = err.Error()
		return
	}

	s.statedb.Commit(true)
	s.statedb.IntermediateRoot(true)

	if s.conf.Dump {
		fmt.Println(string(s.statedb.Dump(false, false, true)))
	}

	if s.conf.Debug {
		if debugLogger != nil {
			fmt.Fprintln(os.Stderr, "#### TRACE ####")
			vm.WriteTrace(os.Stderr, debugLogger.StructLogs())
		}
		fmt.Fprintln(os.Stderr, "#### LOGS ####")
		vm.WriteLogs(os.Stderr, s.statedb.Logs())
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

func (s *Server) handleCallContract(input CallContractInput) (output CallContractOutput) {
	logconfig := &vm.LogConfig{
		DisableMemory:     s.conf.DisableMemory,
		DisableStack:      s.conf.DisableStack,
		DisableStorage:    s.conf.DisableStorage,
		DisableReturnData: s.conf.DisableReturnData,
		Debug:             s.conf.Debug,
	}

	if s.conf.Genesis.GasLimit != 0 {
		input.Gas = s.conf.Genesis.GasLimit
	}

	var (
		tracer      vm.Tracer
		debugLogger *vm.StructLogger
	)

	if s.conf.Machine {
		tracer = vm.NewJSONLogger(logconfig, os.Stdout)
	} else if s.conf.Debug {
		debugLogger = vm.NewStructLogger(logconfig)
		tracer = debugLogger
	} else {
		debugLogger = vm.NewStructLogger(logconfig)
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
			Tracer:         tracer,
			Debug:          s.conf.Debug || s.conf.Machine,
			EVMInterpreter: s.conf.EVMInterpreter,
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
		output.ErrMsg = err.Error()
		return
	}

	s.statedb.Commit(true)
	s.statedb.IntermediateRoot(true)
	if s.conf.Dump {
		fmt.Println(string(s.statedb.Dump(false, false, true)))
	}

	if s.conf.Debug {
		if debugLogger != nil {
			fmt.Fprintln(os.Stderr, "#### TRACE ####")
			vm.WriteTrace(os.Stderr, debugLogger.StructLogs())
		}
		fmt.Fprintln(os.Stderr, "#### LOGS ####")
		vm.WriteLogs(os.Stderr, s.statedb.Logs())
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
