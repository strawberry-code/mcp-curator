package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/strawberry-code/mcp-curator/internal/application"
)

const (
	appID = "com.strawberry-code.mcp-curator"
)

// App rappresenta l'applicazione principale
type App struct {
	fyneApp    fyne.App
	mainWindow *MainWindow
	service    *application.MCPService
}

// NewApp crea una nuova applicazione
func NewApp() (*App, error) {
	service, err := application.NewMCPService()
	if err != nil {
		return nil, err
	}

	if err := service.Load(); err != nil {
		return nil, err
	}

	fyneApp := app.NewWithID(appID)
	fyneApp.Settings().SetTheme(&CuratorTheme{})

	a := &App{
		fyneApp: fyneApp,
		service: service,
	}

	return a, nil
}

// Run avvia l'applicazione
func (a *App) Run() {
	a.mainWindow = NewMainWindow(a.fyneApp, a.service)
	a.mainWindow.Show()
	a.fyneApp.Run()
}
