# Zenoh Go Binding

This repository contains the Go bindings for Zenoh. Zenoh is a protocol that unifies data in motion, data at rest, and computations. This project allows you to use Zenoh in your Go applications.

## Dependencies

Before building and running the examples, you need to have the following dependencies installed:

- **Zenoh-C**: The C implementation of the Zenoh protocol.

### Installing Zenoh-C

You can follow the instructions provided in the [Zenoh-C repository](https://github.com/eclipse-zenoh/zenoh-c) to install zenoh-c.
It is required that zenoh-c is built with unstable features support (i.e. with -DZENOHC_BUILD_WITH_UNSTABLE_API=ON cmake flag).

## Building the Examples

This project includes several examples located in the `examples` directory. Each example is in a subdirectory prefixed with `z_`. You can build all the examples using the provided `Makefile`.

### Build All Examples

To build all examples, simply run:

```bash
make all
```

This command will compile all the examples and place the binaries in the `bin` directory.

### Build a Specific Example

To build a specific example, use the example's name. For instance, to build the `z_pub` example:

```bash
make z_pub
```

## Running the Examples

After building the examples, you can run them from the `bin` directory.
Description of each example can be found [here](./examples/README.md).

### Run a Specific Example

To run a specific example, navigate to the `bin` directory and execute the binary. For example, to run the `z_pub` example:

```bash
./bin/z_pub
```

### Run Examples Directly with go run ###

You can also run the examples directly using go run without building the binaries. For example, to run the z_sub example:

```bash
go run examples/z_pub/z_pub.go
```

## Project Structure

- `examples/`: This directory contains all the example subdirectories. Each example has its own subdirectory prefixed with `z_`.
- `bin/`: This directory will contain the compiled binaries for the examples.

## Makefile

The provided `Makefile` includes the following targets:

- `all`: Builds all the examples.
- `fmt`: Formats the source code using `go fmt`.
- `clean`: Cleans up all generated binaries.

### Makefile Usage

- **Build all binaries**: `make all`
- **Build specific binary**: `make <example_name>`
- **Format the source code**: `make fmt`
- **Clean up binaries**: `make clean`
