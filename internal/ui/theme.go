package ui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// Colori base dell'applicazione
var (
	ColorWhite     = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	ColorAnthracite = color.RGBA{R: 45, G: 45, B: 48, A: 255} // Grigio antracite scuro
	ColorGrayLight = color.RGBA{R: 80, G: 80, B: 85, A: 255}  // Per bordi e separatori
	ColorGrayText  = color.RGBA{R: 180, G: 180, B: 180, A: 255} // Per testo secondario
)

// CuratorTheme è il tema personalizzato dell'applicazione
type CuratorTheme struct{}

var _ fyne.Theme = (*CuratorTheme)(nil)

// Color restituisce i colori del tema
func (t *CuratorTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return ColorAnthracite
	case theme.ColorNameForeground:
		return ColorWhite
	case theme.ColorNameButton:
		return ColorGrayLight
	case theme.ColorNameDisabledButton:
		return color.RGBA{R: 60, G: 60, B: 65, A: 255}
	case theme.ColorNameDisabled:
		return color.RGBA{R: 100, G: 100, B: 105, A: 255}
	case theme.ColorNamePlaceHolder:
		return ColorGrayText
	case theme.ColorNamePrimary:
		return ColorWhite
	case theme.ColorNameHover:
		return color.RGBA{R: 70, G: 70, B: 75, A: 255}
	case theme.ColorNameFocus:
		return ColorWhite
	case theme.ColorNameSelection:
		return color.RGBA{R: 80, G: 80, B: 90, A: 255}
	case theme.ColorNameInputBackground:
		return color.RGBA{R: 35, G: 35, B: 38, A: 255}
	case theme.ColorNameInputBorder:
		return ColorGrayLight
	case theme.ColorNameScrollBar:
		return ColorGrayLight
	case theme.ColorNameShadow:
		return color.RGBA{R: 0, G: 0, B: 0, A: 100}
	case theme.ColorNameSeparator:
		return ColorGrayLight
	case theme.ColorNameHeaderBackground:
		return color.RGBA{R: 55, G: 55, B: 60, A: 255}
	case theme.ColorNameMenuBackground:
		return ColorAnthracite
	case theme.ColorNameOverlayBackground:
		return color.RGBA{R: 40, G: 40, B: 45, A: 230}
	case theme.ColorNameSuccess:
		return color.RGBA{R: 120, G: 200, B: 120, A: 255}
	case theme.ColorNameWarning:
		return color.RGBA{R: 220, G: 180, B: 80, A: 255}
	case theme.ColorNameError:
		return color.RGBA{R: 220, G: 100, B: 100, A: 255}
	}
	return theme.DefaultTheme().Color(name, variant)
}

// Font restituisce il font del tema
func (t *CuratorTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

// Icon restituisce le icone del tema
func (t *CuratorTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

// Size restituisce le dimensioni del tema
func (t *CuratorTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}
