package tui

import (
	"fmt"
	// Caminho de importação corrigido aqui
	tea "github.com/charmbracelet/bubbletea"
)

// Start é a função pública que inicia toda a aplicação TUI.
func Start() error {
	// Usamos newModel() que é privado ao pacote
	p := tea.NewProgram(newModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("erro ao executar o programa TUI: %w", err)
	}
	return nil
}
