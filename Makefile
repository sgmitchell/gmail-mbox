fmt:
	go mod tidy
	golangci-lint run --fix -p format

build: fmt
	go build -o gmail-mbox .


