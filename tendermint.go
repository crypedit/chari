package chari

import (
	"fmt"

	"github.com/op/go-logging"
	"github.com/spf13/viper"
	cmd "github.com/tendermint/tendermint/cmd/tendermint/commands"
	nm "github.com/tendermint/tendermint/node"
	"github.com/tendermint/tmlibs/log"
)

var loggerTendermint = logging.MustGetLogger("tendermint")

const rootDir = "/etc/hyperledger/chari"

type Log struct {
	logger *logging.Logger
	kvs    string
}

func startTendermint() {
	logging.SetLevel(logging.ERROR, "tendermint")
	viper.SetConfigName("config")
	viper.AddConfigPath(rootDir)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	config, err := cmd.ParseConfig()
	if err != nil {
		panic(err)
	}
	config.SetRoot(rootDir)

	node, err := nm.DefaultNewNode(config, &Log{logger: loggerTendermint}) //log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	if err != nil {
		panic(err)
	}

	if _, err = node.Start(); err != nil {
		panic(err)
	}
	defer node.Stop()

	wait := make(chan int)
	<-wait
}

func (this *Log) out(level string, msg string, keyvals ...interface{}) {
	var kvs string
	if len(keyvals) != 0 {
		if len(keyvals)%2 != 0 {
			keyvals = append(keyvals, "nil")
		}
		loop := len(keyvals) / 2
		for k := 0; k < loop; k++ {
			kvs = fmt.Sprint(kvs, " ", keyvals[k*2], "=", keyvals[k*2+1])
		}
	}

	switch level {
	case "Info":
		this.logger.Info(msg, this.kvs, kvs)
	case "Debug":
		this.logger.Debug(msg, this.kvs, kvs)
	case "Error":
		this.logger.Error(msg, this.kvs, kvs)
	}
}

func (this *Log) Info(msg string, keyvals ...interface{}) {
	this.out("Info", msg, keyvals)
}
func (this *Log) Debug(msg string, keyvals ...interface{}) {
	this.out("Debug", msg, keyvals)
}
func (this *Log) Error(msg string, keyvals ...interface{}) {
	this.out("Error", msg, keyvals)
}

func (this *Log) With(keyvals ...interface{}) log.Logger {
	var kvs string
	if len(keyvals) != 0 {
		if len(keyvals)%2 != 0 {
			keyvals = append(keyvals, "nil")
		}
		loop := len(keyvals) / 2
		for k := 0; k < loop; k++ {
			kvs = fmt.Sprint(kvs, " ", keyvals[k*2], "=", keyvals[k*2+1])
		}
	}
	return &Log{logger: this.logger, kvs: kvs}
}
