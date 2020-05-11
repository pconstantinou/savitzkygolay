all: get build test

clean:
	rm *.png *.out

get:
	go get .

build:
	go build

test:
	go test -coverprofile=coverage.out