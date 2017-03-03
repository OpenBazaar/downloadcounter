counter:
	docker run --rm -v "$$PWD":/go/src/github.com/OpenBazaar/downloadcounter -w /go/src/github.com/OpenBazaar/downloadcounter/bin iron/go:dev go build -o ../counter

docker: counter
	docker build -t openbazaarproject/downloadcounter .

docker_hub:
	docker login -u="openbazaarproject" -p="$$DOCKERHUB_KEY"
	docker push openbazaarproject/downloadcounter

all: docker