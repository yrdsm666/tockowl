# old version
# protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative Tockowl.proto
protoc -I=. --go_out=.. Tockowl.proto
