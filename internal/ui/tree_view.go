package ui

import (
	"fmt"
	"path/filepath"
	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/strawberry-code/mcp-curator/internal/domain"
	"github.com/strawberry-code/mcp-curator/internal/i18n"
	"github.com/strawberry-code/mcp-curator/internal/infrastructure"
)

// createTree crea il widget tree per navigare scope e server
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

// getLocalServers restituisce i server MCP di un progetto (da ~/.claude.json e file locali)
func (mw *MainWindow) getLocalServers(path string, project *domain.Project) map[string]domain.MCPServer {
	localServers := make(map[string]domain.MCPServer)

	// Prima aggiungi i server da ~/.claude.json projects.[path].mcpServers
	for name, server := range project.MCPServers {
		localServers[name] = server
	}

	// Poi aggiungi/sovrascrivi con i server dai file locali (.mcp.json e .mcp.local.json)
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
