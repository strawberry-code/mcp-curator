package ui

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/strawberry-code/mcp-curator/internal/application"
	"github.com/strawberry-code/mcp-curator/internal/domain"
	"github.com/strawberry-code/mcp-curator/internal/i18n"
	"github.com/strawberry-code/mcp-curator/internal/infrastructure"
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

// createToolbar crea la toolbar
func (mw *MainWindow) createToolbar() *fyne.Container {
	mw.addBtn = widget.NewButtonWithIcon(i18n.T("toolbar.add_server"), theme.ContentAddIcon(), func() {
		mw.showAddServerDialog()
	})

	mw.langSelect = widget.NewSelect(i18n.SupportedLangs, func(lang string) {
		mw.changeLanguage(lang)
	})
	mw.langSelect.SetSelected(string(i18n.CurrentLang()))

	mw.refreshBtn = widget.NewButtonWithIcon(i18n.T("toolbar.refresh"), theme.ViewRefreshIcon(), func() {
		mw.refresh()
	})

	rightControls := container.NewHBox(
		mw.langSelect,
		mw.refreshBtn,
	)

	toolbar := container.NewHBox(
		mw.addBtn,
		layout.NewSpacer(),
		rightControls,
	)

	// Header con padding e separatore
	return container.NewVBox(
		container.NewPadded(toolbar),
		widget.NewSeparator(),
	)
}

// changeLanguage cambia la lingua dell'interfaccia
func (mw *MainWindow) changeLanguage(lang string) {
	i18n.SetLang(lang)
}

// updateUIStrings aggiorna tutte le stringhe dell'interfaccia
func (mw *MainWindow) updateUIStrings() {
	// Aggiorna bottoni toolbar
	mw.addBtn.SetText(i18n.T("toolbar.add_server"))
	mw.refreshBtn.SetText(i18n.T("toolbar.refresh"))

	// Aggiorna tree view
	mw.tree.Refresh()

	// Aggiorna pannello dettagli
	if mw.selectedID != "" {
		mw.updateDetailPanel(mw.selectedID)
	} else {
		mw.detailPanel.RemoveAll()
		mw.detailPanel.Add(widget.NewLabel(i18n.T("detail.select_server")))
	}
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
			// Server locali di un progetto (da .mcp.json e .mcp.local.json)
			if len(id) > 8 && id[:8] == "project:" {
				projectPath := id[8:]
				if project, ok := config.Projects[projectPath]; ok {
					localServers := mw.getLocalServers(projectPath, project)
					ids := make([]string, 0, len(localServers))
					for name := range localServers {
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
			return mw.isBranchWithChildren(id)
		},
		// create
		func(branch bool) fyne.CanvasObject {
			return container.NewHBox(
				widget.NewIcon(nil),
				widget.NewLabel(""),
			)
		},
		// update
		func(id widget.TreeNodeID, branch bool, o fyne.CanvasObject) {
			config := mw.service.GetConfiguration()
			box := o.(*fyne.Container)
			icon := box.Objects[0].(*widget.Icon)
			label := box.Objects[1].(*widget.Label)

			text := mw.getNodeText(id, config)
			nodeIcon := mw.getNodeIcon(id, config)

			icon.SetResource(nodeIcon)

			if branch {
				count := mw.getChildCount(id, config)
				label.SetText(fmt.Sprintf("%s (%d)", text, count))
			} else {
				label.SetText(text)
			}
		},
	)

	tree.OnSelected = func(id widget.TreeNodeID) {
		mw.selectedID = id
		mw.toggleBranchIfNeeded(tree, id)
		mw.updateDetailPanel(id)
	}

	return tree
}

// toggleBranchIfNeeded espande o collassa un branch quando viene selezionato
func (mw *MainWindow) toggleBranchIfNeeded(tree *widget.Tree, id widget.TreeNodeID) {
	if !mw.isBranchWithChildren(id) {
		return
	}

	if tree.IsBranchOpen(id) {
		tree.CloseBranch(id)
	} else {
		tree.OpenBranch(id)
	}
}

// isBranch verifica se un nodo è strutturalmente un branch (può avere figli)
func (mw *MainWindow) isBranch(id widget.TreeNodeID) bool {
	return id == "global" || id == "projects" || (len(id) > 8 && id[:8] == "project:")
}

// isBranchWithChildren verifica se un nodo è un branch con almeno un figlio
func (mw *MainWindow) isBranchWithChildren(id widget.TreeNodeID) bool {
	// Root, global e projects sono sempre branch
	if id == "" || id == "global" || id == "projects" {
		return true
	}

	// I progetti sono branch solo se hanno server
	if len(id) > 8 && id[:8] == "project:" {
		config := mw.service.GetConfiguration()
		return mw.getChildCount(id, config) > 0
	}

	return false
}

// getNodeIcon restituisce l'icona appropriata per un nodo
func (mw *MainWindow) getNodeIcon(id widget.TreeNodeID, config *domain.Configuration) fyne.Resource {
	switch {
	case id == "global":
		return theme.HomeIcon()
	case id == "projects":
		return theme.FolderIcon()
	case len(id) > 8 && id[:8] == "project:":
		// Progetto: cartella piena o vuota in base ai server
		if mw.getChildCount(id, config) > 0 {
			return theme.FolderIcon()
		}
		return theme.FolderOpenIcon()
	case len(id) > 7 && id[:7] == "global:":
		// Server globale
		return theme.ComputerIcon()
	case len(id) > 14 && id[:14] == "projectserver:":
		// Server di progetto
		return theme.ComputerIcon()
	}
	return nil
}

// getNodeText restituisce il testo da visualizzare per un nodo
func (mw *MainWindow) getNodeText(id widget.TreeNodeID, config *domain.Configuration) string {
	switch {
	case id == "global":
		return i18n.T("tree.global")
	case id == "projects":
		return i18n.T("tree.projects")
	case len(id) > 7 && id[:7] == "global:":
		return id[7:]
	case len(id) > 8 && id[:8] == "project:":
		path := id[8:]
		if project, ok := config.Projects[path]; ok {
			return project.Name
		}
		return path
	case len(id) > 14 && id[:14] == "projectserver:":
		rest := id[14:]
		for i := len(rest) - 1; i >= 0; i-- {
			if rest[i] == ':' {
				return rest[i+1:]
			}
		}
	}
	return id
}

// getChildCount restituisce il numero di figli di un nodo branch
func (mw *MainWindow) getChildCount(id widget.TreeNodeID, config *domain.Configuration) int {
	switch {
	case id == "global":
		return len(config.GlobalServers)
	case id == "projects":
		return len(config.Projects)
	case len(id) > 8 && id[:8] == "project:":
		path := id[8:]
		if project, ok := config.Projects[path]; ok {
			// Conta solo i server locali (da .mcp.json e .mcp.local.json)
			return mw.countLocalServers(path, project)
		}
	}
	return 0
}

// countLocalServers conta i server MCP locali di un progetto
func (mw *MainWindow) countLocalServers(path string, project *domain.Project) int {
	return len(mw.getLocalServers(path, project))
}

// getLocalServers restituisce i server MCP locali di un progetto
func (mw *MainWindow) getLocalServers(path string, project *domain.Project) map[string]domain.MCPServer {
	localServers := make(map[string]domain.MCPServer)
	if project.HasMCPJson {
		mcpJsonPath := filepath.Join(path, ".mcp.json")
		servers := infrastructure.LoadMCPFileServers(mcpJsonPath)
		for name, server := range servers {
			localServers[name] = server
		}
	}
	if project.HasMCPLocal {
		mcpLocalPath := filepath.Join(path, ".mcp.local.json")
		servers := infrastructure.LoadMCPFileServers(mcpLocalPath)
		for name, server := range servers {
			localServers[name] = server
		}
	}
	return localServers
}

// updateDetailPanel aggiorna il pannello dettagli
func (mw *MainWindow) updateDetailPanel(id string) {
	config := mw.service.GetConfiguration()

	switch {
	case len(id) > 8 && id[:8] == "project:":
		projectPath := id[8:]
		if project, ok := config.Projects[projectPath]; ok {
			mw.showProjectDetails(projectPath, project)
			return
		}
	case len(id) > 7 && id[:7] == "global:":
		serverName := id[7:]
		if s, ok := config.GlobalServers[serverName]; ok {
			mw.showServerDetails(serverName, &s, i18n.T("tree.global"), "")
			return
		}
	case len(id) > 14 && id[:14] == "projectserver:":
		rest := id[14:]
		for i := len(rest) - 1; i >= 0; i-- {
			if rest[i] == ':' {
				projectPath := rest[:i]
				serverName := rest[i+1:]
				if project, ok := config.Projects[projectPath]; ok {
					// Cerca nei server locali
					localServers := mw.getLocalServers(projectPath, project)
					if s, ok := localServers[serverName]; ok {
						mw.showServerDetails(serverName, &s, i18n.T("detail.project")+": "+project.Name, projectPath)
						return
					}
				}
				break
			}
		}
	}

	mw.detailPanel.RemoveAll()
	mw.detailPanel.Add(widget.NewLabel(i18n.T("detail.select_item")))
}

// showProjectDetails mostra i dettagli di un progetto
func (mw *MainWindow) showProjectDetails(path string, project *domain.Project) {
	mw.detailPanel.RemoveAll()

	// Nome progetto
	mw.detailPanel.Add(widget.NewLabelWithStyle(i18n.T("detail.project")+": "+project.Name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	mw.detailPanel.Add(widget.NewSeparator())

	// Path con bottoni Copy e Open
	pathLabel := widget.NewLabel(i18n.T("detail.path") + ": " + path)

	copyBtn := widget.NewButtonWithIcon("", theme.ContentCopyIcon(), func() {
		mw.window.Clipboard().SetContent(path)
	})
	copyBtn.Importance = widget.LowImportance

	openBtn := widget.NewButtonWithIcon("", theme.FolderOpenIcon(), func() {
		mw.openFileWithDefaultApp(path)
	})
	openBtn.Importance = widget.LowImportance

	pathRow := container.NewHBox(pathLabel, copyBtn, openBtn)
	mw.detailPanel.Add(pathRow)

	// Ottieni configurazione per contare i server globali
	config := mw.service.GetConfiguration()
	globalCount := len(config.GlobalServers)

	// Carica server locali da file .mcp.json e .mcp.local.json
	localServers := make(map[string]domain.MCPServer)
	var localFiles []string

	if project.HasMCPJson {
		mcpJsonPath := filepath.Join(path, ".mcp.json")
		servers := infrastructure.LoadMCPFileServers(mcpJsonPath)
		for name, server := range servers {
			localServers[name] = server
		}
		if len(servers) > 0 {
			localFiles = append(localFiles, ".mcp.json")
		}
	}
	if project.HasMCPLocal {
		mcpLocalPath := filepath.Join(path, ".mcp.local.json")
		servers := infrastructure.LoadMCPFileServers(mcpLocalPath)
		for name, server := range servers {
			localServers[name] = server
		}
		if len(servers) > 0 {
			localFiles = append(localFiles, ".mcp.local.json")
		}
	}
	localCount := len(localServers)

	// Sezione Configurazioni Globali (collassata di default)
	mw.detailPanel.Add(widget.NewSeparator())
	globalContent := container.NewVBox()
	if globalCount > 0 {
		var globalNames []string
		for name := range config.GlobalServers {
			globalNames = append(globalNames, name)
		}
		sort.Strings(globalNames)
		for _, name := range globalNames {
			globalContent.Add(widget.NewLabel("  • " + name))
		}
	}
	globalAccordion := widget.NewAccordion(
		widget.NewAccordionItem(
			fmt.Sprintf("%s (%d)", i18n.T("detail.global_configs"), globalCount),
			globalContent,
		),
	)
	// Globali: collassato di default (non apriamo nessun item)
	mw.detailPanel.Add(globalAccordion)

	// Sezione Configurazioni Locali (espansa di default)
	mw.detailPanel.Add(widget.NewSeparator())
	localContent := container.NewVBox()
	if localCount > 0 {
		// Mostra i file di configurazione locali
		for _, file := range localFiles {
			filePath := filepath.Join(path, file)
			localContent.Add(mw.createConfigFileLink(file, filePath))
		}
		// Lista nomi server locali
		var localNames []string
		for name := range localServers {
			localNames = append(localNames, name)
		}
		sort.Strings(localNames)
		for _, name := range localNames {
			localContent.Add(widget.NewLabel("  • " + name))
		}
	} else {
		localContent.Add(widget.NewLabel("  " + i18n.T("detail.no_local_servers")))
		// Mostra link ai file locali se esistono ma sono vuoti
		if project.HasMCPJson {
			mcpJsonPath := filepath.Join(path, ".mcp.json")
			localContent.Add(mw.createConfigFileLink(".mcp.json (vuoto)", mcpJsonPath))
		}
		if project.HasMCPLocal {
			mcpLocalPath := filepath.Join(path, ".mcp.local.json")
			localContent.Add(mw.createConfigFileLink(".mcp.local.json (vuoto)", mcpLocalPath))
		}
	}
	localAccordion := widget.NewAccordion(
		widget.NewAccordionItem(
			fmt.Sprintf("%s (%d)", i18n.T("detail.local_configs"), localCount),
			localContent,
		),
	)
	// Locali: espanso di default
	localAccordion.Open(0)
	mw.detailPanel.Add(localAccordion)

	// Bottone per aggiungere server al progetto
	mw.detailPanel.Add(widget.NewSeparator())
	addServerBtn := widget.NewButtonWithIcon(i18n.T("btn.add_server"), theme.ContentAddIcon(), func() {
		mw.showAddServerToProjectDialog(path)
	})
	mw.detailPanel.Add(container.NewCenter(addServerBtn))
}

// createConfigFileLink crea un link cliccabile per un file di configurazione
func (mw *MainWindow) createConfigFileLink(displayName, filePath string) *fyne.Container {
	btn := widget.NewButtonWithIcon(displayName, theme.FileIcon(), func() {
		mw.openFileWithDefaultApp(filePath)
	})
	btn.Importance = widget.LowImportance
	return container.NewHBox(layout.NewSpacer(), btn, layout.NewSpacer())
}

// openFileWithDefaultApp apre un file con l'applicazione di default del sistema
func (mw *MainWindow) openFileWithDefaultApp(filePath string) {
	var cmd *exec.Cmd
	// macOS usa "open", Linux "xdg-open", Windows "start"
	cmd = exec.Command("open", filePath)
	if err := cmd.Start(); err != nil {
		dialog.ShowError(fmt.Errorf("impossibile aprire il file: %v", err), mw.window)
	}
}

// showServerDetails mostra i dettagli di un server
func (mw *MainWindow) showServerDetails(name string, server *domain.MCPServer, scope, projectPath string) {
	mw.detailPanel.RemoveAll()

	isGlobal := scope == i18n.T("tree.global")

	// Nome e scope
	mw.detailPanel.Add(widget.NewLabelWithStyle(i18n.T("detail.server")+": "+name, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
	mw.detailPanel.Add(widget.NewLabel(i18n.T("detail.scope")+": "+scope))
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
	mw.detailPanel.Add(widget.NewLabel(i18n.T("detail.type")+": "+serverType))

	// Comando (per stdio)
	if server.Command != "" {
		mw.detailPanel.Add(widget.NewLabel(i18n.T("detail.command")+": "+server.Command))
	}

	// Args
	if len(server.Args) > 0 {
		mw.detailPanel.Add(widget.NewLabel(i18n.T("detail.args")+":"))
		for _, arg := range server.Args {
			mw.detailPanel.Add(widget.NewLabel("  " + arg))
		}
	}

	// URL (per http/sse)
	if server.URL != "" {
		mw.detailPanel.Add(widget.NewLabel(i18n.T("detail.url")+": "+server.URL))
	}

	// Env
	if len(server.Env) > 0 {
		mw.detailPanel.Add(widget.NewSeparator())
		mw.detailPanel.Add(widget.NewLabel(i18n.T("detail.env")+":"))
		for k, v := range server.Env {
			mw.detailPanel.Add(widget.NewLabel("  " + k + " = " + v))
		}
	}

	// Timeout
	if server.Timeout > 0 {
		mw.detailPanel.Add(widget.NewLabel(i18n.T("detail.timeout")+": "+fmt.Sprintf("%dms", server.Timeout)))
	}

	// Bottoni azione
	mw.detailPanel.Add(widget.NewSeparator())

	editBtn := widget.NewButtonWithIcon(i18n.T("btn.edit"), theme.DocumentCreateIcon(), func() {
		mw.showEditServerDialog(name, server, isGlobal, projectPath)
	})

	deleteBtn := widget.NewButtonWithIcon(i18n.T("btn.delete"), theme.DeleteIcon(), func() {
		mw.confirmDeleteServer(name, isGlobal, projectPath)
	})

	moveBtn := widget.NewButtonWithIcon(i18n.T("btn.move"), theme.MoveDownIcon(), func() {
		mw.showMoveServerDialog(name, isGlobal, projectPath)
	})

	cloneBtn := widget.NewButtonWithIcon(i18n.T("btn.clone"), theme.ContentCopyIcon(), func() {
		mw.showCloneServerDialog(name, server, isGlobal, projectPath)
	})

	mw.detailPanel.Add(container.NewCenter(container.NewHBox(editBtn, moveBtn, cloneBtn, deleteBtn)))
}

// showAddServerDialog mostra il dialog per aggiungere un server
func (mw *MainWindow) showAddServerDialog() {
	form := NewServerForm(mw.service, nil, "", true, "")

	d := dialog.NewCustomConfirm(i18n.T("dialog.add_server"), i18n.T("btn.save"), i18n.T("btn.cancel"),
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

// showAddServerToProjectDialog mostra il dialog per aggiungere un server a un progetto specifico
func (mw *MainWindow) showAddServerToProjectDialog(projectPath string) {
	form := NewServerForm(mw.service, nil, "", false, projectPath)

	d := dialog.NewCustomConfirm(i18n.T("dialog.add_server_to_project"), i18n.T("btn.save"), i18n.T("btn.cancel"),
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

	d := dialog.NewCustomConfirm(i18n.T("dialog.edit_server"), i18n.T("btn.save"), i18n.T("btn.cancel"),
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
	dialog.ShowConfirm(i18n.T("dialog.delete_server"),
		fmt.Sprintf(i18n.T("dialog.delete_confirm"), name),
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
			dialog.ShowInformation(i18n.T("dialog.no_projects"),
				i18n.T("dialog.no_projects_msg"),
				mw.window)
			return
		}

		selectEntry := widget.NewSelect(options, nil)

		d := dialog.NewCustomConfirm(i18n.T("dialog.move_server"),
			i18n.T("btn.move"), i18n.T("btn.cancel"),
			container.NewVBox(
				widget.NewLabel(i18n.T("dialog.move_to_project")),
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
		dialog.ShowConfirm(i18n.T("dialog.move_server"),
			fmt.Sprintf(i18n.T("dialog.move_to_global"), name),
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

// showCloneServerDialog mostra il dialog per clonare un server su altri scope
func (mw *MainWindow) showCloneServerDialog(name string, server *domain.MCPServer, isGlobal bool, projectPath string) {
	config := mw.service.GetConfiguration()

	// Costruisci lista destinazioni: Globale + tutti i progetti
	var options []string

	// Aggiungi "Globale" solo se il server non è già globale
	if !isGlobal {
		options = append(options, i18n.T("form.scope_global"))
	}

	// Aggiungi tutti i progetti (escluso quello corrente se il server è di progetto)
	for path := range config.Projects {
		if !isGlobal && path == projectPath {
			continue // Salta il progetto corrente
		}
		options = append(options, path)
	}

	if len(options) == 0 {
		dialog.ShowInformation(i18n.T("dialog.no_projects"),
			i18n.T("dialog.no_projects_msg"),
			mw.window)
		return
	}

	// Usa CheckGroup per selezione multipla
	checkGroup := widget.NewCheckGroup(options, nil)

	content := container.NewVBox(
		widget.NewLabel(i18n.T("dialog.clone_to")),
		checkGroup,
	)

	d := dialog.NewCustomConfirm(i18n.T("dialog.clone_server"),
		i18n.T("btn.save"), i18n.T("btn.cancel"),
		container.NewScroll(content),
		func(ok bool) {
			if ok && len(checkGroup.Selected) > 0 {
				var clonedCount int
				for _, dest := range checkGroup.Selected {
					var err error
					if dest == i18n.T("form.scope_global") {
						// Clona su globale
						err = mw.service.AddGlobalServer(name, *server)
					} else {
						// Clona su progetto
						err = mw.service.AddProjectServer(dest, name, *server)
					}
					if err != nil {
						dialog.ShowError(err, mw.window)
						continue
					}
					clonedCount++
				}
				if clonedCount > 0 {
					mw.refresh()
					dialog.ShowInformation(i18n.T("dialog.clone_server"),
						fmt.Sprintf(i18n.T("dialog.clone_success"), name),
						mw.window)
				}
			}
		},
		mw.window,
	)
	d.Resize(fyne.NewSize(500, 400))
	d.Show()
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
