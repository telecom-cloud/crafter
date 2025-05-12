#! /usr/bin/env bash

# const value define
moduleName="github.com/telecom-cloud/tools/example"
project="../example"
curDir=`pwd`
protobuf2IDL=$curDir"/testdata/protobuf2/psm/psm.proto"
proto2Search=$curDir"/testdata/protobuf2"
protobuf3IDL=$curDir"/testdata/protobuf3/psm/psm.proto"
proto3Search=$curDir"/testdata/protobuf3"
protoSearch="/usr/local/include"

judge_exit() {
  code=$1
  if [ $code != 0 ]; then
    exit $code
  fi
}

compile_cft() {
  go build -o cft cmd/cft/cft.go
  judge_exit "$?"
  mv cft $GOPATH/bin/
}

install_dependent_tools() {
  PROTOC_VERSION=25.1
  OS=$(uname)
  ARCH=$(uname -m)
  PROTOC_DIR="protoc-$PROTOC_VERSION-linux-$ARCH"
  if [ "$OS" == "Darwin" ]; then
      PROTOC_DIR="protoc-$PROTOC_VERSION-osx-$ARCH"
  fi

  if [ -f "/usr/local/bin/protoc" ]; then
      echo "protoc file already exists. skipping download."
      return
  fi

  echo "downloading protoc"
  # install protoc
  wget "https://github.com/protocolbuffers/protobuf/releases/download/v$PROTOC_VERSION/$PROTOC_DIR.zip"
  unzip -d $PROTOC_DIR $PROTOC_DIR.zip
  rm -f $PROTOC_DIR.zip
  cp $PROTOC_DIR/bin/protoc /usr/local/bin/protoc
  cp -r $PROTOC_DIR/include/google /usr/local/include/google
}

test_protobuf2() {
  # test protobuf2
  mkdir -p example
  cd example
  cft new -I=$protoSearch -I=$proto2Search --idl=$protobuf2IDL --mod=$moduleName -f --model_dir=model --handler_dir=handler --router_dir=router
  judge_exit "$?"
  go mod tidy && go build .
  judge_exit "$?"
  cft update -I=$protoSearch -I=$proto2Search --idl=$protobuf2IDL
  judge_exit "$?"
  cft model -I=$protoSearch -I=$proto2Search --idl=$protobuf2IDL --model_dir=model
  judge_exit "$?"
  cft client -I=$protoSearch -I=$proto2Search --idl=$protobuf2IDL --client_dir=client
  judge_exit "$?"
  cd ..
  rm -rf example
}

test_protobuf3() {
  # test protobuf3
  mkdir -p $project
  cd $project
  cft new -I=$protoSearch -I=$proto3Search --idl=$protobuf3IDL --mod=$moduleName -f --model_dir=model --handler_dir=handler --router_dir=router
  judge_exit "$?"
  go mod tidy && go build .
  judge_exit "$?"
  ../cft update -I=$protoSearch -I=$proto3Search --idl=$protobuf3IDL
  judge_exit "$?"
  ../cft model -I=$protoSearch -I=$proto3Search --idl=$protobuf3IDL --model_dir=model
  judge_exit "$?"
  cft client -I=$protoSearch -I=$proto3Search --idl=$protobuf3IDL --mod=$moduleName --model_dir=model --client_dir=service --service_group=eci
  judge_exit "$?"
  cd ..
#  rm -rf example
}

main() {
  compile_cft
  judge_exit "$?"
  install_dependent_tools
  judge_exit "$?"
#  echo "test protobuf2......"
#  test_protobuf2
#  judge_exit "$?"
#  echo "test protobuf3......"
#  test_protobuf3
#  judge_exit "$?"
  echo "cft execute success"
}
main
