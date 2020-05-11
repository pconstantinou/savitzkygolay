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

doc:
	${GOPATH}/bin/autoreadme -f