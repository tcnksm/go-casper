# Example

```bash
$ openssl genrsa -out rootCA.key 2048
$ openssl req -x509 -sha256 -new -nodes -key rootCA.key -days 1024 -out rootCA.pem
```

```bash
$ openssl genrsa -out server.key 2048
$ openssl req -new -key server.key -out server.csr -sha256
$ openssl x509 -req -in server.csr -CA rootCA.pem -CAkey rootCA.key -CAcreateserial -out server.crt -days 500 -sha256
```
