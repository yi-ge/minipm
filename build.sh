#/bin/bash
if [ ! -d bin ]; then
  mkdir bin
fi

echo "env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags \"-s -w\" -o ./bin/minipm-arm64 ."
env CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o ./bin/minipm-arm64 .
echo "env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags \"-s -w\" -o ./bin/minipm-amd64 ."
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o ./bin/minipm-amd64 .
