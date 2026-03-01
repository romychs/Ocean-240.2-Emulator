#! /bin/bash

echo "Build OkEmu"
export GOROOT=/home/roma/projects/go/1.25.7/
export PATH=/home/roma/projects/go/1.25.7/bin/:$PATH

version=$(git describe --tags HEAD)
#commit=$(git rev-parse HEAD)
timestamp=$(date +%Y-%m-%d' '%H:%M:%S)

go build -ldflags "-X 'main.Version=$version' -X 'main.BuildTime=$timestamp'" .

#echo "Copy to kubelogtst01"
#scp stash boykovra@kubelogtst01:
