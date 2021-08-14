PACKAGE=klottr
.PHONY: test test_intg run build build_docker clean db_seed
test:
	go test ./...
test_intg:
	go run cmd/intg_test/*.go
run:
	go run main.go
build:
	go build -o build/$(PACKAGE) .
build_docker:
	docker build -t $(PACKAGE) .
clean:
	rm -rf build
db_seed:
	go run cmd/seed/main.go