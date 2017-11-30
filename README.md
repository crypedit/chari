# Chari
Plugable consensus PBFT for hyperledger/fabric

Fabric    | Tests  | Date
----------|-------|-------
[9bf243e](https://github.com/hyperledger/fabric/commit/9bf243e0ed701b47bb5e255ceafd647fd7c26b72 "9bf243e")    |  pass|Nov 28, 2017

## Install
`go get -u github.com/amamina/chari`

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



