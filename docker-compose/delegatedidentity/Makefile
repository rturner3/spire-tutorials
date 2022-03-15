all: build

.PHONY: proto caller callee proxy
build: caller callee proxy

caller:
	go build -o ./build/bin/caller ./cmd/caller

callee:
	go build -o ./build/bin/callee ./cmd/callee

proxy:
	go build -o ./build/bin/proxy ./cmd/proxy

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/hello/hello.proto

clean:
	rm ./build/bin/*

.PHONY: images
images: callee-image caller-image proxy-image spire-registrant

callee-image:
	docker build --target callee -t callee -f ./build/package/Dockerfile .

caller-image:
	docker build --target caller -t caller -f ./build/package/Dockerfile .

proxy-image:
	docker build --target proxy -t proxy -f ./build/package/Dockerfile .

spire-registrant-image:
	docker build --target spire-registrant -t spire-registrant -f ./build/package/Dockerfile .

.PHONY: env cleanenv
env:
	VAGRANT_CWD=deployments/ubuntu1 vagrant up
	VAGRANT_CWD=deployments/ubuntu2 vagrant up

cleanenv:
	VAGRANT_CWD=deployments/ubuntu1 vagrant destroy -f
	VAGRANT_CWD=deployments/ubuntu2 vagrant destroy -f
