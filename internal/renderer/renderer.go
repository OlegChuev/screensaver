// Package renderer provides 3D to 2D projection and terminal rendering for the wave screensaver.
package renderer

import (
	"math"

	"github.com/gdamore/tcell/v2"
	"github.com/olegchuev/screensaver/internal/wave"
)

// ASCII characters for 3D shading effect - from darkest/furthest to brightest/closest
var shadeChars = []rune{'·', ':', '÷', '≈', '≠', '≡', '∫', '#', '▓', '█'}

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

// RenderWave renders the particle-based ocean surface to the buffer.
func (r *Renderer) RenderWave(w *wave.Wave) {
	gridDepth, gridWidth := w.Size()
	minZ, maxZ := w.MinZ, w.MaxZ
	zRange := maxZ - minZ
	if zRange == 0 {
		zRange = 1
	}

	// Render surface grid
	for depth := 0; depth < gridDepth-1; depth++ {
		for width := 0; width < gridWidth-1; width++ {
			// Get four corners of the grid cell
			p1 := w.GridPoints[depth][width]
			p2 := w.GridPoints[depth][width+1]
			p3 := w.GridPoints[depth+1][width]
			p4 := w.GridPoints[depth+1][width+1]

			// Project to screen space
			x1, y1, d1 := r.project3D(p1)
			x2, y2, d2 := r.project3D(p2)
			x3, y3, d3 := r.project3D(p3)
			x4, y4, d4 := r.project3D(p4)

			// Calculate average properties for the quad
			avgZ := (p1.Z + p2.Z + p3.Z + p4.Z) / 4.0
			normalizedZ := (avgZ - minZ) / zRange
			avgDepth := (d1 + d2 + d3 + d4) / 4.0
			depthFactor := float64(depth) / float64(gridDepth-1)

			// Get style based on wave height
			style := r.getStyle(normalizedZ, depthFactor)
			char := r.getShadeChar(normalizedZ, depthFactor)

			// Draw the grid cell edges
			r.drawShadedLine(x1, y1, x2, y2, (d1+d2)/2, normalizedZ, depthFactor, style)
			r.drawShadedLine(x1, y1, x3, y3, (d1+d3)/2, normalizedZ, depthFactor, style)

			// Fill the quad center with a character
			centerX := (x1 + x2 + x3 + x4) / 4
			centerY := (y1 + y2 + y3 + y4) / 4
			r.setCell(centerX, centerY, char, avgDepth, style)
		}
	}

	// Render particles (spray/foam effect)
	for _, particle := range w.Particles {
		px, py, pd := r.project3D(particle.Pos)
		// Particles use brighter colors and special characters
		particleStyle := tcell.StyleDefault.Foreground(tcell.NewRGBColor(255, 255, 255))
		r.setCell(px, py, '•', pd, particleStyle)
	}
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
	{0.15, 30, 30, 30},    // Dark Grey
	{0.30, 80, 80, 80},    // Dim Grey
	{0.45, 120, 120, 120}, // Grey
	{0.60, 160, 160, 160}, // Silver
	{0.75, 200, 200, 200}, // Light Grey
	{0.90, 230, 230, 230}, // Gainsboro
	{2.00, 255, 255, 255}, // White
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
