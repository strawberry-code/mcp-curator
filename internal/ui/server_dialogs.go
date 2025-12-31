package ui

import (
	"encoding/json"
	"errors"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/strawberry-code/mcp-curator/internal/domain"
	"github.com/strawberry-code/mcp-curator/internal/i18n"
)

// showAddServerDialog mostra la scelta del metodo di aggiunta (form o JSON)
func (mw *MainWindow) showAddServerDialog() {
	mw.showAddMethodDialog(true, "")
}

// showAddMethodDialog mostra la scelta tra form e JSON per aggiungere un server
func (mw *MainWindow) showAddMethodDialog(isGlobal bool, projectPath string) {
	formBtn := widget.NewButtonWithIcon(i18n.T("dialog.add_via_form"), theme.DocumentCreateIcon(), func() {})
	jsonBtn := widget.NewButtonWithIcon(i18n.T("dialog.add_via_json"), theme.ContentPasteIcon(), func() {})

	content := container.NewVBox(
		widget.NewLabel(i18n.T("dialog.add_method")),
		widget.NewSeparator(),
		container.NewGridWithColumns(2, formBtn, jsonBtn),
	)

	d := dialog.NewCustomWithoutButtons(i18n.T("dialog.add_server"), content, mw.window)

	formBtn.OnTapped = func() {
		d.Hide()
		mw.showAddServerFormDialog(isGlobal, projectPath)
	}

	jsonBtn.OnTapped = func() {
		d.Hide()
		mw.showAddServerJSONDialog(isGlobal, projectPath)
	}

	d.Resize(fyne.NewSize(400, 150))
	d.Show()
}

// showAddServerFormDialog mostra il dialog form per aggiungere un server
func (mw *MainWindow) showAddServerFormDialog(isGlobal bool, projectPath string) {
	form := NewServerForm(mw.service, nil, "", isGlobal, projectPath)

	title := i18n.T("dialog.add_server")
	if !isGlobal && projectPath != "" {
		title = i18n.T("dialog.add_server_to_project")
	}

	d := dialog.NewCustomConfirm(title, i18n.T("btn.save"), i18n.T("btn.cancel"),
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

// showAddServerToProjectDialog mostra la scelta del metodo di aggiunta per un progetto
func (mw *MainWindow) showAddServerToProjectDialog(projectPath string) {
	mw.showAddMethodDialog(false, projectPath)
}

// showAddServerJSONDialog mostra il dialog per aggiungere un server tramite JSON raw
func (mw *MainWindow) showAddServerJSONDialog(isGlobal bool, projectPath string) {
	jsonEntry := widget.NewMultiLineEntry()
	jsonEntry.SetPlaceHolder(i18n.T("dialog.add_json_hint"))
	jsonEntry.Wrapping = fyne.TextWrapWord

	content := container.NewBorder(
		widget.NewLabel(i18n.T("dialog.add_json_hint")),
		nil, nil, nil,
		jsonEntry,
	)

	d := dialog.NewCustomConfirm(i18n.T("dialog.add_json_title"), i18n.T("btn.save"), i18n.T("btn.cancel"),
		content,
		func(ok bool) {
			if !ok {
				return
			}

			name, server, err := mw.parseServerJSON(jsonEntry.Text)
			if err != nil {
				dialog.ShowError(err, mw.window)
				return
			}

			if err := mw.saveServer(name, server, isGlobal, projectPath); err != nil {
				dialog.ShowError(err, mw.window)
				return
			}

			mw.refresh()
		},
		mw.window,
	)
	d.Resize(fyne.NewSize(500, 400))
	d.Show()
}

// parseServerJSON valida e converte il JSON in un MCPServer
func (mw *MainWindow) parseServerJSON(jsonText string) (string, domain.MCPServer, error) {
	if jsonText == "" {
		return "", domain.MCPServer{}, errors.New(i18n.T("dialog.json_invalid"))
	}

	var serverData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonText), &serverData); err != nil {
		return "", domain.MCPServer{}, fmt.Errorf("%s: %v", i18n.T("dialog.json_invalid"), err)
	}

	name, server, err := mw.service.ParseServerFromJSON(serverData)
	if err != nil {
		return "", domain.MCPServer{}, errors.New(i18n.T("dialog.json_missing_fields"))
	}

	return name, server, nil
}

// saveServer salva un server nello scope appropriato
func (mw *MainWindow) saveServer(name string, server domain.MCPServer, isGlobal bool, projectPath string) error {
	if isGlobal {
		return mw.service.AddGlobalServer(name, server)
	}
	return mw.service.AddProjectServer(projectPath, name, server)
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

	// Copia il server per evitare problemi con closure
	serverCopy := *server

	d := dialog.NewCustomConfirm(i18n.T("dialog.clone_server"),
		i18n.T("btn.save"), i18n.T("btn.cancel"),
		container.NewScroll(content),
		func(ok bool) {
			if ok && len(checkGroup.Selected) > 0 {
				var clonedCount int
				var lastErr error
				for _, dest := range checkGroup.Selected {
					var err error
					if dest == i18n.T("form.scope_global") {
						// Clona su globale
						err = mw.service.AddGlobalServer(name, serverCopy)
					} else {
						// Clona su progetto
						err = mw.service.AddProjectServer(dest, name, serverCopy)
					}
					if err != nil {
						lastErr = err
						continue
					}
					clonedCount++
				}
				if clonedCount > 0 {
					mw.refresh()
					dialog.ShowInformation(i18n.T("dialog.clone_server"),
						fmt.Sprintf(i18n.T("dialog.clone_success"), name),
						mw.window)
				} else if lastErr != nil {
					dialog.ShowError(lastErr, mw.window)
				}
			}
		},
		mw.window,
	)
	d.Resize(fyne.NewSize(500, 400))
	d.Show()
}
