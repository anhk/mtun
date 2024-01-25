export GOPROXY=goproxy.cn,direct

OBJ = mtun

all: $(OBJ)

# Build manager binary
$(OBJ):
	go build -o $(OBJ) ./main/

clean:
	rm -fr $(OBJ)

-include .deps

generate:
	protoc --go_out=./ --go_opt=paths=source_relative --go-grpc_out=./ --go-grpc_opt=paths=source_relative proto/stream.proto

dep:
	echo '$(OBJ): \\'> .deps
	find . -path ./vendor -prune -o -name '*.go' -print | awk '{print $$0 " \\"}' >> .deps
	echo "" >> .deps