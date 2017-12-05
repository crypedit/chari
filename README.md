# Chari
Plugable consensus PBFT for hyperledger/fabric

Fabric    | Tests  | Date
----------|-------|-------
[9bf243e](https://github.com/hyperledger/fabric/commit/9bf243e0ed701b47bb5e255ceafd647fd7c26b72 "9bf243e")    |  pass|Nov 28, 2017

## Install
`go get -u github.com/amamina/chari`

## How to use
1. Add PBFT consenters into orderer, modify fabric/orderer/main.go

    **import** `"github.com/amamina/chari"`  

    **add**  `consenters["pbft"] = chari.New()`  after  
    
    ```go
    consenters["solo"] = solo.New()
    consenters["kafka"] = kafka.New(conf.Kafka.TLS, conf.Kafka.Retry, conf.Kafka.Version)
    ```
2. Support pbft in configtx.yaml, modify fabric/common/configtx/tool/provisional/provisional.go

    **add** `ConsensusTypePBFT = "pbft"` after
    
    ```go
    // ConsensusTypeSolo identifies the solo consensus implementation.
    ConsensusTypeSolo = "solo"
    // ConsensusTypeKafka identifies the Kafka-based consensus implementation.
    ConsensusTypeKafka = "kafka"
    ```
    
    **add** `case ConsensusTypePBFT: //do nothing` after
    
    ```go
    switch conf.Orderer.OrdererType {
    		case ConsensusTypeSolo:
    		case ConsensusTypeKafka:
    			bs.ordererGroups = append(bs.ordererGroups, config.TemplateKafkaBrokers(conf.Orderer.Kafka.Brokers))
    ```
3. Change consensus to PBFT in configtx.yaml

    `OrdererType: pbft`

4. Make docker image **hyperledger/fabric-orderer**
    Cause we have modify fabric/orderer/main.go, a new image hyperledger/fabric-orderer should be remake.
    There is a script to in chari/make, cd **chari/make** and $>make, it will auto to make image hyperledger/fabric-orderer:latest.

5. Configuration files
    Chari depends on [tendermint](https://github.com/tendermint/tendermint "tendermint"), for provide some configuration files  to startup tendermint, a tool called **charigen** in chari/charigen which will auto create configuration files.
`Notice: charigen should be in the same dir of configtxgen and cryptogen.`

6. Startup Orderer
    Chari supports PBFT instead of Kafka in orderer, so Kakfa no needs. Now, orderers must meet 3F +1,  and additional configurations likes below  should be added to orderer shell.
```yaml
-p 3450:3450 \
-e BFT_P2P_PORT=3450 \
 -e BFT_P2P_SEEDS=192.168.110.128:3450,192.168.110.128:3451,192.168.110.128:3452,192.168.110.128:3453 \
-v /etc/hyperledger/crypto-config/chari0:/etc/hyperledger/chari \
-v /etc/hyperledger/channel-artifacts/chari.genesis.json:/etc/hyperledger/chari/genesis.json \
```

## Warning
The current version, when orderer shutdown and restart, the one will be ejected by other orderers, cause the missing blocks during down time will re-execute after orderer restart, this operation will be regarded as do evil by other orderers. To avoid this issue, the down orderer should delete and re-startup but not restart. In the next version will fix this issue.
