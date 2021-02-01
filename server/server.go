package server

import (
	"fmt"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/log"
	"github.com/gin-gonic/gin"
	"github.com/zhiqiangxu/evm-lab/config"
	"github.com/zhiqiangxu/util/mutex"
)

const (
	// CreateContractEndpoint ...
	CreateContractEndpoint = "/create_contract"
	// CallContractEndpoint ...
	CallContractEndpoint = "/call_contract"
)

// Server ...
type Server struct {
	conf    config.Config
	tmutex  *mutex.TMutex
	statedb *state.StateDB
}

// New ...
func New(conf config.Config) *Server {
	return &Server{tmutex: mutex.New(), conf: conf}
}

// Start ...
func (s *Server) Start() (err error) {
	err = s.initState()
	if err != nil {
		return
	}
	err = s.startHTTP()
	return
}

func (s *Server) initState() (err error) {

	db := rawdb.NewMemoryDatabase()
	if s.conf.Genesis != nil {
		genesis := s.conf.Genesis.ToBlock(db)
		s.statedb, _ = state.New(genesis.Root(), state.NewDatabase(db), nil)
	} else {
		s.statedb, _ = state.New(common.Hash{}, state.NewDatabase(db), nil)
		s.conf.Genesis = &core.Genesis{}
	}

	return
}

func (s *Server) initLogger() {
	glogger := log.NewGlogHandler(log.StreamHandler(os.Stderr, log.TerminalFormat(false)))
	glogger.Verbosity(log.Lvl(s.conf.Verbosity))
	log.Root().SetHandler(glogger)
}

func (s *Server) startHTTP() error {
	r := gin.Default()

	r.POST(CreateContractEndpoint, s.createContract)
	r.POST(CallContractEndpoint, s.callContract)

	return r.Run(fmt.Sprintf(":%d", s.conf.Port))

}
