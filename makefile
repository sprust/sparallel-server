GOLANG_CLI="docker-compose run -it --rm --user $$(id -u):$$(id -g) golang"

bash:
	"$(GOLANG_CLI)" bash

run:
	go run ./cmd/server/main.go ${c}

build:
	CGO_ENABLED=0 GOOS=linux go build -v -a -o ./bin/sparallel_server ./cmd/server/main.go \
		&& chmod +x ./bin/sparallel_server

bin-start:
	./bin/sparallel_server ${c}
