# revprox
Lightweight reverse proxy 1 port

## Installation

- linux-amd64

    ```bash
    $ curl -O https://storage.googleapis.com/acoshift/revprox
    $ sudo chmod +x revprox
    $ sudo mv revprox /usr/local/bin/
    ```

- go

    ```bash
    $ go get -u github.com/acoshift/revprox
    ```

## Build from source

```bash
$ git clone https://github.com/acoshift/revprox.git
$ cd revprox
$ make
```

## How to Run

```
$ revprox -addr=:8080 -target=http://localhost:9000
```

This command will start single host reverse proxy on address `:8080`.
