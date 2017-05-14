build:
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o build/revprox -a -ldflags '-w -s' cmd/revprox/main.go
	chmod +x build/revprox

dev:
	go run cmd/revprox/main.go

clean:
	rm -rf build/

compress:
	tar czf build/revprox.tar.gz -C build revprox

upload:
	gsutil -h "Cache-Control: public, max-age=30" cp -a public-read build/revprox.tar.gz gs://acoshift/

deploy:	clean build compress upload
