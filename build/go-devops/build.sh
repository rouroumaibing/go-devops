#!/bin/bash
PROJECT_ROOT=$(cd `dirname $0/`/../../;pwd)
VERSIONDIR="github.com/rouroumaibing/go-devops/pkg/apis/versions"
OUTPUTDIR="${PROJECT_ROOT}/output"
component=go-devops
version=${1:-"1.0.0"}

gitTag=$(if [ "`git describe --tags --abbrev=0 2>/dev/null`" != "" ];then git describe --tags --abbrev=0; else git log --pretty=format:'%h' -n 1; fi)
buildDate=$(TZ=Asia/Shanghai date +%FT%T%z)
gitCommit=$(git log --pretty=format:'%H' -n 1)
gitBranch=$(git rev-parse --abbrev-ref HEAD)

# 待上库前清理
gitTag="tagtest"
gitCommit="committest"
gitBranch="branchtest"

ldflags="-X ${VERSIONDIR}.version=${version} -X ${VERSIONDIR}.gitTag=${gitTag} -X ${VERSIONDIR}.buildDate=${buildDate} -X ${VERSIONDIR}.gitBranch=${gitBranch} -X ${VERSIONDIR}.gitCommit=${gitCommit}"

function prepare_go_mod(){
    export GO111MODULE=on
    export GOPROXY=https://goproxy.cn,direct
    export GONOSUMDB=*
    export GOSUMDB=off
}

function prepare_build_file(){
    mkdir -p ${OUTPUTDIR}
    cp -r ${PROJECT_ROOT}/build/${component}/charts ${OUTPUTDIR}/
    cp -r ${PROJECT_ROOT}/build/${component}/images ${OUTPUTDIR}/
}

function build_go_devops(){

    pushd "${PROJECT_ROOT}/cmd/${component}"
    GOOS=linux GOARCH=amd64 go build -o ${component} -ldflags "${ldflags}"
    
    if [ -f ./${component} ]; then
        tar -zcvf ${OUTPUTDIR}/images/${component}.tar.gz ./${component} ./conf
        echo "${component} build successfully."
        rm ./${component}
    else
        echo  "${component} build failed."
        exit 1
    fi
    popd 
}

function build_go_devops_docker_image(){
    pushd "${OUTPUTDIR}/images"
    docker build --network host . -t ${component}:${version}
    docker save -o ${component}-${version}.tar ${component}:${version} 
    if [ -f ./${component}-${version}.tar ]; then
        echo "${component} build image successfully."
        rm ./${component}.tar.gz
        rm ./Dockerfile*
        docker rmi ${component}:${version}
    else
        echo "${component} build image failed."
    fi
    popd
}

function charts_pack(){
    pushd "${OUTPUTDIR}/charts"
    tar -zcvf ${OUTPUTDIR}/charts/${component}-${version}.tgz ${component}
    if [ -f ./${component}-${version}.tgz ]; then
        echo "${component} charts pack successfully."
        rm -r ./${component}
    else
        echo "${component} charts pack failed."
    fi

    popd
}

function pack(){
    pushd "${OUTPUTDIR}"
    tar -zcvf ${component}-${version}.tar.gz charts images
    if [ -f ./${component}-${version}.tar.gz ]; then
        echo "${component} pack successfully."
        rm -r charts images
    else
        echo "${component} pack failed."
    fi

    popd
}

function clean(){
    rm -rf ${OUTPUTDIR}
}


clean
prepare_go_mod
prepare_build_file
build_go_devops
build_go_devops_docker_image
charts_pack
pack
