# Screensaver

A tiny flowing wave terminal screensaver written in Go.

![demo](assets/demo.gif)

## Features

Flowing ribbon wave animation spanning full screen width.
Elegant grey-silver-white color gradient.
Smooth 3D wireframe rendering.
Multiple sine waves for organic movement.
Responsive to terminal size.

## Requirements

Go 1.21 or later.
A terminal that supports Unicode and true color (24-bit).

## Installation

Clone the repository and build the binary.

```bash
git clone https://github.com/OlegChuev/screensaver.git
cd screensaver
make build
```

The binary will be available at `bin/screensaver`.

## Usage

```bash
./bin/screensaver
```

### Controls

Press `q`, `Q`, `Esc`, or `Ctrl+C` to quit.

## Development

The project includes a Makefile for common tasks.

```bash
# Build binary
make build

# Run application
make run

# Run tests
make test

# Lint code
make lint

# Create release binaries
make release

# Run demo
make demo

# Clean build artifacts
make clean
```

## How it works

The screensaver creates a flowing ribbon wave using multiple layered sine waves. The wave spans the full width of the terminal and animates smoothly from left to right. Colors transition through a grey-silver-white gradient based on wave height and layer depth, creating a metallic 3D effect.

## License

MIT License - see [LICENSE](LICENSE) for details.

---

Built with ❤️ and AI
