// Package wave provides wave simulation and generation for the screensaver.
package wave

import (
	"math"
)

// Config holds wave simulation parameters for controlling the wave appearance and behavior.
type Config struct {
	// Grid dimensions
	GridWidth int
	GridDepth int
	// Particle density
	ParticleDensity float64
	// Wave parameters using Gerstner wave equations
	WaveCount int
}

// DefaultConfig returns sensible defaults for a particle-based ocean wave.
func DefaultConfig() Config {
	return Config{
		GridWidth:       80,
		GridDepth:       60,
		ParticleDensity: 0.3,
		WaveCount:       3,
	}
}

// WaveParams represents parameters for a single Gerstner wave component.
type WaveParams struct {
	Amplitude  float64
	Wavelength float64
	Speed      float64
	Direction  [2]float64 // Normalized direction vector
	Steepness  float64    // 0-1, controls wave sharpness
}

// Point3D represents a point in 3D space with X, Y, Z coordinates.
type Point3D struct {
	X, Y, Z float64
}

// Particle represents a water particle with position and properties.
type Particle struct {
	Pos      Point3D
	Velocity float64
}

// Wave represents a particle-based ocean surface using Gerstner waves.
type Wave struct {
	config     Config
	Particles  []Particle
	GridPoints [][]Point3D // Surface grid for rendering
	waves      []WaveParams
	MinZ       float64
	MaxZ       float64
}

// NewWave creates a new particle-based ocean wave with the given configuration.
func NewWave(cfg Config) *Wave {
	w := &Wave{
		config:     cfg,
		Particles:  make([]Particle, 0),
		GridPoints: make([][]Point3D, cfg.GridDepth),
		waves:      make([]WaveParams, cfg.WaveCount),
	}

	// Initialize grid
	for i := range w.GridPoints {
		w.GridPoints[i] = make([]Point3D, cfg.GridWidth)
	}

	// Initialize Gerstner wave components with different characteristics
	w.waves[0] = WaveParams{
		Amplitude:  0.15,
		Wavelength: 1.5,
		Speed:      0.8,
		Direction:  [2]float64{1.0, 0.3},
		Steepness:  0.6,
	}
	w.waves[1] = WaveParams{
		Amplitude:  0.08,
		Wavelength: 0.8,
		Speed:      1.2,
		Direction:  [2]float64{0.7, -0.5},
		Steepness:  0.4,
	}
	w.waves[2] = WaveParams{
		Amplitude:  0.05,
		Wavelength: 0.4,
		Speed:      1.6,
		Direction:  [2]float64{-0.3, 0.8},
		Steepness:  0.3,
	}

	// Normalize wave directions
	for i := range w.waves {
		len := math.Sqrt(w.waves[i].Direction[0]*w.waves[i].Direction[0] +
			w.waves[i].Direction[1]*w.waves[i].Direction[1])
		w.waves[i].Direction[0] /= len
		w.waves[i].Direction[1] /= len
	}

	return w
}

// Update recalculates the ocean surface using Gerstner wave equations.
func (w *Wave) Update(t float64) {
	cfg := w.config
	w.MinZ = math.MaxFloat64
	w.MaxZ = -math.MaxFloat64

	// Update surface grid using Gerstner waves
	for depth := 0; depth < cfg.GridDepth; depth++ {
		for width := 0; width < cfg.GridWidth; width++ {
			// Original position on the grid
			x0 := (float64(width)/float64(cfg.GridWidth-1))*2.0 - 1.0 // -1 to 1
			y0 := (float64(depth)/float64(cfg.GridDepth-1))*2.0 - 1.0 // -1 to 1

			// Apply Gerstner wave displacement
			x, y, z := w.gerstnerWave(x0, y0, t)

			w.GridPoints[depth][width] = Point3D{X: x, Y: y, Z: z}

			if z < w.MinZ {
				w.MinZ = z
			}
			if z > w.MaxZ {
				w.MaxZ = z
			}
		}
	}

	// Generate particles on the surface (spray/foam effect)
	w.Particles = w.Particles[:0] // Clear existing particles
	for depth := 0; depth < cfg.GridDepth; depth += 3 {
		for width := 0; width < cfg.GridWidth; width += 3 {
			if math.Mod(float64(width+depth), 1.0/cfg.ParticleDensity) < 1.0 {
				p := w.GridPoints[depth][width]
				// Add particles at wave peaks
				if p.Z > (w.MaxZ-w.MinZ)*0.6+w.MinZ {
					w.Particles = append(w.Particles, Particle{
						Pos:      p,
						Velocity: p.Z,
					})
				}
			}
		}
	}
}

// gerstnerWave calculates the position of a point using Gerstner wave equations.
func (w *Wave) gerstnerWave(x0, y0, t float64) (float64, float64, float64) {
	x, y, z := x0, y0, 0.0

	for _, wave := range w.waves {
		// Wave parameters
		k := 2.0 * math.Pi / wave.Wavelength // Wave number
		c := wave.Speed                      // Phase speed
		Q := wave.Steepness / (k * wave.Amplitude * float64(len(w.waves)))

		// Direction components
		dx, dy := wave.Direction[0], wave.Direction[1]

		// Phase
		phase := k*(dx*x0+dy*y0) - c*t

		// Gerstner wave displacement
		x += Q * wave.Amplitude * dx * math.Cos(phase)
		y += Q * wave.Amplitude * dy * math.Cos(phase)
		z += wave.Amplitude * math.Sin(phase)
	}

	return x, y, z
}

// Size returns the dimensions of the wave grid (depth, width).
func (w *Wave) Size() (int, int) {
	return w.config.GridDepth, w.config.GridWidth
}
