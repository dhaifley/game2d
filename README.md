# game2d
2D gaming framework

This is a basic framework providing an interface for accessing the ebitengine
2D game engine. It is implemented using a declarative schema with embedded Lua
scripting and game assets.

## Requirements

* [go](https://go.dev/dl/)
* [docker](https://docs.docker.com/get-docker/)
* [make](https://www.gnu.org/software/make/)

## Building and Testing

To build the service locally:

```sh
$ make clean
$ make build
```

To run just the unit tests, which do not start any test containers:

```sh
$ make test-quick
```

To run all tests, including integration tests, which start test containers:

```sh
$ make test
```

To start the test environment containers locally:

```sh
$ make start
```

To run the service locally, for testing:

```sh
$ make run
```

Finally, to shutdown and cleanup the test environment:

```sh
$ make stop
```

## Documentation

While the service is running locally, interactive documentation, which can be
used for testing requests to the service, can be accessed using:
* http://localhost:8080/api/v1/docs
