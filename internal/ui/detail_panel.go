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

	"github.com/strawberry-code/mcp-curator/internal/domain"
	"github.com/strawberry-code/mcp-curator/internal/i18n"
	"github.com/strawberry-code/mcp-curator/internal/infrastructure"
)

// updateDetailPanel aggiorna il pannello dettagli in base all'elemento selezionato
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
		mw.app.Clipboard().SetContent(path)
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

	// Carica server del progetto da tutte le sorgenti
	localServers := make(map[string]domain.MCPServer)
	var localFiles []string

	// Server da ~/.claude.json projects.[path].mcpServers
	projectSettingsCount := len(project.MCPServers)
	for name, server := range project.MCPServers {
		localServers[name] = server
	}
	if projectSettingsCount > 0 {
		localFiles = append(localFiles, "~/.claude.json")
	}

	// Server da file .mcp.json
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

	// Server da file .mcp.local.json
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
		// Mostra i file di configurazione
		for _, file := range localFiles {
			var filePath string
			if file == "~/.claude.json" {
				filePath = mw.service.GetConfigPath()
			} else {
				filePath = filepath.Join(path, file)
			}
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

// showServerDetails mostra i dettagli di un server MCP
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
	// macOS usa "open", Linux "xdg-open", Windows "start"
	cmd := exec.Command("open", filePath)
	if err := cmd.Start(); err != nil {
		dialog.ShowError(fmt.Errorf("impossibile aprire il file: %v", err), mw.window)
	}
}
