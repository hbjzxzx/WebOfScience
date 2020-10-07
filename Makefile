gen:
	protoc --proto_path=proto proto/*.proto --go_out=plugins=grpc:pb

clean:
	rm pb/*.go

runs:
	go run cmd/server/server.go --address 0.0.0.0 --port 2233 --spub cert/server.pem --skey cert/server.key 

runc:
	go run cmd/client/client.go --laddress 0.0.0.0 --lport 7891 --address ltest.ts --port 2233 --pub cert/server.pem

win:
	CGO_ENABLED=0;GOOS=windows; GOARCH=amd64; go build -o build/winClient.exe win10/main.go

linux:
	go build -o build/linxuClient cmd/client/client.go