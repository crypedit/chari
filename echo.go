package chari

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric/orderer/multichain"
	cb "github.com/hyperledger/fabric/protos/common"
	"github.com/tendermint/abci/types"
	"github.com/tendermint/tendermint/version"
)

type echo struct {
	lock            sync.Mutex
	lastBlockHeight uint64
	supports        map[string]multichain.ConsenterSupport
}

func (this *echo) NewSupport(channelId string, support multichain.ConsenterSupport) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.supports[channelId] = support
}

func (this *echo) getSupport(channelId string) multichain.ConsenterSupport {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.supports[channelId]
}

func (this *echo) Info(req types.RequestInfo) types.ResponseInfo {
	logger.Info("Info", req)

	if val, err := db.Get([]byte("lastBlockHeight"), nil); err == nil {
		this.lastBlockHeight = binary.BigEndian.Uint64(val)
	}
	logger.Info("Info LastBlockHeight", this.lastBlockHeight)

	return types.ResponseInfo{
		Version:          version.Version,
		LastBlockHeight:  this.lastBlockHeight,
		LastBlockAppHash: []byte(appHash),
	}
}

func (this *echo) SetOption(key string, value string) (log string) {
	logger.Info("SetOption", key, value)
	return ""
}

func (this *echo) InitChain(req types.RequestInitChain) {
	logger.Info("InitChain", req)
}

func (this *echo) CheckTx(tx []byte) types.Result {
	logger.Info("CheckTx" /*, string(tx)*/)
	return types.OK
}

func (this *echo) BeginBlock(req types.RequestBeginBlock) {
	logger.Info("BeginBlock", req)
}

func (this *echo) DeliverTx(tx []byte) types.Result {
	channelId := string(tx[1 : tx[0]+1])
	logger.Info("DeliverTx", channelId /*, string(tx[tx[0]+1:])*/)

	env := new(cb.Envelope)
	if err := proto.Unmarshal(tx[tx[0]+1:], env); err != nil {
		logger.Error(err)
		return types.NewResult(types.CodeType_InternalError, nil, err.Error())
	}

	support := this.getSupport(channelId)
	if support == nil {
		return types.NewResult(types.CodeType_InternalError, nil, fmt.Sprint("not found channelID ", channelId, " 's support"))
	}

	batches, committers, _, _ := support.BlockCutter().Ordered(env)
	if len(batches) > 0 {
		for i, batch := range batches {
			block := support.CreateNextBlock(batch)
			support.WriteBlock(block, committers[i], nil)
		}
	}
	return types.OK
}

func (this *echo) EndBlock(height uint64) (resp types.ResponseEndBlock) {
	logger.Info("EndBlock", height)
	this.lastBlockHeight = height
	return
}

func (this *echo) Commit() types.Result {
	logger.Info("Commit")

	this.lock.Lock()
	defer this.lock.Unlock()

	for _, support := range this.supports {
		batch, committers := support.BlockCutter().Cut()
		if len(batch) != 0 {
			block := support.CreateNextBlock(batch)
			support.WriteBlock(block, committers, nil)
		}
	}

	height := make([]byte, 8)
	binary.BigEndian.PutUint64(height, this.lastBlockHeight)
	if err := db.Put([]byte("lastBlockHeight"), height, nil); err != nil {
		logger.Error(err)
		return types.NewResult(types.CodeType_InternalError, nil, err.Error())
	}
	return types.NewResultOK([]byte(appHash), "")
}

func (this *echo) Query(req types.RequestQuery) (resp types.ResponseQuery) {
	logger.Info("Query", req)
	return
}
