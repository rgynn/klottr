PACKAGE=klottr
VERSION=$(shell git rev-parse HEAD)
BUILDDATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
.PHONY: test test_intg run build build_docker run clean db_seed
test:
	go test ./...
test_intg:
	go run cmd/intg_test/*.go
run:
	go run main.go -env-files .env
build:
	go build -ldflags '-X github.com/rgynn/klottr/pkg/config.VERSION=${VERSION} -X github.com/rgynn/klottr/pkg/config.BUILDDATE=${BUILDDATE}' -o build/$(PACKAGE) .
build_docker:
	docker build -t $(PACKAGE) .
clean:
	rm -rf build
db_seed:
	go run cmd/seed/main.go