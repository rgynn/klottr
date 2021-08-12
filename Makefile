PACKAGE=klottr
.PHONY: test run build build_docker clean
test:
	go test ./...
test_int:
	go run cmd/intg_test/main.go
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