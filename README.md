# game2d
2D gaming framework

This is a basic framework providing an interface for accessing the ebitengine
2D game engine and a Lua scripting interpreter. It is implemented using a
declarative schema that can be rendered in JSON or YAML. It is intended to be
produced and consumed by generative A.I.

## Overview

The game2d service consists of three primary components.

### game2d

This is the main client and protocol, which is usually compiled into WASM for
distribution via the application user interface. It can also be built for
various native architectures.

### game2d-api

This is the back-end service providing the REST API and the main service
functionality. It provides the interface with generative A.I. services. It also
serves the application user interface.

### game2d-app

This is a simple user interface application for interacting with the service.
It is used to list, select and save game definitions, update user profiles and
settings, and submit game definitions and prompts to generative A.I. services.
The primary portion of the user interface runs a WASM version of the client
directly within the user interface.

## Getting Started

### Requirements

- [go](https://go.dev/dl/)
- [docker](https://docs.docker.com/get-docker/)
- [make](https://www.gnu.org/software/make/)
- [node.js](https://nodejs.org/)
- [npm](https://www.npmjs.com/)
- [typescript](https://www.typescriptlang.org/)

### Other Tools and Libraries

- [ebitengine](https://ebitengine.org/)
- [lua](https://www.lua.org/)
- [wasm](https://webassembly.org/)
- [react](https://react.dev/)
- [react router](https://reactrouter.com/)
- [vite](https://vite.dev/)

### Building and Testing

To build the service components locally:

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

### Documentation

While the service is running locally, interactive documentation, which can be
used for testing requests to the service, can be accessed using:

- http://localhost:8080/api/v1/docs

### Running

While the service is running locally, the user interface application can be
accessed using:

- http://localhost:8080/
