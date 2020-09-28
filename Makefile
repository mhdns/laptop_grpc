gen:
	protoc --proto_path=proto --go_out=plugins=grpc:. proto/*.proto
	# protoc --proto_path=proto --go_out=plugins=grpc:pb proto/*.proto

clean:
	rm pb/*.go

server:
	go run cmd/server/main.go -port 5000

client:
	go run cmd/client/main.go -address 0.0.0.0:5000

test:
	go test -cover -race ./...