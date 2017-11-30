package chari

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/orderer/multichain"
	cb "github.com/hyperledger/fabric/protos/common"
	"github.com/op/go-logging"
	//	"github.com/spf13/viper"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tendermint/abci/server"
	rpc "github.com/tendermint/tendermint/rpc/client"
	cmn "github.com/tendermint/tmlibs/common"
)

var logger = logging.MustGetLogger("chari")
var (
	once   sync.Once
	client *rpc.HTTP
	db     *leveldb.DB
	_echo  *echo
)

const (
	appHash = "AramakiMisaki" //4172616D616B694D6973616B69
)

type chain struct {
	support  multichain.ConsenterSupport
	sendChan chan *cb.Envelope
	exitChan chan struct{}
}

func _Init() {
	once.Do(func() {
		if os.Getenv("BFT_PROXY_APP") == "" {
			panic("BFT_PROXY_APP not found")
		}
		if os.Getenv("BFT_RPC_ADDR") == "" {
			panic("BFT_RPC_ADDR not found")
		}

		format := logging.MustStringFormatter(`[%{module}] %{time:2006-01-02 15:04:05} [%{level}] [%{longpkg} %{shortfile}] { %{message} }`)

		backendConsole := logging.NewLogBackend(os.Stderr, "", 0)
		backendConsole2Formatter := logging.NewBackendFormatter(backendConsole, format)

		logging.SetBackend(backendConsole2Formatter)
		logging.SetLevel(logging.INFO, "")

		var err error
		db, err = leveldb.OpenFile("/var/hyperledger/production/chari", nil)
		if err != nil {
			panic(err)
		}
		defer db.Close()

		client = rpc.NewHTTP(fmt.Sprintf("tcp://%s", os.Getenv("BFT_RPC_ADDR")), "/websocket")

		_echo = &echo{supports: make(map[string]multichain.ConsenterSupport, 10)}
		srv, err := server.NewServer(os.Getenv("BFT_PROXY_APP"), "socket", _echo)
		if err != nil {
			panic(err)
		}

		logger.Info("trying to listen on", os.Getenv("BFT_PROXY_APP"))
		if _, err = srv.Start(); err != nil {
			panic(err)
		}

		// Wait forever
		cmn.TrapSignal(func() {
			// Cleanup
			srv.Stop()
			logger.Error("ABCI Server shutdown")
		})
	})
}

func New() (consenter multichain.Consenter) {
	go _Init()
	time.Sleep(time.Second * 2)
	return &chain{}
}

func (this *chain) HandleChain(support multichain.ConsenterSupport, metadata *cb.Metadata) (multichain.Chain, error) {
	if len(support.ChainID()) > 255 {
		panic("support.ChainID() length over 255")
	}
	logger.Info("new chain", support.ChainID())
	this.support = support
	this.sendChan = make(chan *cb.Envelope)
	this.exitChan = make(chan struct{})
	return this, nil
}

func (this *chain) Start() {
	_echo.NewSupport(this.support.ChainID(), this.support)
}

func (this *chain) Halt() {
	select {
	case <-this.exitChan:
		logger.Warningf("[channel: %s] Halting of chain requested again", this.support.ChainID())
	default:
		panic(fmt.Sprintf("[channel: %s] Halting of chain requested", this.support.ChainID()))
		close(this.exitChan)
	}
}

func (this *chain) Errored() <-chan struct{} {
	return this.exitChan
}

func (this *chain) Enqueue(env *cb.Envelope) bool {
	logger.Info(this.support.ChainID(), "Enqueue")
	select {
	case <-this.exitChan:
		return false
	default:
		raw, err := proto.Marshal(env)
		if err != nil {
			logger.Error(err)
			return false
		}

		l := len(this.support.ChainID())
		channelId := make([]byte, (1 + l))
		channelId[0] = byte(l)
		copy(channelId[1:], []byte(this.support.ChainID()))

		resp, err := client.BroadcastTxSync(append(channelId, raw...))
		if err != nil {
			logger.Error(err)
			return false
		}

		if !resp.Code.IsOK() {
			logger.Error(resp.Log)
			return false
		}
		return true
	}
}
