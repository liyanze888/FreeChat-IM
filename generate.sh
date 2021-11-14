#!/bin/sh

set -e

if ! which protoc-gen-go; then
  go get -u google.golang.org/protobuf/cmd/protoc-gen-go
fi

if ! which protoc-gen-go-grpc; then
  go get -u google.golang.org/grpc/cmd/protoc-gen-go-grpc
fi

if ! which sqlingo-gen-mysql; then
  go get -u github.com/lqs/sqlingo/sqlingo-gen-mysql
fi

if ! which go-localize; then
  go get -u github.com/m1/go-localize
fi

rm -rf generated
mkdir -p generated/grpc
echo "Generating proto"
FILES=`find -L im-proto -name '*.proto' |cut -sd / -f 2-`

MODULE=freechat/im
m="paths=source_relative,"
for file in $FILES; do
  DIR=`dirname ${file}`
  m="M${file}=$MODULE/generated/grpc/${DIR};`basename $DIR`pb,$m"
done

echo $m

protoc -I im-proto $FILES --go_out=generated/grpc --go_opt=$m --go-grpc_out=generated/grpc --go-grpc_opt=$m

mkdir -p generated/dsl
MYSQL="${MYSQL:-root:liyanze@3.1415926@tcp(152.136.28.100:3306)/freechat-im}"
sqlingo-gen-mysql $MYSQL > generated/dsl/dsl.go

#go-localize -input i18n -output generated/localizations



