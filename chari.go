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
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/tendermint/abci/server"
	rpc "github.com/tendermint/tendermint/rpc/client"
)

var logger = logging.MustGetLogger("chari")
var (
	once   sync.Once
	client *rpc.HTTP
	db     *leveldb.DB
	_echo  *echo
)

type chain struct {
	support  multichain.ConsenterSupport
	sendChan chan *cb.Envelope
	exitChan chan struct{}
}

func _Init() {
	once.Do(func() {
		format := logging.MustStringFormatter(`[%{module}] %{time:2006-01-02 15:04:05} [%{level}] [%{longpkg} %{shortfile}] { %{message} }`)

		backendConsole := logging.NewLogBackend(os.Stderr, "", 0)
		backendConsole2Formatter := logging.NewBackendFormatter(backendConsole, format)

		logging.SetBackend(backendConsole2Formatter)
		logging.SetLevel(logging.INFO, "")

		err := initConfig()
		if err != nil {
			panic(err)
		}

		db, err = leveldb.OpenFile("/var/hyperledger/production/chari", nil)
		if err != nil {
			panic(err)
		}
		defer db.Close()

		client = rpc.NewHTTP(rpc_laddr, "/websocket")

		_echo = &echo{supports: make(map[string]multichain.ConsenterSupport, 10)}
		srv, err := server.NewServer(proxy_app, "socket", _echo)
		if err != nil {
			panic(err)
		}

		go func() {
			time.Sleep(time.Second * 5)
			logger.Info("trying to startup tendermint")
			startTendermint()
		}()

		logger.Info("trying to listen on", proxy_app)
		if _, err = srv.Start(); err != nil {
			panic(err)
		}

		wait := make(chan int)
		<-wait
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
