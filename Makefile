IMAGE=acoshift/revprox
TAG=1.0.0
GOLANG_VERSION=1.8
REPO=github.com/acoshift/revprox

revprox: main.go
	go get -v
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -ldflags '-w -s' -o revprox ./main.go

build:
	docker pull golang:$(GOLANG_VERSION)
	docker run --rm -it -v $(PWD):/go/src/$(REPO) -w /go/src/$(REPO) golang:$(GOLANG_VERSION) /bin/bash -c "make revprox"
	docker build --pull -t $(IMAGE):$(TAG) .

build-linux:
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o revprox -a -ldflags '-w -s' main.go
	chmod +x revprox

push: clean build
	docker push $(IMAGE):$(TAG)

dev:
	go run main.go

clean:
	rm -f revprox revprox.tar.gz

compress:
	tar czf revprox.tar.gz revprox

upload:
	gsutil -h "Cache-Control: public, max-age=30" cp -a public-read revprox.tar.gz gs://acoshift/

deploy-gcs:	clean build-linux compress upload
