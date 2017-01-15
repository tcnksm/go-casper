#!/bin/bash
set -e

DIR=$(cd $(dirname ${0}) && pwd)
cd $DIR

echo "Generate CA pem"
openssl genrsa -out rootCA.key 2048
openssl req -x509 -sha256 -new -nodes -key rootCA.key -days 1024 -out rootCA.pem

echo "Generate self signed crts"
openssl genrsa -out server.key 2048
openssl req -new -key server.key -out server.csr -sha256
openssl x509 -req -in server.csr -CA rootCA.pem -CAkey rootCA.key -CAcreateserial -out server.crt -days 500 -sha256
rm crts/rootCA.srl