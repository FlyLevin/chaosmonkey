test:
	go test -v -cover ./...

lint:
	go vet ./...
	golint -set_exit_status ./...

deps:
	go get \
		github.com/golang/lint/golint \
		github.com/ryanuber/columnize
