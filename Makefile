.PHONY: image shell deps build run

image:
	docker build -t dockercp .

shell:
	docker run --rm -it --name dockercp \
	  -v /var/run/docker.sock:/host/var/run/docker.sock \
	  -v "`pwd`:/go/src/github.com/divolgin/dockercp" \
	  dockercp

deps:
	govendor install +local

build:
	go build -o ./dockercp .

run:
	./dockercp
