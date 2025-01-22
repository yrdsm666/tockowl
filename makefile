build: build_client build_server build_remote
	@:

build_genkey:
	go build -o genkey ./cmd/genkey

build_client:
	go build -o client ./cmd/client

build_server:
	go build -o server ./cmd/server

build_remote:
	go build -o remote ./cmd/remote

win_bulid:
	go build -o client.exe ./cmd/client
	go build -o server.exe ./cmd/server

test:
	go clean -testcache
	go test -v ./...

compile_proto:
	protoc pb/*.proto --go_out=. --go-grpc_out=.

win_compile_proto:
	protoc --go_out=plugins=grpc:. pb/*.proto

run_server:
	./consensus

run_client:
	./client

generate_key: 
	./consensus

win_generate_key: 
	./genkey.exe -p keys/ -k 3 -l 4
