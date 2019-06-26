# turn on go mod flag so go mod doesn't download dependencies
ENVFLAGS = GOFLAGS=-mod=vendor

generate:
	$(ENVFLAGS) go generate ./resources/...

test: generate
	$(ENVFLAGS) go test ./...

build: test
	$(ENVFLAGS) go build -o secureworks

