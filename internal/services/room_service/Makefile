.PHONY: test
test:
	go test -v ./...

.PHONY: run
run:
	go run cmd/termit.go

.PHONY: gen-proto
gen-proto:
	mkdir -p pb && protoc -I proto proto/room.proto --go_out=./pb --go_opt=paths=source_relative --go-grpc_out=./pb --go-grpc_opt=paths=source_relative