// Package wave provides wave simulation and generation for the screensaver.
package wave

import (
	"math"
)

const (
	secondaryWaveSpeed = 0.7
	tertiaryWaveAmp    = 0.3
	tertiaryWaveFreq   = 1.5
	tertiaryWaveSpeed  = 1.3
	ribbonDepthScale   = 0.3
	heightVariation    = 0.2
)

// Config holds wave simulation parameters for controlling the wave appearance and behavior.
type Config struct {
	// Number of horizontal samples across the screen
	NumPoints int
	// Number of vertical layers in the ribbon
	NumLayers int
	// Wave amplitude (height of peaks)
	Amplitude float64
	// Wave frequency (number of waves visible)
	Frequency float64
	// Secondary wave parameters for complexity
	Amplitude2 float64
	Frequency2 float64
}

// DefaultConfig returns sensible defaults for a flowing ribbon wave.
func DefaultConfig() Config {
	return Config{
		NumPoints:  120,
		NumLayers:  20,
		Amplitude:  0.15,
		Frequency:  2.0,
		Amplitude2: 0.08,
		Frequency2: 3.5,
	}
}

// Point3D represents a point in 3D space with X, Y, Z coordinates.
type Point3D struct {
	X, Y, Z float64
}

// Wave represents a flowing ribbon wave that moves horizontally across the screen.
type Wave struct {
	config Config
	Points [][]Point3D // [layer][point along wave]
	MinZ   float64
	MaxZ   float64
}

// NewWave creates a new flowing ribbon wave with the given configuration.
func NewWave(cfg Config) *Wave {
	w := &Wave{
		config: cfg,
		Points: make([][]Point3D, cfg.NumLayers),
	}
	for i := range w.Points {
		w.Points[i] = make([]Point3D, cfg.NumPoints)
	}
	return w
}

// Update recalculates all wave points for the given time t, animating the wave motion.
func (w *Wave) Update(t float64) {
	cfg := w.config
	w.MinZ = math.MaxFloat64
	w.MaxZ = -math.MaxFloat64

	for layer := 0; layer < cfg.NumLayers; layer++ {
		// Layer offset creates the ribbon depth
		layerOffset := float64(layer) / float64(cfg.NumLayers-1)

		for i := 0; i < cfg.NumPoints; i++ {
			// X position: spans from left (-0.5) to right (0.5)
			x := float64(i)/float64(cfg.NumPoints-1) - 0.5

			// Calculate wave height using multiple sine waves for organic look
			// Primary wave moves with time
			z := cfg.Amplitude * math.Sin(x*2*math.Pi*cfg.Frequency+t)
			// Secondary wave adds complexity
			z += cfg.Amplitude2 * math.Sin(x*2*math.Pi*cfg.Frequency2-t*secondaryWaveSpeed)
			// Third harmonic for more detail
			z += cfg.Amplitude * tertiaryWaveAmp * math.Sin(x*2*math.Pi*cfg.Frequency*tertiaryWaveFreq+t*tertiaryWaveSpeed)

			// Y position: creates the ribbon depth/thickness
			// Each layer is slightly offset in Y
			y := (layerOffset - 0.5) * ribbonDepthScale

			// Add slight Y variation based on wave height for 3D effect
			y += z * heightVariation

			if z < w.MinZ {
				w.MinZ = z
			}
			if z > w.MaxZ {
				w.MaxZ = z
			}

			w.Points[layer][i] = Point3D{X: x, Y: y, Z: z}
		}
	}
}

// Size returns the dimensions of the wave grid (layers, points per layer).
func (w *Wave) Size() (int, int) {
	return w.config.NumLayers, w.config.NumPoints
}
