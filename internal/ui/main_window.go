package ui

import (
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/strawberry-code/mcp-curator/internal/application"
	"github.com/strawberry-code/mcp-curator/internal/domain"
	"github.com/strawberry-code/mcp-curator/internal/version"
)

// MainWindow è la finestra principale dell'applicazione
type MainWindow struct {
	window      fyne.Window
	service     *application.MCPService
	tree        *widget.Tree
	detailPanel *fyne.Container
	selectedID  string
	mainContent fyne.CanvasObject
}

// NewMainWindow crea la finestra principale
func NewMainWindow(app fyne.App, service *application.MCPService) *MainWindow {
	window := app.NewWindow(version.Name + " v" + version.Version)
	window.Resize(fyne.NewSize(900, 600))

	mw := &MainWindow{
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
		widget.NewLabel("Seleziona un server per vedere i dettagli"),
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
}

// createToolbar crea la toolbar
func (mw *MainWindow) createToolbar() *fyne.Container {
	addBtn := widget.NewButtonWithIcon("Aggiungi Server", theme.ContentAddIcon(), func() {
		mw.showAddServerDialog()
	})

	refreshBtn := widget.NewButtonWithIcon("Aggiorna", theme.ViewRefreshIcon(), func() {
		mw.refresh()
	})

	return container.NewHBox(
		addBtn,
		layout.NewSpacer(),
		refreshBtn,
	)
}

// createTree crea il tree view
func (mw *MainWindow) createTree() *widget.Tree {
	tree := widget.NewTree(
		// childUIDs - restituisce i figli di un nodo in ordine alfabetico
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			config := mw.service.GetConfiguration()

			if id == "" {
				return []string{"global", "projects"}
			}
			if id == "global" {
				ids := make([]string, 0, len(config.GlobalServers))
				for name := range config.GlobalServers {
					ids = append(ids, "global:"+name)
				}
				sort.Strings(ids)
				return ids
			}
			if id == "projects" {
				ids := make([]string, 0, len(config.Projects))
				for path := range config.Projects {
					ids = append(ids, "project:"+path)
				}
				sort.Strings(ids)
				return ids
			}
			// Server di un progetto
			if len(id) > 8 && id[:8] == "project:" {
				projectPath := id[8:]
				if project, ok := config.Projects[projectPath]; ok {
					ids := make([]string, 0, len(project.MCPServers))
					for name := range project.MCPServers {
						ids = append(ids, "projectserver:"+projectPath+":"+name)
					}
					sort.Strings(ids)
					return ids
				}
			}
			return nil
		},
		// isBranch
		func(id widget.TreeNodeID) bool {
			return id == "" || id == "global" || id == "projects" ||
				(len(id) > 8 && id[:8] == "project:")
		},
		// create
		func(branch bool) fyne.CanvasObject {
			return widget.NewLabel("")
		},
		// update
		func(id widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			config := mw.service.GetConfiguration()
			label := o.(*widget.Label)

			switch {
			case id == "global":
				label.SetText("Globale")
			case id == "projects":
				label.SetText("Progetti")
			case len(id) > 7 && id[:7] == "global:":
				label.SetText(id[7:])
			case len(id) > 8 && id[:8] == "project:":
				path := id[8:]
				if project, ok := config.Projects[path]; ok {
					label.SetText(project.Name)
				} else {
					label.SetText(path)
				}
			case len(id) > 14 && id[:14] == "projectserver:":
				// projectserver:path:name
				rest := id[14:]
				for i := len(rest) - 1; i >= 0; i-- {
					if rest[i] == ':' {
						label.SetText(rest[i+1:])
						break
					}
				}
			default:
				label.SetText(id)
			}
		},
	)

	tree.OnSelected = func(id widget.TreeNodeID) {
		mw.selectedID = id
		mw.updateDetailPanel(id)
	}

	return tree
}

// updateDetailPanel aggiorna il pannello dettagli
func (mw *MainWindow) updateDetailPanel(id string) {
	config := mw.service.GetConfiguration()

	var server *domain.MCPServer
	var serverName string
	var scope string
	var projectPath string

	switch {
	case len(id) > 7 && id[:7] == "global:":
		serverName = id[7:]
		if s, ok := config.GlobalServers[serverName]; ok {
			server = &s
			scope = "Globale"
		}
	case len(id) > 14 && id[:14] == "projectserver:":
		rest := id[14:]
		for i := len(rest) - 1; i >= 0; i-- {
			if rest[i] == ':' {
				projectPath = rest[:i]
				serverName = rest[i+1:]
				break
			}
		}
		if project, ok := config.Projects[projectPath]; ok {
			if s, ok := project.MCPServers[serverName]; ok {
				server = &s
				scope = "Progetto: " + project.Name
			}
		}
	default:
		mw.detailPanel.RemoveAll()
		mw.detailPanel.Add(widget.NewLabel("Seleziona un server per vedere i dettagli"))
		return
	}

	if server == nil {
		mw.detailPanel.RemoveAll()
		mw.detailPanel.Add(widget.NewLabel("Server non trovato"))
		return
	}

	mw.showServerDetails(serverName, server, scope, projectPath)
}

// showServerDetails mostra i dettagli di un server
func (mw *MainWindow) showServerDetails(name string, server *domain.MCPServer, scope, projectPath string) {
	mw.detailPanel.RemoveAll()

	// Nome e scope
	mw.detailPanel.Add(widget.NewLabelWithStyle("Server: "+name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	mw.detailPanel.Add(widget.NewLabel("Scope: " + scope))
	mw.detailPanel.Add(widget.NewSeparator())

	// Tipo
	serverType := string(server.Type)
	if serverType == "" {
		if server.Command != "" {
			serverType = "stdio"
		} else if server.URL != "" {
			serverType = "http/sse"
		}
	}
	mw.detailPanel.Add(widget.NewLabel("Tipo: " + serverType))

	// Comando (per stdio)
	if server.Command != "" {
		mw.detailPanel.Add(widget.NewLabel("Comando: " + server.Command))
	}

	// Args
	if len(server.Args) > 0 {
		mw.detailPanel.Add(widget.NewLabel("Args:"))
		for _, arg := range server.Args {
			mw.detailPanel.Add(widget.NewLabel("  " + arg))
		}
	}

	// URL (per http/sse)
	if server.URL != "" {
		mw.detailPanel.Add(widget.NewLabel("URL: " + server.URL))
	}

	// Env
	if len(server.Env) > 0 {
		mw.detailPanel.Add(widget.NewSeparator())
		mw.detailPanel.Add(widget.NewLabel("Variabili d'ambiente:"))
		for k, v := range server.Env {
			mw.detailPanel.Add(widget.NewLabel("  " + k + " = " + v))
		}
	}

	// Timeout
	if server.Timeout > 0 {
		mw.detailPanel.Add(widget.NewLabel("Timeout: " + string(rune(server.Timeout)) + "ms"))
	}

	// Bottoni azione
	mw.detailPanel.Add(widget.NewSeparator())

	editBtn := widget.NewButtonWithIcon("Modifica", theme.DocumentCreateIcon(), func() {
		mw.showEditServerDialog(name, server, scope == "Globale", projectPath)
	})

	deleteBtn := widget.NewButtonWithIcon("Elimina", theme.DeleteIcon(), func() {
		mw.confirmDeleteServer(name, scope == "Globale", projectPath)
	})

	moveBtn := widget.NewButtonWithIcon("Sposta", theme.MoveDownIcon(), func() {
		mw.showMoveServerDialog(name, scope == "Globale", projectPath)
	})

	mw.detailPanel.Add(container.NewCenter(container.NewHBox(editBtn, moveBtn, deleteBtn)))
}

// showAddServerDialog mostra il dialog per aggiungere un server
func (mw *MainWindow) showAddServerDialog() {
	form := NewServerForm(mw.service, nil, "", true, "")

	d := dialog.NewCustomConfirm("Aggiungi Server MCP", "Salva", "Annulla",
		form.Container(),
		func(ok bool) {
			if ok {
				if err := form.Save(); err != nil {
					dialog.ShowError(err, mw.window)
					return
				}
				mw.refresh()
			}
		},
		mw.window,
	)
	d.Resize(fyne.NewSize(500, 400))
	d.Show()
}

// showEditServerDialog mostra il dialog per modificare un server
func (mw *MainWindow) showEditServerDialog(name string, server *domain.MCPServer, isGlobal bool, projectPath string) {
	form := NewServerForm(mw.service, server, name, isGlobal, projectPath)

	d := dialog.NewCustomConfirm("Modifica Server MCP", "Salva", "Annulla",
		form.Container(),
		func(ok bool) {
			if ok {
				if err := form.Update(); err != nil {
					dialog.ShowError(err, mw.window)
					return
				}
				mw.refresh()
			}
		},
		mw.window,
	)
	d.Resize(fyne.NewSize(500, 400))
	d.Show()
}

// confirmDeleteServer conferma l'eliminazione di un server
func (mw *MainWindow) confirmDeleteServer(name string, isGlobal bool, projectPath string) {
	dialog.ShowConfirm("Elimina Server",
		"Sei sicuro di voler eliminare il server '"+name+"'?",
		func(ok bool) {
			if ok {
				var err error
				if isGlobal {
					err = mw.service.RemoveGlobalServer(name)
				} else {
					err = mw.service.RemoveProjectServer(projectPath, name)
				}
				if err != nil {
					dialog.ShowError(err, mw.window)
					return
				}
				mw.refresh()
			}
		},
		mw.window,
	)
}

// showMoveServerDialog mostra il dialog per spostare un server
func (mw *MainWindow) showMoveServerDialog(name string, isGlobal bool, projectPath string) {
	config := mw.service.GetConfiguration()

	if isGlobal {
		// Sposta da globale a progetto
		var options []string
		for path := range config.Projects {
			options = append(options, path)
		}

		if len(options) == 0 {
			dialog.ShowInformation("Nessun Progetto",
				"Non ci sono progetti configurati in cui spostare il server.",
				mw.window)
			return
		}

		selectEntry := widget.NewSelect(options, nil)

		d := dialog.NewCustomConfirm("Sposta Server",
			"Sposta", "Annulla",
			container.NewVBox(
				widget.NewLabel("Seleziona il progetto di destinazione:"),
				selectEntry,
			),
			func(ok bool) {
				if ok && selectEntry.Selected != "" {
					if err := mw.service.MoveServerToProject(name, selectEntry.Selected); err != nil {
						dialog.ShowError(err, mw.window)
						return
					}
					mw.refresh()
				}
			},
			mw.window,
		)
		d.Show()
	} else {
		// Sposta da progetto a globale
		dialog.ShowConfirm("Sposta Server",
			"Vuoi spostare il server '"+name+"' nello scope globale?",
			func(ok bool) {
				if ok {
					if err := mw.service.MoveServerToGlobal(projectPath, name); err != nil {
						dialog.ShowError(err, mw.window)
						return
					}
					mw.refresh()
				}
			},
			mw.window,
		)
	}
}

// refresh ricarica la configurazione e aggiorna l'UI senza ricreare tutto
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
