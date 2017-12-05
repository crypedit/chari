wd := $(PWD)

.PHONY: orderer-docker
orderer-docker:
	rm -rf payload
	mkdir -p payload
	rm -f $(GOPATH)/src/github.com/hyperledger/fabric/build/bin/orderer
	$(MAKE) -C $(GOPATH)/src/github.com/hyperledger/fabric orderer
	cp $(GOPATH)/src/github.com/hyperledger/fabric/build/bin/orderer $(wd)/payload
	cd $(GOPATH)/src/github.com/hyperledger/fabric/sampleconfig && tar jcvf sampleconfig.tar.bz2 * && mv sampleconfig.tar.bz2 $(wd)/payload
	cp /opt/tendermint/tendermint payload/
	docker build -t hyperledger/fabric-orderer:latest .