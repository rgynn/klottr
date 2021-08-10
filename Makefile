PACKAGE=klottr
.PHONY: test run build build_docker clean
test:
	go test ./...
run:
	go run main.go
build:
	go build -o build/$(PACKAGE) .
build_docker:
	docker build -t $(PACKAGE) .
clean:
	rm -rf build
seed:
	go run cmd/seed/main.go