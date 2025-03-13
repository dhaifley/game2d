# game2d
2D gaming framework

This is a basic framework providing an interface for accessing the ebitengine
2D game engine. It is implemented using a declarative schema with embedded Lua
scripting and game assets.

## Requirements

* [go](https://go.dev/dl/)

## Building and Testing

To build the application:

```sh
$ go build ./cmd/game2d
```

To run the unit tests:

```sh
$ go test -cover -race ./...
```

To run the application:

```sh
$ go run ./cmd/game2d
```
