GIT_VERSION=$(shell git rev-parse --short HEAD)
BUILD_TIME=$(shell date '+%FT%T')
PKG=github.com/9sheng/foobar-operator/pkg/version
LDFLAGS=-ldflags "-X ${PKG}.buildVersion=${GIT_VERSION} -X ${PKG}.buildTime=${BUILD_TIME}"

.PHONY: all image test
all:
	(cd cmd/foobar && go build -v ${LDFLAGS})

.PHONY: dep gen
dep:
	godep restore
	rm -rf Godeps vendor
	godep save ./...

gen:
	sh hack/update-codegen.sh
