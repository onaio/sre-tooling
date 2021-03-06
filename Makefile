export GO111MODULE=on
GOPROXY ?= https://gocenter.io,direct
export GOPROXY

BIN := sre-tooling
DESTDIR :=
PREFIX := /usr/local

GOFLAGS := -v -mod=mod
EXTRA_GOFLAGS ?=

test:
	go test ./... -bench=. -v -covermode=atomic -coverprofile=cover.out

build: test
	go build $(GOFLAGS) $(EXTRA_GOFLAGS)

clean:
	go clean

install: build
	install -Dm755 ${BIN} ${DESTDIR}${PREFIX}/bin/${BIN}

uninstall:
	rm -f ${DESTDIR}${PREFIX}/bin/${BIN}
