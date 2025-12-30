package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/strawberry-code/mcp-curator/internal/application"
	"github.com/strawberry-code/mcp-curator/internal/domain"
)

// ServerForm è il form per aggiungere/modificare un server
type ServerForm struct {
	service     *application.MCPService
	server      *domain.MCPServer
	name        string
	isGlobal    bool
	projectPath string

	nameEntry     *widget.Entry
	typeSelect    *widget.Select
	commandEntry  *widget.Entry
	argsEntry     *widget.Entry
	urlEntry      *widget.Entry
	envEntry      *widget.Entry
	globalRadio   *widget.RadioGroup
	projectSelect *widget.Select

	container *fyne.Container
}

// NewServerForm crea un nuovo form
func NewServerForm(service *application.MCPService, server *domain.MCPServer, name string, isGlobal bool, projectPath string) *ServerForm {
	sf := &ServerForm{
		service:     service,
		server:      server,
		name:        name,
		isGlobal:    isGlobal,
		projectPath: projectPath,
	}

	sf.build()
	return sf
}

// build costruisce il form
func (sf *ServerForm) build() {
	// Nome
	sf.nameEntry = widget.NewEntry()
	sf.nameEntry.SetPlaceHolder("Nome del server")
	if sf.name != "" {
		sf.nameEntry.SetText(sf.name)
		sf.nameEntry.Disable()
	}

	// Tipo
	sf.typeSelect = widget.NewSelect([]string{"stdio", "http", "sse"}, nil)
	sf.typeSelect.SetSelected("stdio")

	// Command
	sf.commandEntry = widget.NewEntry()
	sf.commandEntry.SetPlaceHolder("Comando (es: uvx, npx)")

	// Args
	sf.argsEntry = widget.NewMultiLineEntry()
	sf.argsEntry.SetPlaceHolder("Argomenti (uno per riga)")
	sf.argsEntry.SetMinRowsVisible(3)

	// URL
	sf.urlEntry = widget.NewEntry()
	sf.urlEntry.SetPlaceHolder("URL (per http/sse)")

	// Env
	sf.envEntry = widget.NewMultiLineEntry()
	sf.envEntry.SetPlaceHolder("Variabili ambiente (KEY=value, una per riga)")
	sf.envEntry.SetMinRowsVisible(2)

	// Scope
	config := sf.service.GetConfiguration()
	var projectOptions []string
	for path := range config.Projects {
		projectOptions = append(projectOptions, path)
	}

	sf.projectSelect = widget.NewSelect(projectOptions, nil)
	sf.globalRadio = widget.NewRadioGroup([]string{"Globale", "Progetto"}, func(selected string) {
		sf.projectSelect.Enable()
		if selected == "Globale" {
			sf.projectSelect.Disable()
		}
	})
	sf.globalRadio.Horizontal = true

	// Popola se modifica
	if sf.server != nil {
		if sf.server.Type != "" {
			sf.typeSelect.SetSelected(string(sf.server.Type))
		}
		sf.commandEntry.SetText(sf.server.Command)
		if len(sf.server.Args) > 0 {
			sf.argsEntry.SetText(strings.Join(sf.server.Args, "\n"))
		}
		sf.urlEntry.SetText(sf.server.URL)
		if len(sf.server.Env) > 0 {
			var envLines []string
			for k, v := range sf.server.Env {
				envLines = append(envLines, k+"="+v)
			}
			sf.envEntry.SetText(strings.Join(envLines, "\n"))
		}
	}

	// Imposta scope
	if sf.isGlobal {
		sf.globalRadio.SetSelected("Globale")
		sf.projectSelect.Disable()
	} else {
		sf.globalRadio.SetSelected("Progetto")
		sf.projectSelect.SetSelected(sf.projectPath)
	}

	// Se modifica, disabilita cambio scope
	if sf.server != nil {
		sf.globalRadio.Disable()
		sf.projectSelect.Disable()
	}

	// Costruisci container
	sf.container = container.NewVBox(
		widget.NewLabel("Nome:"),
		sf.nameEntry,
		widget.NewSeparator(),
		widget.NewLabel("Scope:"),
		sf.globalRadio,
		sf.projectSelect,
		widget.NewSeparator(),
		widget.NewLabel("Tipo:"),
		sf.typeSelect,
		widget.NewLabel("Comando:"),
		sf.commandEntry,
		widget.NewLabel("Argomenti:"),
		sf.argsEntry,
		widget.NewLabel("URL:"),
		sf.urlEntry,
		widget.NewLabel("Variabili Ambiente:"),
		sf.envEntry,
	)
}

// Container restituisce il container del form
func (sf *ServerForm) Container() *fyne.Container {
	return sf.container
}

// getServer costruisce il server dai dati del form
func (sf *ServerForm) getServer() domain.MCPServer {
	server := domain.MCPServer{
		Type:    domain.ServerType(sf.typeSelect.Selected),
		Command: sf.commandEntry.Text,
		URL:     sf.urlEntry.Text,
	}

	// Parse args
	argsText := strings.TrimSpace(sf.argsEntry.Text)
	if argsText != "" {
		server.Args = strings.Split(argsText, "\n")
	}

	// Parse env
	envText := strings.TrimSpace(sf.envEntry.Text)
	if envText != "" {
		server.Env = make(map[string]string)
		for _, line := range strings.Split(envText, "\n") {
			line = strings.TrimSpace(line)
			if idx := strings.Index(line, "="); idx > 0 {
				key := strings.TrimSpace(line[:idx])
				value := strings.TrimSpace(line[idx+1:])
				server.Env[key] = value
			}
		}
	}

	return server
}

// Save salva un nuovo server
func (sf *ServerForm) Save() error {
	name := strings.TrimSpace(sf.nameEntry.Text)
	server := sf.getServer()

	if sf.globalRadio.Selected == "Globale" {
		return sf.service.AddGlobalServer(name, server)
	}

	projectPath := sf.projectSelect.Selected
	return sf.service.AddProjectServer(projectPath, name, server)
}

// Update aggiorna un server esistente
func (sf *ServerForm) Update() error {
	server := sf.getServer()

	if sf.isGlobal {
		return sf.service.UpdateGlobalServer(sf.name, server)
	}

	return sf.service.UpdateProjectServer(sf.projectPath, sf.name, server)
}
