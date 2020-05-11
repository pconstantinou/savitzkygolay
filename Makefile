all: get build test doc

clean:
	rm *.png *.out

get:
	go get .
	go get github.com/jimmyfrasche/autoreadme

build:
	go build

test:
	go test -coverprofile=coverage.out

coverage: test
	curl -s https://codecov.io/bash | bash


doc: coverage
	${GOPATH}/bin/autoreadme -f  -template=README.md.template

