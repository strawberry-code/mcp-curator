package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/strawberry-code/mcp-curator/internal/application"
	"github.com/strawberry-code/mcp-curator/internal/i18n"
	"github.com/strawberry-code/mcp-curator/internal/version"
)

// MainWindow è la finestra principale dell'applicazione
type MainWindow struct {
	app         fyne.App
	window      fyne.Window
	service     *application.MCPService
	tree        *widget.Tree
	detailPanel *fyne.Container
	selectedID  string
	mainContent fyne.CanvasObject

	// Elementi UI che richiedono aggiornamento su cambio lingua
	addBtn     *widget.Button
	refreshBtn *widget.Button
	langSelect *widget.Select
}

// NewMainWindow crea la finestra principale
func NewMainWindow(app fyne.App, service *application.MCPService) *MainWindow {
	window := app.NewWindow(version.Name + " v" + version.Version)
	window.Resize(fyne.NewSize(900, 800))

	mw := &MainWindow{
		app:     app,
		window:  window,
		service: service,
	}

	return mw
}

// Show mostra la finestra con splash screen iniziale
func (mw *MainWindow) Show() {
	mw.window.Show()

	// Mostra splash view, poi passa al contenuto principale
	splash := NewSplashView(mw.window, func() {
		mw.buildUI()
		mw.window.SetContent(mw.mainContent)
	})

	mw.window.SetContent(splash.Content())
	splash.StartAnimation()
}

// buildUI costruisce l'interfaccia utente
func (mw *MainWindow) buildUI() {
	// Tree view per scope e server
	mw.tree = mw.createTree()

	// Pannello dettagli (inizialmente vuoto)
	mw.detailPanel = container.NewVBox(
		widget.NewLabel(i18n.T("detail.select_server")),
	)

	// Toolbar
	toolbar := mw.createToolbar()

	// Split view
	split := container.NewHSplit(
		container.NewScroll(mw.tree),
		container.NewScroll(mw.detailPanel),
	)
	split.SetOffset(0.35)

	// Layout principale
	mw.mainContent = container.NewBorder(
		toolbar,
		nil,
		nil,
		nil,
		split,
	)

	// Registra callback per cambio lingua
	i18n.OnChange(func(lang i18n.Lang) {
		mw.updateUIStrings()
	})
}

// createToolbar crea la toolbar con i bottoni principali
func (mw *MainWindow) createToolbar() *fyne.Container {
	mw.addBtn = widget.NewButtonWithIcon(i18n.T("toolbar.add_server"), theme.ContentAddIcon(), func() {
		mw.showAddServerDialog()
	})

	mw.refreshBtn = widget.NewButtonWithIcon(i18n.T("toolbar.refresh"), theme.ViewRefreshIcon(), func() {
		mw.refresh()
	})

	// Selettore lingua compatto
	langs := []string{"IT", "EN", "FR", "DE", "ES", "PT", "JA", "KO", "CN", "UK"}
	mw.langSelect = widget.NewSelect(langs, func(selected string) {
		mw.changeLanguage(selected)
	})
	mw.langSelect.SetSelected(string(i18n.CurrentLang()))

	return container.NewHBox(
		mw.addBtn,
		mw.refreshBtn,
		widget.NewSeparator(),
		mw.langSelect,
	)
}

// changeLanguage cambia la lingua dell'interfaccia
func (mw *MainWindow) changeLanguage(lang string) {
	i18n.SetLang(lang)
}

// updateUIStrings aggiorna tutte le stringhe dell'UI dopo cambio lingua
func (mw *MainWindow) updateUIStrings() {
	// Aggiorna bottoni toolbar
	mw.addBtn.SetText(i18n.T("toolbar.add_server"))
	mw.refreshBtn.SetText(i18n.T("toolbar.refresh"))

	// Aggiorna tree
	mw.tree.Refresh()

	// Aggiorna pannello dettagli
	if mw.selectedID != "" {
		mw.updateDetailPanel(mw.selectedID)
	} else {
		mw.detailPanel.RemoveAll()
		mw.detailPanel.Add(widget.NewLabel(i18n.T("detail.select_server")))
	}
}

// refresh ricarica la configurazione e aggiorna l'UI
func (mw *MainWindow) refresh() {
	if err := mw.service.Load(); err != nil {
		dialog.ShowError(err, mw.window)
		return
	}

	// Aggiorna solo il tree senza ricrearlo
	mw.tree.Refresh()

	// Aggiorna il pannello dettagli se c'è una selezione
	if mw.selectedID != "" {
		mw.updateDetailPanel(mw.selectedID)
	}
}
