#!/bin/bash

set -o xtrace

arch=$(go env GOARCH)
os=$(go env GOOS)
version="2.3.1"
targetPath="/usr/local/kubebuilder"

# download the release
#curl -L -O "https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${version}/kubebuilder_${version}_linux_${arch}.tar.gz"
curl -L https://go.kubebuilder.io/dl/${version}/${os}/${arch} | tar -xz -C /tmp/

sudo mv /tmp/kubebuilder_${version}_${os}_${arch} ${targetPath}

# update your PATH to include /usr/local/kubebuilder/bin
export PATH=$PATH:$targetPath/bin
