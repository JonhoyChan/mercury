#!/bin/bash

DEFAULT_PROTOC_GEN="go"
DEFAULT_PROTOC="protoc"

function _install_protoc() {
    osname=$(uname -s)
    echo "install protoc ..."
    case $osname in
        "Darwin" )
            brew install protobuf
            ;;
        *)
            echo "unknown operating system, need install protobuf manual see: https://developers.google.com/protocol-buffers"
            exit 1
            ;;
    esac
}

function _install_protoc_gen() {
    local protoc_gen=$1
    case $protoc_gen in
        "gofast" )
            echo "install protoc-gen-gofast from github.com/gogo/protobuf/protoc-gen-gofast"
            go get -u github.com/gogo/protobuf/protoc-gen-gofast
            ;;
        "gogofast" )
            echo "install protoc-gen-gogofast from github.com/gogo/protobuf/protoc-gen-gogofast"
            go get -u github.com/gogo/protobuf/protoc-gen-gogofast
            ;;
        "gogo" )
            echo "install protoc-gen-gogo from github.com/gogo/protobuf/protoc-gen-gogo"
            go get -u github.com/gogo/protobuf/protoc-gen-gogo
            ;;
        "go" )
            echo "install protoc-gen-go from github.com/golang/protobuf"
            go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
            ;;
        *)
            echo "can't install protoc-gen-${protoc_gen} automatic !"
            exit 1;
            ;;
    esac
}

function _install_protoc_gen_micro() {
    echo "install protoc-gen-micro from github.com/micro/protoc-gen-micro/v2"
    go get -u github.com/micro/protoc-gen-micro/v2
}

if [[ -z $PROTOC ]]; then
    PROTOC=${DEFAULT_PROTOC}
    which $PROTOC
    if [[ "$?" -ne "0" ]]; then
        _install_protoc
    fi
fi

if [[ -z $PROTOC_GEN ]]; then
    PROTOC_GEN=${DEFAULT_PROTOC_GEN}
    which protoc-gen-$PROTOC_GEN
    if [[ "$?" -ne "0" ]]; then
        _install_protoc_gen $PROTOC_GEN
    fi
fi

which protoc-gen-micro
if [[ "$?" -ne "0" ]]; then
    _install_protoc_gen_micro
fi