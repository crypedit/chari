# Copyright Greg Haskins All Rights Reserved
#
# SPDX-License-Identifier: Apache-2.0
#
FROM hyperledger/fabric-baseos:latest
ENV FABRIC_CFG_PATH /etc/hyperledger/fabric
RUN mkdir -p /var/hyperledger/production $FABRIC_CFG_PATH
RUN apt-get update && apt-get install -y libltdl7
COPY payload/orderer /usr/local/bin
ADD payload/sampleconfig.tar.bz2 $FABRIC_CFG_PATH/
EXPOSE 7050
CMD ["orderer"]
LABEL org.hyperledger.fabric.version=1.0.5 \
      org.hyperledger.fabric.base.version=0.3.2
