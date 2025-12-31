package i18n

import "sync"

// Lang rappresenta una lingua supportata
type Lang string

const (
	IT Lang = "IT"
	EN Lang = "EN"
	FR Lang = "FR"
	DE Lang = "DE"
	ES Lang = "ES"
	PT Lang = "PT"
	JA Lang = "JA"
	KO Lang = "KO"
	CN Lang = "CN"
	UK Lang = "UK"
)

// SupportedLangs lista delle lingue supportate
var SupportedLangs = []string{"IT", "EN", "FR", "DE", "ES", "PT", "JA", "KO", "CN", "UK"}

// Translator gestisce le traduzioni
type Translator struct {
	mu          sync.RWMutex
	currentLang Lang
	strings     map[Lang]map[string]string
	onChange    []func(Lang)
}

var defaultTranslator = NewTranslator()

// NewTranslator crea un nuovo translator
func NewTranslator() *Translator {
	t := &Translator{
		currentLang: IT,
		strings:     make(map[Lang]map[string]string),
		onChange:    make([]func(Lang), 0),
	}
	t.loadStrings()
	return t
}

// Get restituisce il translator globale
func Get() *Translator {
	return defaultTranslator
}

// T è una scorciatoia per tradurre una stringa
func T(key string) string {
	return defaultTranslator.Translate(key)
}

// SetLang imposta la lingua corrente
func SetLang(lang string) {
	defaultTranslator.SetLanguage(Lang(lang))
}

// CurrentLang restituisce la lingua corrente
func CurrentLang() Lang {
	return defaultTranslator.GetLanguage()
}

// OnChange registra un callback per quando cambia la lingua
func OnChange(fn func(Lang)) {
	defaultTranslator.OnLanguageChange(fn)
}

// SetLanguage imposta la lingua
func (t *Translator) SetLanguage(lang Lang) {
	t.mu.Lock()
	t.currentLang = lang
	callbacks := make([]func(Lang), len(t.onChange))
	copy(callbacks, t.onChange)
	t.mu.Unlock()

	// Notifica i listener
	for _, fn := range callbacks {
		fn(lang)
	}
}

// GetLanguage restituisce la lingua corrente
func (t *Translator) GetLanguage() Lang {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.currentLang
}

// Translate traduce una chiave
func (t *Translator) Translate(key string) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if langStrings, ok := t.strings[t.currentLang]; ok {
		if val, ok := langStrings[key]; ok {
			return val
		}
	}
	// Fallback a IT
	if langStrings, ok := t.strings[IT]; ok {
		if val, ok := langStrings[key]; ok {
			return val
		}
	}
	return key
}

// OnLanguageChange registra un callback
func (t *Translator) OnLanguageChange(fn func(Lang)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.onChange = append(t.onChange, fn)
}

// loadStrings carica le stringhe di traduzione
func (t *Translator) loadStrings() {
	// Italiano (default)
	t.strings[IT] = map[string]string{
		// Toolbar
		"toolbar.add_server": "Aggiungi Server",
		"toolbar.refresh":    "Aggiorna",

		// Tree
		"tree.global":   "Globale",
		"tree.projects": "Progetti",

		// Detail panel
		"detail.select_item":   "Seleziona un elemento per vedere i dettagli",
		"detail.select_server": "Seleziona un server per vedere i dettagli",
		"detail.project":       "Progetto",
		"detail.server":        "Server",
		"detail.scope":         "Scope",
		"detail.type":          "Tipo",
		"detail.command":       "Comando",
		"detail.args":          "Args",
		"detail.url":           "URL",
		"detail.env":           "Variabili d'ambiente",
		"detail.timeout":       "Timeout",
		"detail.path":              "Path",
		"detail.mcp_servers":       "Server MCP",
		"detail.configs":           "Configurazioni",
		"detail.global_configs":    "Configurazioni Globali",
		"detail.local_configs":     "Configurazioni Locali",
		"detail.no_local_servers":  "Nessun server MCP locale configurato",

		// Buttons
		"btn.edit":       "Modifica",
		"btn.delete":     "Elimina",
		"btn.move":       "Sposta",
		"btn.save":       "Salva",
		"btn.cancel":     "Annulla",
		"btn.add_server": "Aggiungi Server",

		// Dialogs
		"dialog.add_server":            "Aggiungi Server MCP",
		"dialog.add_server_to_project": "Aggiungi Server MCP al Progetto",
		"dialog.edit_server":           "Modifica Server MCP",
		"dialog.delete_server":         "Elimina Server",
		"dialog.delete_confirm":        "Sei sicuro di voler eliminare il server '%s'?",
		"dialog.move_server":           "Sposta Server",
		"dialog.move_to_project":       "Seleziona il progetto di destinazione:",
		"dialog.move_to_global":        "Vuoi spostare il server '%s' nello scope globale?",
		"dialog.no_projects":           "Nessun Progetto",
		"dialog.no_projects_msg":       "Non ci sono progetti configurati in cui spostare il server.",

		// Form
		"form.name":        "Nome",
		"form.name_hint":   "Nome del server",
		"form.scope":       "Scope",
		"form.scope_global":  "Globale",
		"form.scope_project": "Progetto",
		"form.type":        "Tipo",
		"form.command":     "Comando",
		"form.command_hint": "Comando (es: uvx, npx)",
		"form.args":        "Argomenti",
		"form.args_hint":   "Argomenti separati da spazio (es: -y mcp-server)",
		"form.url":         "URL",
		"form.url_hint":    "URL (per http/sse)",
		"form.env":         "Variabili Ambiente",
		"form.env_hint":    "KEY=value, KEY2=value2 (separati da virgola)",
	}

	// English
	t.strings[EN] = map[string]string{
		"toolbar.add_server": "Add Server",
		"toolbar.refresh":    "Refresh",
		"tree.global":        "Global",
		"tree.projects":      "Projects",
		"detail.select_item":   "Select an item to view details",
		"detail.select_server": "Select a server to view details",
		"detail.project":       "Project",
		"detail.server":        "Server",
		"detail.scope":         "Scope",
		"detail.type":          "Type",
		"detail.command":       "Command",
		"detail.args":          "Args",
		"detail.url":           "URL",
		"detail.env":           "Environment Variables",
		"detail.timeout":       "Timeout",
		"detail.path":          "Path",
		"detail.mcp_servers":       "MCP Servers",
		"detail.configs":           "Configurations",
		"detail.global_configs":    "Global Configurations",
		"detail.local_configs":     "Local Configurations",
		"detail.no_local_servers":  "No local MCP servers configured",
		"btn.edit":             "Edit",
		"btn.delete":           "Delete",
		"btn.move":             "Move",
		"btn.save":             "Save",
		"btn.cancel":           "Cancel",
		"btn.add_server":       "Add Server",
		"dialog.add_server":            "Add MCP Server",
		"dialog.add_server_to_project": "Add MCP Server to Project",
		"dialog.edit_server":           "Edit MCP Server",
		"dialog.delete_server":         "Delete Server",
		"dialog.delete_confirm":        "Are you sure you want to delete the server '%s'?",
		"dialog.move_server":           "Move Server",
		"dialog.move_to_project":       "Select destination project:",
		"dialog.move_to_global":        "Do you want to move the server '%s' to global scope?",
		"dialog.no_projects":           "No Projects",
		"dialog.no_projects_msg":       "There are no configured projects to move the server to.",
		"form.name":           "Name",
		"form.name_hint":      "Server name",
		"form.scope":          "Scope",
		"form.scope_global":   "Global",
		"form.scope_project":  "Project",
		"form.type":           "Type",
		"form.command":        "Command",
		"form.command_hint":   "Command (e.g.: uvx, npx)",
		"form.args":           "Arguments",
		"form.args_hint":      "Space-separated arguments (e.g.: -y mcp-server)",
		"form.url":            "URL",
		"form.url_hint":       "URL (for http/sse)",
		"form.env":            "Environment Variables",
		"form.env_hint":       "KEY=value, KEY2=value2 (comma-separated)",
	}

	// French
	t.strings[FR] = map[string]string{
		"toolbar.add_server": "Ajouter Serveur",
		"toolbar.refresh":    "Actualiser",
		"tree.global":        "Global",
		"tree.projects":      "Projets",
		"detail.select_item":   "Sélectionnez un élément pour voir les détails",
		"detail.select_server": "Sélectionnez un serveur pour voir les détails",
		"detail.project":       "Projet",
		"detail.server":        "Serveur",
		"detail.scope":         "Portée",
		"detail.type":          "Type",
		"detail.command":       "Commande",
		"detail.args":          "Args",
		"detail.url":           "URL",
		"detail.env":           "Variables d'environnement",
		"detail.timeout":       "Délai",
		"detail.path":          "Chemin",
		"detail.mcp_servers":       "Serveurs MCP",
		"detail.configs":           "Configurations",
		"detail.global_configs":    "Configurations Globales",
		"detail.local_configs":     "Configurations Locales",
		"detail.no_local_servers":  "Aucun serveur MCP local configuré",
		"btn.edit":             "Modifier",
		"btn.delete":           "Supprimer",
		"btn.move":             "Déplacer",
		"btn.save":             "Enregistrer",
		"btn.cancel":           "Annuler",
		"btn.add_server":       "Ajouter Serveur",
		"dialog.add_server":            "Ajouter Serveur MCP",
		"dialog.add_server_to_project": "Ajouter Serveur MCP au Projet",
		"dialog.edit_server":           "Modifier Serveur MCP",
		"dialog.delete_server":         "Supprimer Serveur",
		"dialog.delete_confirm":        "Êtes-vous sûr de vouloir supprimer le serveur '%s'?",
		"dialog.move_server":           "Déplacer Serveur",
		"dialog.move_to_project":       "Sélectionnez le projet de destination:",
		"dialog.move_to_global":        "Voulez-vous déplacer le serveur '%s' vers la portée globale?",
		"dialog.no_projects":           "Aucun Projet",
		"dialog.no_projects_msg":       "Il n'y a pas de projets configurés pour déplacer le serveur.",
		"form.name":           "Nom",
		"form.name_hint":      "Nom du serveur",
		"form.scope":          "Portée",
		"form.scope_global":   "Global",
		"form.scope_project":  "Projet",
		"form.type":           "Type",
		"form.command":        "Commande",
		"form.command_hint":   "Commande (ex: uvx, npx)",
		"form.args":           "Arguments",
		"form.args_hint":      "Arguments séparés par espace (ex: -y mcp-server)",
		"form.url":            "URL",
		"form.url_hint":       "URL (pour http/sse)",
		"form.env":            "Variables d'Environnement",
		"form.env_hint":       "CLÉ=valeur, CLÉ2=valeur2 (séparés par virgule)",
	}

	// German
	t.strings[DE] = map[string]string{
		"toolbar.add_server": "Server hinzufügen",
		"toolbar.refresh":    "Aktualisieren",
		"tree.global":        "Global",
		"tree.projects":      "Projekte",
		"detail.select_item":   "Wählen Sie ein Element, um Details anzuzeigen",
		"detail.select_server": "Wählen Sie einen Server, um Details anzuzeigen",
		"detail.project":       "Projekt",
		"detail.server":        "Server",
		"detail.scope":         "Bereich",
		"detail.type":          "Typ",
		"detail.command":       "Befehl",
		"detail.args":          "Args",
		"detail.url":           "URL",
		"detail.env":           "Umgebungsvariablen",
		"detail.timeout":       "Zeitüberschreitung",
		"detail.path":          "Pfad",
		"detail.mcp_servers":       "MCP Server",
		"detail.configs":           "Konfigurationen",
		"detail.global_configs":    "Globale Konfigurationen",
		"detail.local_configs":     "Lokale Konfigurationen",
		"detail.no_local_servers":  "Keine lokalen MCP-Server konfiguriert",
		"btn.edit":             "Bearbeiten",
		"btn.delete":           "Löschen",
		"btn.move":             "Verschieben",
		"btn.save":             "Speichern",
		"btn.cancel":           "Abbrechen",
		"btn.add_server":       "Server hinzufügen",
		"dialog.add_server":            "MCP Server hinzufügen",
		"dialog.add_server_to_project": "MCP Server zum Projekt hinzufügen",
		"dialog.edit_server":           "MCP Server bearbeiten",
		"dialog.delete_server":         "Server löschen",
		"dialog.delete_confirm":        "Sind Sie sicher, dass Sie den Server '%s' löschen möchten?",
		"dialog.move_server":           "Server verschieben",
		"dialog.move_to_project":       "Wählen Sie das Zielprojekt:",
		"dialog.move_to_global":        "Möchten Sie den Server '%s' in den globalen Bereich verschieben?",
		"dialog.no_projects":           "Keine Projekte",
		"dialog.no_projects_msg":       "Es gibt keine konfigurierten Projekte, in die der Server verschoben werden kann.",
		"form.name":           "Name",
		"form.name_hint":      "Servername",
		"form.scope":          "Bereich",
		"form.scope_global":   "Global",
		"form.scope_project":  "Projekt",
		"form.type":           "Typ",
		"form.command":        "Befehl",
		"form.command_hint":   "Befehl (z.B.: uvx, npx)",
		"form.args":           "Argumente",
		"form.args_hint":      "Argumente durch Leerzeichen getrennt (z.B.: -y mcp-server)",
		"form.url":            "URL",
		"form.url_hint":       "URL (für http/sse)",
		"form.env":            "Umgebungsvariablen",
		"form.env_hint":       "SCHLÜSSEL=wert, SCHLÜSSEL2=wert2 (durch Komma getrennt)",
	}

	// Spanish
	t.strings[ES] = map[string]string{
		"toolbar.add_server": "Agregar Servidor",
		"toolbar.refresh":    "Actualizar",
		"tree.global":        "Global",
		"tree.projects":      "Proyectos",
		"detail.select_item":   "Selecciona un elemento para ver los detalles",
		"detail.select_server": "Selecciona un servidor para ver los detalles",
		"detail.project":       "Proyecto",
		"detail.server":        "Servidor",
		"detail.scope":         "Ámbito",
		"detail.type":          "Tipo",
		"detail.command":       "Comando",
		"detail.args":          "Args",
		"detail.url":           "URL",
		"detail.env":           "Variables de entorno",
		"detail.timeout":       "Tiempo de espera",
		"detail.path":          "Ruta",
		"detail.mcp_servers":       "Servidores MCP",
		"detail.configs":           "Configuraciones",
		"detail.global_configs":    "Configuraciones Globales",
		"detail.local_configs":     "Configuraciones Locales",
		"detail.no_local_servers":  "No hay servidores MCP locales configurados",
		"btn.edit":             "Editar",
		"btn.delete":           "Eliminar",
		"btn.move":             "Mover",
		"btn.save":             "Guardar",
		"btn.cancel":           "Cancelar",
		"btn.add_server":       "Agregar Servidor",
		"dialog.add_server":            "Agregar Servidor MCP",
		"dialog.add_server_to_project": "Agregar Servidor MCP al Proyecto",
		"dialog.edit_server":           "Editar Servidor MCP",
		"dialog.delete_server":         "Eliminar Servidor",
		"dialog.delete_confirm":        "¿Estás seguro de que deseas eliminar el servidor '%s'?",
		"dialog.move_server":           "Mover Servidor",
		"dialog.move_to_project":       "Selecciona el proyecto de destino:",
		"dialog.move_to_global":        "¿Deseas mover el servidor '%s' al ámbito global?",
		"dialog.no_projects":           "Sin Proyectos",
		"dialog.no_projects_msg":       "No hay proyectos configurados a los que mover el servidor.",
		"form.name":           "Nombre",
		"form.name_hint":      "Nombre del servidor",
		"form.scope":          "Ámbito",
		"form.scope_global":   "Global",
		"form.scope_project":  "Proyecto",
		"form.type":           "Tipo",
		"form.command":        "Comando",
		"form.command_hint":   "Comando (ej: uvx, npx)",
		"form.args":           "Argumentos",
		"form.args_hint":      "Argumentos separados por espacio (ej: -y mcp-server)",
		"form.url":            "URL",
		"form.url_hint":       "URL (para http/sse)",
		"form.env":            "Variables de Entorno",
		"form.env_hint":       "CLAVE=valor, CLAVE2=valor2 (separados por coma)",
	}

	// Portuguese
	t.strings[PT] = map[string]string{
		"toolbar.add_server": "Adicionar Servidor",
		"toolbar.refresh":    "Atualizar",
		"tree.global":        "Global",
		"tree.projects":      "Projetos",
		"detail.select_item":   "Selecione um item para ver os detalhes",
		"detail.select_server": "Selecione um servidor para ver os detalhes",
		"detail.project":       "Projeto",
		"detail.server":        "Servidor",
		"detail.scope":         "Escopo",
		"detail.type":          "Tipo",
		"detail.command":       "Comando",
		"detail.args":          "Args",
		"detail.url":           "URL",
		"detail.env":           "Variáveis de ambiente",
		"detail.timeout":       "Tempo limite",
		"detail.path":          "Caminho",
		"detail.mcp_servers":       "Servidores MCP",
		"detail.configs":           "Configurações",
		"detail.global_configs":    "Configurações Globais",
		"detail.local_configs":     "Configurações Locais",
		"detail.no_local_servers":  "Nenhum servidor MCP local configurado",
		"btn.edit":             "Editar",
		"btn.delete":           "Excluir",
		"btn.move":             "Mover",
		"btn.save":             "Salvar",
		"btn.cancel":           "Cancelar",
		"btn.add_server":       "Adicionar Servidor",
		"dialog.add_server":            "Adicionar Servidor MCP",
		"dialog.add_server_to_project": "Adicionar Servidor MCP ao Projeto",
		"dialog.edit_server":           "Editar Servidor MCP",
		"dialog.delete_server":         "Excluir Servidor",
		"dialog.delete_confirm":        "Tem certeza de que deseja excluir o servidor '%s'?",
		"dialog.move_server":           "Mover Servidor",
		"dialog.move_to_project":       "Selecione o projeto de destino:",
		"dialog.move_to_global":        "Deseja mover o servidor '%s' para o escopo global?",
		"dialog.no_projects":           "Sem Projetos",
		"dialog.no_projects_msg":       "Não há projetos configurados para mover o servidor.",
		"form.name":           "Nome",
		"form.name_hint":      "Nome do servidor",
		"form.scope":          "Escopo",
		"form.scope_global":   "Global",
		"form.scope_project":  "Projeto",
		"form.type":           "Tipo",
		"form.command":        "Comando",
		"form.command_hint":   "Comando (ex: uvx, npx)",
		"form.args":           "Argumentos",
		"form.args_hint":      "Argumentos separados por espaço (ex: -y mcp-server)",
		"form.url":            "URL",
		"form.url_hint":       "URL (para http/sse)",
		"form.env":            "Variáveis de Ambiente",
		"form.env_hint":       "CHAVE=valor, CHAVE2=valor2 (separados por vírgula)",
	}

	// Japanese
	t.strings[JA] = map[string]string{
		"toolbar.add_server": "サーバー追加",
		"toolbar.refresh":    "更新",
		"tree.global":        "グローバル",
		"tree.projects":      "プロジェクト",
		"detail.select_item":   "詳細を表示するアイテムを選択",
		"detail.select_server": "詳細を表示するサーバーを選択",
		"detail.project":       "プロジェクト",
		"detail.server":        "サーバー",
		"detail.scope":         "スコープ",
		"detail.type":          "タイプ",
		"detail.command":       "コマンド",
		"detail.args":          "引数",
		"detail.url":           "URL",
		"detail.env":           "環境変数",
		"detail.timeout":       "タイムアウト",
		"detail.path":          "パス",
		"detail.mcp_servers":       "MCPサーバー",
		"detail.configs":           "設定",
		"detail.global_configs":    "グローバル設定",
		"detail.local_configs":     "ローカル設定",
		"detail.no_local_servers":  "ローカルMCPサーバーが設定されていません",
		"btn.edit":             "編集",
		"btn.delete":           "削除",
		"btn.move":             "移動",
		"btn.save":             "保存",
		"btn.cancel":           "キャンセル",
		"btn.add_server":       "サーバー追加",
		"dialog.add_server":            "MCPサーバー追加",
		"dialog.add_server_to_project": "プロジェクトにMCPサーバー追加",
		"dialog.edit_server":           "MCPサーバー編集",
		"dialog.delete_server":         "サーバー削除",
		"dialog.delete_confirm":        "サーバー '%s' を削除しますか？",
		"dialog.move_server":           "サーバー移動",
		"dialog.move_to_project":       "移動先プロジェクトを選択:",
		"dialog.move_to_global":        "サーバー '%s' をグローバルスコープに移動しますか？",
		"dialog.no_projects":           "プロジェクトなし",
		"dialog.no_projects_msg":       "サーバーを移動できるプロジェクトがありません。",
		"form.name":           "名前",
		"form.name_hint":      "サーバー名",
		"form.scope":          "スコープ",
		"form.scope_global":   "グローバル",
		"form.scope_project":  "プロジェクト",
		"form.type":           "タイプ",
		"form.command":        "コマンド",
		"form.command_hint":   "コマンド (例: uvx, npx)",
		"form.args":           "引数",
		"form.args_hint":      "スペース区切りの引数 (例: -y mcp-server)",
		"form.url":            "URL",
		"form.url_hint":       "URL (http/sse用)",
		"form.env":            "環境変数",
		"form.env_hint":       "KEY=value, KEY2=value2 (カンマ区切り)",
	}

	// Korean
	t.strings[KO] = map[string]string{
		"toolbar.add_server": "서버 추가",
		"toolbar.refresh":    "새로고침",
		"tree.global":        "전역",
		"tree.projects":      "프로젝트",
		"detail.select_item":   "세부 정보를 보려면 항목을 선택하세요",
		"detail.select_server": "세부 정보를 보려면 서버를 선택하세요",
		"detail.project":       "프로젝트",
		"detail.server":        "서버",
		"detail.scope":         "범위",
		"detail.type":          "유형",
		"detail.command":       "명령어",
		"detail.args":          "인수",
		"detail.url":           "URL",
		"detail.env":           "환경 변수",
		"detail.timeout":       "시간 초과",
		"detail.path":          "경로",
		"detail.mcp_servers":       "MCP 서버",
		"detail.configs":           "구성",
		"detail.global_configs":    "전역 구성",
		"detail.local_configs":     "로컬 구성",
		"detail.no_local_servers":  "구성된 로컬 MCP 서버가 없습니다",
		"btn.edit":             "편집",
		"btn.delete":           "삭제",
		"btn.move":             "이동",
		"btn.save":             "저장",
		"btn.cancel":           "취소",
		"btn.add_server":       "서버 추가",
		"dialog.add_server":            "MCP 서버 추가",
		"dialog.add_server_to_project": "프로젝트에 MCP 서버 추가",
		"dialog.edit_server":           "MCP 서버 편집",
		"dialog.delete_server":         "서버 삭제",
		"dialog.delete_confirm":        "서버 '%s'을(를) 삭제하시겠습니까?",
		"dialog.move_server":           "서버 이동",
		"dialog.move_to_project":       "대상 프로젝트 선택:",
		"dialog.move_to_global":        "서버 '%s'을(를) 전역 범위로 이동하시겠습니까?",
		"dialog.no_projects":           "프로젝트 없음",
		"dialog.no_projects_msg":       "서버를 이동할 구성된 프로젝트가 없습니다.",
		"form.name":           "이름",
		"form.name_hint":      "서버 이름",
		"form.scope":          "범위",
		"form.scope_global":   "전역",
		"form.scope_project":  "프로젝트",
		"form.type":           "유형",
		"form.command":        "명령어",
		"form.command_hint":   "명령어 (예: uvx, npx)",
		"form.args":           "인수",
		"form.args_hint":      "공백으로 구분된 인수 (예: -y mcp-server)",
		"form.url":            "URL",
		"form.url_hint":       "URL (http/sse용)",
		"form.env":            "환경 변수",
		"form.env_hint":       "KEY=value, KEY2=value2 (쉼표로 구분)",
	}

	// Chinese (Simplified)
	t.strings[CN] = map[string]string{
		"toolbar.add_server": "添加服务器",
		"toolbar.refresh":    "刷新",
		"tree.global":        "全局",
		"tree.projects":      "项目",
		"detail.select_item":   "选择一个项目以查看详情",
		"detail.select_server": "选择一个服务器以查看详情",
		"detail.project":       "项目",
		"detail.server":        "服务器",
		"detail.scope":         "范围",
		"detail.type":          "类型",
		"detail.command":       "命令",
		"detail.args":          "参数",
		"detail.url":           "URL",
		"detail.env":           "环境变量",
		"detail.timeout":       "超时",
		"detail.path":          "路径",
		"detail.mcp_servers":       "MCP服务器",
		"detail.configs":           "配置",
		"detail.global_configs":    "全局配置",
		"detail.local_configs":     "本地配置",
		"detail.no_local_servers":  "未配置本地MCP服务器",
		"btn.edit":             "编辑",
		"btn.delete":           "删除",
		"btn.move":             "移动",
		"btn.save":             "保存",
		"btn.cancel":           "取消",
		"btn.add_server":       "添加服务器",
		"dialog.add_server":            "添加MCP服务器",
		"dialog.add_server_to_project": "向项目添加MCP服务器",
		"dialog.edit_server":           "编辑MCP服务器",
		"dialog.delete_server":         "删除服务器",
		"dialog.delete_confirm":        "确定要删除服务器 '%s' 吗？",
		"dialog.move_server":           "移动服务器",
		"dialog.move_to_project":       "选择目标项目：",
		"dialog.move_to_global":        "要将服务器 '%s' 移动到全局范围吗？",
		"dialog.no_projects":           "没有项目",
		"dialog.no_projects_msg":       "没有可以移动服务器的已配置项目。",
		"form.name":           "名称",
		"form.name_hint":      "服务器名称",
		"form.scope":          "范围",
		"form.scope_global":   "全局",
		"form.scope_project":  "项目",
		"form.type":           "类型",
		"form.command":        "命令",
		"form.command_hint":   "命令 (例: uvx, npx)",
		"form.args":           "参数",
		"form.args_hint":      "空格分隔的参数 (例: -y mcp-server)",
		"form.url":            "URL",
		"form.url_hint":       "URL (用于 http/sse)",
		"form.env":            "环境变量",
		"form.env_hint":       "KEY=value, KEY2=value2 (逗号分隔)",
	}

	// Ukrainian
	t.strings[UK] = map[string]string{
		"toolbar.add_server": "Додати сервер",
		"toolbar.refresh":    "Оновити",
		"tree.global":        "Глобальний",
		"tree.projects":      "Проекти",
		"detail.select_item":   "Виберіть елемент для перегляду деталей",
		"detail.select_server": "Виберіть сервер для перегляду деталей",
		"detail.project":       "Проект",
		"detail.server":        "Сервер",
		"detail.scope":         "Область",
		"detail.type":          "Тип",
		"detail.command":       "Команда",
		"detail.args":          "Аргументи",
		"detail.url":           "URL",
		"detail.env":           "Змінні середовища",
		"detail.timeout":       "Час очікування",
		"detail.path":          "Шлях",
		"detail.mcp_servers":       "MCP Сервери",
		"detail.configs":           "Конфігурації",
		"detail.global_configs":    "Глобальні конфігурації",
		"detail.local_configs":     "Локальні конфігурації",
		"detail.no_local_servers":  "Локальні MCP сервери не налаштовані",
		"btn.edit":             "Редагувати",
		"btn.delete":           "Видалити",
		"btn.move":             "Перемістити",
		"btn.save":             "Зберегти",
		"btn.cancel":           "Скасувати",
		"btn.add_server":       "Додати сервер",
		"dialog.add_server":            "Додати MCP сервер",
		"dialog.add_server_to_project": "Додати MCP сервер до проекту",
		"dialog.edit_server":           "Редагувати MCP сервер",
		"dialog.delete_server":         "Видалити сервер",
		"dialog.delete_confirm":        "Ви впевнені, що хочете видалити сервер '%s'?",
		"dialog.move_server":           "Перемістити сервер",
		"dialog.move_to_project":       "Виберіть проект призначення:",
		"dialog.move_to_global":        "Перемістити сервер '%s' до глобальної області?",
		"dialog.no_projects":           "Немає проектів",
		"dialog.no_projects_msg":       "Немає налаштованих проектів для переміщення сервера.",
		"form.name":           "Назва",
		"form.name_hint":      "Назва сервера",
		"form.scope":          "Область",
		"form.scope_global":   "Глобальний",
		"form.scope_project":  "Проект",
		"form.type":           "Тип",
		"form.command":        "Команда",
		"form.command_hint":   "Команда (напр.: uvx, npx)",
		"form.args":           "Аргументи",
		"form.args_hint":      "Аргументи розділені пробілом (напр.: -y mcp-server)",
		"form.url":            "URL",
		"form.url_hint":       "URL (для http/sse)",
		"form.env":            "Змінні середовища",
		"form.env_hint":       "КЛЮЧ=значення, КЛЮЧ2=значення2 (розділені комою)",
	}
}
