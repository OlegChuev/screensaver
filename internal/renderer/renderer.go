// Package renderer provides 3D to 2D projection and terminal rendering for the wave screensaver.
package renderer

import (
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/olegchuev/screensaver/internal/wave"
)

// ASCII characters for 3D shading effect - from darkest/furthest to brightest/closest
var shadeChars = []rune{' ', '.', ':', '-', '=', '+', '*', '#', '%', '@'}

// Block characters for filled areas
var blockChars = []rune{'░', '▒', '▓', '█'}

const (
	scaleXFactor = 0.95
	scaleYFactor = 0.7
	perspectiveY = 0.4
	depthZFactor = 0.3
)

// Renderer handles 3D to 2D projection and drawing to the terminal screen.
type Renderer struct {
	screen  tcell.Screen
	width   int
	height  int
	buffer  [][]cell
	centerX float64
	centerY float64
}

// cell represents a single terminal cell with character, style, and depth information.
type cell struct {
	char  rune
	style tcell.Style
	depth float64
	set   bool
}

// NewRenderer creates a new renderer attached to the given tcell screen.
func NewRenderer(screen tcell.Screen) *Renderer {
	w, h := screen.Size()
	r := &Renderer{
		screen:  screen,
		width:   w,
		height:  h,
		centerX: float64(w) / 2,
		centerY: float64(h) / 2,
	}
	r.initBuffer()
	return r
}

// initBuffer allocates the internal rendering buffer matching screen dimensions.
func (r *Renderer) initBuffer() {
	r.buffer = make([][]cell, r.height)
	for i := range r.buffer {
		r.buffer[i] = make([]cell, r.width)
	}
}

// Resize handles terminal resize events by updating dimensions and reallocating buffers.
func (r *Renderer) Resize() {
	r.width, r.height = r.screen.Size()
	r.centerX = float64(r.width) / 2
	r.centerY = float64(r.height) / 2
	r.initBuffer()
}

// Clear clears the rendering buffer and screen, preparing for a new frame.
func (r *Renderer) Clear() {
	for y := range r.buffer {
		for x := range r.buffer[y] {
			r.buffer[y][x] = cell{depth: -math.MaxFloat64}
		}
	}
	r.screen.Clear()
}

// project3D converts a 3D point to 2D screen coordinates with depth for z-ordering.
func (r *Renderer) project3D(p wave.Point3D) (int, int, float64) {
	// Scale to fill the screen width
	scaleX := float64(r.width) * scaleXFactor
	scaleY := float64(r.height) * scaleYFactor

	// Project X directly (horizontal position)
	screenX := int(r.centerX + p.X*scaleX)

	// Project Y and Z combined for vertical position
	// Z (wave height) affects vertical position, Y (depth) adds perspective
	screenY := int(r.centerY - p.Z*scaleY - p.Y*scaleY*perspectiveY)

	// Depth for z-ordering: elements with higher Y are "further back"
	depth := p.Y + p.Z*depthZFactor

	return screenX, screenY, depth
}

// RenderWave renders the flowing ribbon wave to the buffer using shaded characters.
func (r *Renderer) RenderWave(w *wave.Wave) {
	numLayers, numPoints := w.Size()
	minZ, maxZ := w.MinZ, w.MaxZ
	zRange := maxZ - minZ
	if zRange == 0 {
		zRange = 1
	}

	// Draw from back to front (lower layer index = back)
	for layer := 0; layer < numLayers; layer++ {
		layerFactor := float64(layer) / float64(numLayers-1)

		for i := 0; i < numPoints; i++ {
			r.renderWaveSegment(w, layer, i, layerFactor, minZ, zRange)
		}
	}
}

func (r *Renderer) renderWaveSegment(w *wave.Wave, layer, i int, layerFactor, minZ, zRange float64) {
	numLayers, numPoints := w.Size()
	p1 := w.Points[layer][i]
	x1, y1, d1 := r.project3D(p1)

	// Normalized height for shading (0 = valley, 1 = peak)
	normalizedZ := (p1.Z - minZ) / zRange

	// Get character and style based on depth and height
	char := r.getShadeChar(normalizedZ, layerFactor)
	style := r.getStyle(normalizedZ, layerFactor)

	// Draw horizontal line to next point (along the wave)
	if i < numPoints-1 {
		p2 := w.Points[layer][i+1]
		x2, y2, d2 := r.project3D(p2)
		avgDepth := (d1 + d2) / 2
		avgZ := ((p1.Z - minZ) / zRange + (p2.Z - minZ) / zRange) / 2
		r.drawShadedLine(x1, y1, x2, y2, avgDepth, avgZ, layerFactor, style)
	}

	// Draw vertical line to next layer (creates ribbon depth)
	if layer < numLayers-1 {
		p3 := w.Points[layer+1][i]
		x3, y3, d3 := r.project3D(p3)
		avgDepth := (d1 + d3) / 2
		// Vertical lines use block characters for filled look
		r.drawVerticalFill(x1, y1, x3, y3, avgDepth, normalizedZ, layerFactor, style)
	}

	// Draw the point itself
	r.setCell(x1, y1, char, d1, style)
}

// getShadeChar returns an ASCII character based on depth and height for 3D effect.
func (r *Renderer) getShadeChar(normalizedZ float64, layerFactor float64) rune {
	// Combine height and layer for shading
	// Front layers (high layerFactor) and peaks (high normalizedZ) are brighter
	shade := normalizedZ*0.7 + layerFactor*0.3
	return mapToChar(shade, shadeChars)
}

// getBlockChar returns a block character for filled vertical sections.
func (r *Renderer) getBlockChar(normalizedZ float64, layerFactor float64) rune {
	shade := normalizedZ*0.6 + layerFactor*0.4
	return mapToChar(shade, blockChars)
}

// mapToChar maps a normalized value (0-1) to a character from the set.
func mapToChar(value float64, chars []rune) rune {
	idx := int(value * float64(len(chars)-1))
	if idx < 0 {
		idx = 0
	}
	if idx >= len(chars) {
		idx = len(chars) - 1
	}
	return chars[idx]
}

type colorStop struct {
	threshold float64
	r, g, b   int32
}

var colorGradient = []colorStop{
	{0.15, 30, 50, 120},   // Deep blue
	{0.30, 50, 80, 160},   // Medium blue
	{0.45, 70, 120, 200},  // Blue
	{0.60, 100, 160, 220}, // Light blue
	{0.75, 140, 200, 235}, // Cyan
	{0.90, 180, 225, 245}, // Light cyan
	{2.00, 220, 245, 255}, // Near white (catch-all)
}

// getStyle returns a color style based on normalized height and layer position.
func (r *Renderer) getStyle(normalizedZ float64, layerFactor float64) tcell.Style {
	// Create a blue-cyan-white gradient for 3D depth
	t := normalizedZ*0.6 + layerFactor*0.4

	for _, stop := range colorGradient {
		if t < stop.threshold {
			return tcell.StyleDefault.Foreground(tcell.NewRGBColor(stop.r, stop.g, stop.b))
		}
	}
	// Fallback (should be covered by the last stop)
	last := colorGradient[len(colorGradient)-1]
	return tcell.StyleDefault.Foreground(tcell.NewRGBColor(last.r, last.g, last.b))
}

// drawShadedLine draws a line with varying shade based on position.
func (r *Renderer) drawShadedLine(x1, y1, x2, y2 int, depth, normalizedZ, layerFactor float64, style tcell.Style) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	if x1 > x2 {
		sx = -1
	}
	sy := 1
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy

	// Choose character based on line direction and shading
	var lineChar rune
	if dy > dx {
		// More vertical - use vertical-ish characters
		lineChar = '|'
	} else {
		// More horizontal - use shade character
		lineChar = r.getShadeChar(normalizedZ, layerFactor)
	}

	steps := 0
	maxSteps := dx + dy + 1

	for steps < maxSteps {
		r.setCell(x1, y1, lineChar, depth, style)

		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
		steps++
	}
}

// drawVerticalFill draws vertical connections between layers with block characters.
func (r *Renderer) drawVerticalFill(x1, y1, x2, y2 int, depth, normalizedZ, layerFactor float64, style tcell.Style) {
	// Only draw if there's vertical distance
	if y1 == y2 {
		return
	}

	dy := y2 - y1
	sy := 1
	if dy < 0 {
		sy = -1
		dy = -dy
	}

	// Use block characters for vertical fill
	char := r.getBlockChar(normalizedZ, layerFactor)

	y := y1
	for i := 0; i <= dy; i++ {
		r.setCell(x1, y, char, depth, style)
		y += sy
	}
}

// setCell sets a character at the given position with depth testing for proper z-ordering.
func (r *Renderer) setCell(x, y int, char rune, depth float64, style tcell.Style) {
	if x < 0 || x >= r.width || y < 0 || y >= r.height {
		return
	}
	// Depth test - draw if in front of existing content
	if depth > r.buffer[y][x].depth {
		r.buffer[y][x] = cell{
			char:  char,
			style: style,
			depth: depth,
			set:   true,
		}
	}
}

// Flush renders the internal buffer to the actual screen and displays it.
func (r *Renderer) Flush() {
	for y := 0; y < r.height; y++ {
		for x := 0; x < r.width; x++ {
			c := r.buffer[y][x]
			if c.set {
				r.screen.SetContent(x, y, c.char, nil, c.style)
			}
		}
	}
	r.screen.Show()
}

// abs returns the absolute value of an integer.
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
