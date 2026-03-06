APP := naptd

.PHONY: build test run

build:
	go build -o bin/$(APP) ./prototype/cmd/naptd

test:
	go test ./prototype/...

run:
	go run ./prototype/cmd/naptd --wan-ip 172.31.10.1
