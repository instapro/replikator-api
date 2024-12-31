update:
	go get -u && go mod tidy

test:
	go test -race ./...

lint:
	golangci-lint run

install-tools:
	GO111MODULE=off go get -u golang.org/x/lint/golint
	GO111MODULE=off go get -u golang.org/x/tools/cmd/goimports
	GO111MODULE=off go get -u github.com/GeertJohan/fgt
	GO111MODULE=off go get -u github.com/kisielk/errcheck

go-mod:
	go mod tidy
	go mod verify

verify: go-mod coverage lint

coverage:
	go test -race -coverprofile=coverage.txt -covermode atomic ./...
	gocover-cobertura < coverage.txt > coverage.xml
