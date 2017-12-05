package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/amamina/chari"
	"github.com/op/go-logging"
	"github.com/tendermint/tendermint/types"
)

var logger = logging.MustGetLogger("aramaki")

const (
	chainID string = "chain-Chari"
)

func main() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
			logger.Error(string(debug.Stack()))
		}
	}()

	format := logging.MustStringFormatter(`[%{module}] %{time:2006-01-02 15:04:05} [%{level}] [%{longpkg} %{shortfile}] { %{message} }`)

	backendConsole := logging.NewLogBackend(os.Stderr, "", 0)
	backendConsole2Formatter := logging.NewBackendFormatter(backendConsole, format)

	logging.SetBackend(backendConsole2Formatter)
	logging.SetLevel(logging.INFO, "")

	var size int
	flag.IntVar(&size, "s", 4, "default orderers size")
	flag.Parse()

	if size < 4 {
		logger.Error("orderers size must be bigger than 4 to meet PBFT")
		return
	}

	orderers := make([]*types.PrivValidatorFS, size)
	validators := make([]types.GenesisValidator, size)

	for k := 0; k < size; k++ {
		orderers[k] = types.GenPrivValidatorFS("")
		validators[k] = types.GenesisValidator{
			PubKey: orderers[k].PubKey,
			Power:  1,
			Name:   fmt.Sprintf("orderer%d", k),
		}
	}

	genesisDoc := &types.GenesisDoc{
		GenesisTime: time.Now(),
		ChainID:     chainID,
		Validators:  validators,
		AppHash:     []byte(chari.AppHash),
	}

	genesisRaw, err := json.MarshalIndent(genesisDoc, "", " ")
	if err != nil {
		logger.Error(err)
		return
	}

	if err = os.MkdirAll("channel-artifacts", 0766); err != nil {
		logger.Error(err)
		return
	}
	if err = writeToDisk("channel-artifacts/chari.genesis.json", genesisRaw); err != nil {
		logger.Error(err)
		return
	}

	for k, orderer := range orderers {
		raw, err := json.MarshalIndent(orderer, "", " ")
		if err != nil {
			logger.Error(err)
			return
		}

		if err = os.MkdirAll(fmt.Sprint("crypto-config/chari", k), 0766); err != nil {
			logger.Error(err)
			return
		}

		if err = writeToDisk(fmt.Sprint("crypto-config/chari", k, "/priv_validator.json"), raw); err != nil {
			logger.Error(err)
			return
		}
	}
}

func writeToDisk(fileName string, raw []byte) error {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	if _, err = file.Write(raw); err != nil {
		return err
	}
	if err = file.Sync(); err != nil {
		return err
	}
	return nil
}
