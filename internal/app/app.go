// Package app provides the main application logic for the screensaver.
package app

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/olegchuev/screensaver/internal/renderer"
	"github.com/olegchuev/screensaver/internal/wave"
)

// Config holds application configuration including timing and wave parameters.
type Config struct {
	FrameDelay time.Duration
	WaveConfig wave.Config
}

// DefaultConfig returns default application configuration with sensible defaults.
func DefaultConfig() Config {
	return Config{
		FrameDelay: 50 * time.Millisecond, // Smooth animation at ~20 FPS
		WaveConfig: wave.DefaultConfig(),
	}
}

// App represents the screensaver application with all its components.
type App struct {
	config   Config
	screen   tcell.Screen
	renderer *renderer.Renderer
	wave     *wave.Wave
	running  bool
}

// New creates and initializes a new screensaver application instance.
func New(cfg Config) (*App, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := screen.Init(); err != nil {
		return nil, err
	}

	screen.SetStyle(tcell.StyleDefault.Background(tcell.ColorBlack))
	screen.HideCursor()
	screen.Clear()

	return &App{
		config:   cfg,
		screen:   screen,
		renderer: renderer.NewRenderer(screen),
		wave:     wave.NewWave(cfg.WaveConfig),
		running:  true,
	}, nil
}

// Run starts the main loop of the screensaver, handling events and rendering frames.
func (a *App) Run() error {
	defer a.screen.Fini()

	// Signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(a.config.FrameDelay)
	defer ticker.Stop()

	t := 0.0

	for a.running {
		select {
		case <-sigChan:
			return nil
		case <-ticker.C:
			// Handle pending input events
			if a.screen.HasPendingEvent() {
				ev := a.screen.PollEvent()
				if a.handleEvent(ev) {
					return nil
				}
			}

			// Update wave state and render frame
			a.update(t)
			a.render()

			t += 0.08 // Time progression for wave animation
		}
	}

	return nil
}

// handleEvent processes input events and returns true if the app should quit.
func (a *App) handleEvent(ev tcell.Event) bool {
	switch ev := ev.(type) {
	case *tcell.EventKey:
		switch ev.Key() {
		case tcell.KeyEscape, tcell.KeyCtrlC:
			return true
		case tcell.KeyRune:
			if ev.Rune() == 'q' || ev.Rune() == 'Q' {
				return true
			}
		}
	case *tcell.EventResize:
		a.screen.Sync()
		a.renderer.Resize()
	}
	return false
}

// update advances the wave simulation by the given time delta.
func (a *App) update(t float64) {
	a.wave.Update(t)
}

// render clears the screen and draws the current wave state.
func (a *App) render() {
	a.renderer.Clear()
	a.renderer.RenderWave(a.wave)
	a.renderer.Flush()
}

// Stop signals the application to stop running.
func (a *App) Stop() {
	a.running = false
}
