IMAGE=acoshift/revprox
TAG=1.3.0
GOLANG_VERSION=1.11
REPO=github.com/acoshift/revprox

revprox:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-w -s' -o revprox cmd/revprox/main.go

build-docker:
	docker pull golang:$(GOLANG_VERSION)
	docker run --rm -it -v $(PWD):/go/src/$(REPO) -w /go/src/$(REPO) golang:$(GOLANG_VERSION) /bin/bash -c "make revprox"
	docker build --pull -t $(IMAGE):$(TAG) .

build-linux:
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o revprox -a -ldflags '-w -s' cmd/revprox/main.go
	chmod +x revprox

push-docker: clean build-docker
	docker push $(IMAGE):$(TAG)

dev:
	go run cmd/revprox/main.go

clean:
	rm -f revprox
