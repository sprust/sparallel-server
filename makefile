GOLANG_CLI="docker-compose exec golang"

bash:
	"$(GOLANG_CLI)" bash

linter:
	golangci-lint run

run:
	go run ./cmd/server/main.go ${c}

build:
	CGO_ENABLED=0 GOOS=linux go build -v -a -o ./bin/sparallel_server ./cmd/server/main.go \
		&& chmod +x ./bin/sparallel_server

bin-start:
	./bin/sparallel_server ${c}
