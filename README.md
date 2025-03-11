# empty
A game of emptiness.

This is an empty framework interface for the ebitengine 2D game engine.
It is implemented using a declarative JSON schema with embedded Lua scripting.

## Requirements

* [go](https://go.dev/dl/)

## Building and Testing

To build the application:

```sh
$ go build ./cmd/empty
```

To run the unit tests:

```sh
$ go test -cover -race ./...
```

To run the application:

```sh
$ go run ./cmd/empty
```
