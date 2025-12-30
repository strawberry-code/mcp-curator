package ui

import (
	"image/color"
	"math"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"github.com/strawberry-code/mcp-curator/internal/version"
)

// Colori splash (derivati dal tema)
var (
	splashBg      = ColorAnthracite
	splashFg      = ColorWhite
	splashGrayVer = ColorGrayText
)

// SplashView gestisce la vista splash animata nella stessa finestra
type SplashView struct {
	window    fyne.Window
	onFinish  func()
	container *fyne.Container
	// Elementi grafici
	bg      *canvas.Rectangle
	ring    *canvas.Circle    // Cerchio con bordo bianco (contorno)
	fill    *canvas.Circle    // Cerchio bianco pieno
	maskTop *canvas.Rectangle // Maschera nera che copre la parte superiore del fill
	title   *canvas.Text
	ver     *canvas.Text
}

// NewSplashView crea una nuova vista splash per la finestra esistente
func NewSplashView(window fyne.Window, onFinish func()) *SplashView {
	sv := &SplashView{
		window:   window,
		onFinish: onFinish,
	}

	sv.buildContent()
	return sv
}

// buildContent costruisce il contenuto della splash view
func (sv *SplashView) buildContent() {
	// Background antracite
	sv.bg = canvas.NewRectangle(splashBg)

	// Cerchio bianco pieno (il fill interno)
	sv.fill = canvas.NewCircle(splashFg)

	// Maschera antracite che copre la parte superiore del fill (simula il livello del liquido)
	sv.maskTop = canvas.NewRectangle(splashBg)

	// Cerchio con bordo bianco (il contorno visibile) - sopra tutto
	sv.ring = canvas.NewCircle(color.Transparent)
	sv.ring.StrokeColor = splashFg
	sv.ring.StrokeWidth = 2

	// Titolo
	sv.title = canvas.NewText(version.Name, splashFg)
	sv.title.TextSize = 24
	sv.title.TextStyle = fyne.TextStyle{Monospace: true}

	// Versione
	sv.ver = canvas.NewText("v"+version.Version, splashGrayVer)
	sv.ver.TextSize = 11
	sv.ver.TextStyle = fyne.TextStyle{Monospace: true}

	// Ordine: bg, fill, maskTop (copre parte superiore), ring (bordo sopra), testi
	sv.container = container.NewWithoutLayout(
		sv.bg,
		sv.fill,
		sv.maskTop,
		sv.ring,
		sv.title,
		sv.ver,
	)

	sv.updateLayout(0)
}

// updateLayout aggiorna le posizioni degli elementi
func (sv *SplashView) updateLayout(fillPercent float64) {
	size := sv.window.Canvas().Size()
	centerX := size.Width / 2
	centerY := size.Height/2 - 30

	// Background
	sv.bg.Resize(size)

	// Dimensioni cerchio
	ringSize := float32(70)
	ringX := centerX - ringSize/2
	ringY := centerY - ringSize/2

	// Fill: cerchio bianco pieno delle stesse dimensioni del ring
	sv.fill.Resize(fyne.NewSize(ringSize, ringSize))
	sv.fill.Move(fyne.NewPos(ringX, ringY))

	// Maschera nera che copre la parte superiore del cerchio
	// Quando fillPercent = 0, copre tutto il cerchio (altezza = ringSize)
	// Quando fillPercent = 1, non copre nulla (altezza = 0)
	maskHeight := ringSize * float32(1-fillPercent)
	sv.maskTop.Resize(fyne.NewSize(ringSize, maskHeight))
	sv.maskTop.Move(fyne.NewPos(ringX, ringY))

	// Anello (bordo) sopra tutto
	sv.ring.Resize(fyne.NewSize(ringSize, ringSize))
	sv.ring.Move(fyne.NewPos(ringX, ringY))

	// Titolo centrato sotto l'anello
	sv.title.Move(fyne.NewPos(centerX-70, centerY+55))

	// Versione
	sv.ver.Move(fyne.NewPos(centerX-30, centerY+85))
}

// Content restituisce il container della splash view
func (sv *SplashView) Content() *fyne.Container {
	return sv.container
}

// StartAnimation avvia l'animazione e chiama onFinish al termine
func (sv *SplashView) StartAnimation() {
	go sv.animate()
}

// easeInOutCubic funzione di easing per movimento fluido
func easeInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}

// animate esegue l'animazione della splash view
func (sv *SplashView) animate() {
	startTime := time.Now()
	duration := 2500 * time.Millisecond
	ticker := time.NewTicker(25 * time.Millisecond) // 40 FPS
	defer ticker.Stop()

	for range ticker.C {
		elapsed := time.Since(startTime)
		if elapsed >= duration {
			break
		}

		// Progresso normalizzato 0-1
		progress := float64(elapsed) / float64(duration)

		// Fase 1 (0-0.15): Fade in elementi
		// Fase 2 (0.15-0.85): Riempimento cerchio
		// Fase 3 (0.85-1.0): Fade out

		var fillPercent float64
		var alpha uint8 = 255

		if progress < 0.15 {
			// Fade in
			alpha = uint8(255 * easeInOutCubic(progress/0.15))
			fillPercent = 0
		} else if progress < 0.85 {
			// Riempimento graduale
			fillProgress := (progress - 0.15) / 0.70
			fillPercent = easeInOutCubic(fillProgress)
		} else {
			// Pieno + fade out
			fillPercent = 1.0
			fadeProgress := (progress - 0.85) / 0.15
			alpha = uint8(255 * (1 - easeInOutCubic(fadeProgress)))
		}

		// Copia valori per la closure
		currentAlpha := alpha
		currentFillPercent := fillPercent

		// Aggiorna UI nel thread principale
		fyne.Do(func() {
			sv.ring.StrokeColor = color.RGBA{R: splashFg.R, G: splashFg.G, B: splashFg.B, A: currentAlpha}
			sv.fill.FillColor = color.RGBA{R: splashFg.R, G: splashFg.G, B: splashFg.B, A: currentAlpha}
			sv.title.Color = color.RGBA{R: splashFg.R, G: splashFg.G, B: splashFg.B, A: currentAlpha}
			sv.ver.Color = color.RGBA{R: splashGrayVer.R, G: splashGrayVer.G, B: splashGrayVer.B, A: currentAlpha}

			sv.updateLayout(currentFillPercent)
			sv.container.Refresh()
		})
	}

	// Callback per passare alla vista principale (nel thread UI)
	if sv.onFinish != nil {
		fyne.Do(sv.onFinish)
	}
}
