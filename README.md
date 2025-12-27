# Screensaver

A tiny flowing wave terminal screensaver written in Go. 🌊

![demo](demo.gif)

## Features

- Flowing ribbon wave animation spanning full screen width
- Beautiful blue-pink-purple color gradient
- Smooth 3D wireframe rendering
- Multiple sine waves for organic movement
- Responsive to terminal size

## Requirements

- Go 1.21 or later
- A terminal that supports Unicode and true color (24-bit)

## Installation

### From source

```bash
git clone https://github.com/OlegChuev/screensaver.git
cd screensaver
make build
```

### Install to PATH

```bash
make install
```

## Usage

```bash
# Run directly
./screensaver

# Or if installed
screensaver
```

### Controls

- `q` or `Q` - Quit
- `Esc` - Quit
- `Ctrl+C` - Quit

## Build

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run without building
make run

# Clean build artifacts
make clean
```

## Project Structure

```
.
├── main.go                 # Entry point
├── internal/
│   ├── app/               # Application lifecycle & event handling
│   │   └── app.go
│   ├── renderer/          # 3D to 2D projection & drawing
│   │   └── renderer.go
│   └── wave/              # Wave generation & animation
│       └── surface.go
├── Makefile
└── README.md
```

## How it works

The screensaver creates a flowing ribbon wave using multiple layered sine waves. The wave spans the full width of the terminal and animates smoothly from left to right. Colors transition through a blue-pink-purple gradient based on wave height and layer depth, creating a vibrant 3D effect.

## License

MIT License - see [LICENSE](LICENSE) for details.
