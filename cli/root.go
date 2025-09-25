package cli

import (
	"fmt"
	"os"

	"github.com/missarcadia/DASHBOARD-GO/tui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Um dashboard Go avançado para interagir com o Git.",
	Long: `DASHBOARD é um cliente Git completo e uma ferramenta de visualização
construída em Go. Execute-o sem argumentos para iniciar o dashboard interativo
ou use um dos muitos subcomandos do Git.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Se nenhum subcomando for chamado, inicie o dashboard TUI
		// A chamada foi corrigida de tui.StartDashboard para tui.Start
		if err := tui.Start(); err != nil {
			fmt.Println("Erro ao iniciar o dashboard:", err)
			os.Exit(1)
		}
	},
}

// Execute adiciona todos os comandos filhos ao comando raiz e define as flags apropriadamente.
func Execute() {
	// Adiciona todos os comandos Git como subcomandos
	addGitCommands(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
