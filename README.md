# go-devops - For CICD

## Local Development

### Environmental requirements

go1.25.0

### Development directory creation

```bash

# Create a development directory
mkdir godevops
cd godevops
```

### Get the code

```bash
# Get the code
git clone https://github.com/rouroumaibing/go-devops.git

```

### Startup instructions

### interactive terminal startup instructions

```
export RAFT_CLUSTER_ID=raft-cluster-prod
export CLUSTER_ID=cluster1-id
export MYSQL_DRIVERNAME=mysql
export MYSQL_USER=mysql
export MYSQL_PWD=******
export MYSQL_HOST=x.x.x.x
export MYSQL_PORT=32192
export MYSQL_DBNAME=CMDB
export MYSQL_MAXIDLE=1000
export MYSQL_MAXCONN=2000
export RAFT_PORT=7001
export REDIS_ADDRESS=x.x.x.x:32076
export REDIS_PWD=******
export WECHAT_APPID=wx1234567890
export WECHAT_APPSECRET=wx1234567890

go run cmd/go-devops/devops-server.go --tls-ca-file=certs/ca.crt --tls-cert-file=certs/server.crt --tls-key-file=certs/server.key
```

## ⭐ Star History

[![Star History Chart](https://api.star-history.com/svg?repos=rouroumaibing/go-devops&type=Date)](https://www.star-history.com/#rouroumaibing/go-devops&Date)

## 📄 License

MIT
