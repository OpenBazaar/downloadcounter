counter:
	docker run --rm -v "$$PWD":/go/src/github.com/OpenBazaar/downloadcounter -w /go/src/github.com/OpenBazaar/downloadcounter/bin iron/go:dev go build -o ../counter

docker: counter
	docker build -t openbazaarproject/downloadcounter .

all: docker