package chari

import (
	"errors"
	"os"
	"strings"
)

const (
	defaultConfig string = `
# This is a TOML config file.
# For more information, see https://github.com/toml-lang/toml

proxy_app = "tcp://127.0.0.1:33453"
#moniker = "$moniker"
fast_sync = true
db_backend = "leveldb"
log_level = "state:error,*:error"

[consensus]
create_empty_blocks = false
max_block_size_txs = 500

[rpc]
laddr = "tcp://127.0.0.1:44453"

[p2p]
laddr = "tcp://0.0.0.0:$laddr_port"
seeds = "$seeds"
	`
)

const (
	proxy_app string = "127.0.0.1:33453"
	rpc_laddr string = "tcp://127.0.0.1:44453"
)

func initConfig() error {
	if os.Getenv("BFT_P2P_PORT") == "" {
		return errors.New("BFT_P2P_PORT not found")
	}
	if os.Getenv("BFT_P2P_SEEDS") == "" {
		return errors.New("BFT_P2P_SEEDS not found")
	}

	var config string
	config = strings.Replace(defaultConfig, "$laddr_port", os.Getenv("BFT_P2P_PORT"), 1)
	config = strings.Replace(config, "$seeds", os.Getenv("BFT_P2P_SEEDS"), 1)

	file, err := os.OpenFile("/etc/hyperledger/chari/config.toml", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer file.Close()

	if _, err = file.WriteString(config); err != nil {
		logger.Error(err)
		return err
	}
	return nil
}
