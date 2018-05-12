VERSION=`git describe --tags`
BUILD_DATE=`date +%FT%T%z`

BINARY=goJASM
LDFLAGS="-X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE}"

TARGETS="linux/386 linux/amd64 linux/arm darwin/amd64 windows/amd64"
CROSS_OUT="out/gojasm-{{.OS}}-{{.Arch}}"

.PHONY: clean

build:
	go build -ldflags ${LDFLAGS} -o ${BINARY}

cross:
	gox -ldflags=${LDFLAGS} -osarch=${TARGETS} -output=${CROSS_OUT}

clean:
	rm -rf ${BINARY}
	rm -rf out/