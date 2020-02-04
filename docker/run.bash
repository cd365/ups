#/bin/bash

yum -y install wget

wget -c -P /usr/local https://studygolang.com/dl/golang/go1.13.7.linux-amd64.tar.gz

tar zxf /usr/local/go1.13.7.linux-amd64.tar.gz -C /usr/local

rm -f /usr/local/go1.13.7.linux-amd64.tar.gz

ln -s /usr/local/go/bin/* /usr/local/sbin/

yum -y install git

go mod tidy

go build -ldflags '-w -s' -o /usr/local/bin/ups

yum -y remove wget git

