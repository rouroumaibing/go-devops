#!/bin/bash
component=${1:-go-devops}
version=${2:-1.0.0}
namespace=${3:-devops}

helm uninstall ${component} -n ${namespace}
rm -rf /root/.docker ${component}-${version}.tar.gz charts/ images/ 
docker rmi ${component}:${version} swr.cn-north-4.myhuaweicloud.com/testapp/${component}:${version}
