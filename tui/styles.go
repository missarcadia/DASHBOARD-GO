package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Estilo base para a janela principal
	appStyle = lipgloss.NewStyle().Margin(1, 2)

	// Estilo para o título
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#007BFF")). // Azul
			Padding(0, 1).
			Bold(true)

	// Estilo para o item focado na lista
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))

	// Estilo para o texto de ajuda
	helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Padding(1, 0)

	// Estilo para a viewport que mostrará a saída dos comandos
	viewportStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)
)
