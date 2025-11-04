#!/bin/bash
component=${1:-go-devops}
version=${2:-1.0.0}
namespace=${3:-devops}

tar -zxf ${component}-${version}.tar.gz
cd images
docker load -i ${component}-${version}.tar
docker tag ${component}:${version} swr.cn-north-4.myhuaweicloud.com/testapp/${component}:${version}
docker push swr.cn-north-4.myhuaweicloud.com/testapp/${component}:${version}
cd ..
cd charts
rm -rf /root/.docker
helm install ${component} ${component}-${version}.tgz  --namespace=${namespace} --set image.imageAddr=swr.cn-north-4.myhuaweicloud.com/testapp/${component}:${version} --set image.pullPolicy=Always
