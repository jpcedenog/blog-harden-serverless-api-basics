.PHONY: build clean deploy

build:
	dep ensure -v
	env GOOS=linux go build -ldflags="-s -w" -o bin/createnote createnote/main.go
	env GOOS=linux go build -ldflags="-s -w" -o bin/scan scan/main.go

clean:
	rm -rf ./bin ./vendor Gopkg.lock

deploy: clean build
	sls deploy --verbose

remove:
	sls remove
