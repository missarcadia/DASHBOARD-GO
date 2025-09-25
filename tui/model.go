package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	// Caminho de importação corrigido aqui
	"github.com/charmbracelet/lipgloss"
	"github.com/missarcadia/DASHBOARD-GO/git"
)

// menuItem implementa a interface list.Item para ser usada no nosso menu
type menuItem struct {
	title, desc string
}

func (i menuItem) Title() string       { return i.title }
func (i menuItem) Description() string { return i.desc }
func (i menuItem) FilterValue() string { return i.title }

// Definimos os possíveis estados (telas) da nossa aplicação
type viewState int

const (
	menuView viewState = iota
	commandOutputView
	commitInputView
	loadingView
)

// Mensagem para indicar que um comando terminou de ser executado
type commandFinishedMsg struct {
	output string
	err    error
}

// Model principal da aplicação
type model struct {
	state         viewState
	menu          list.Model
	spinner       spinner.Model
	viewport      viewport.Model
	textarea      textarea.Model
	output        string
	width, height int
}

// newModel cria o modelo inicial da nossa TUI.
func newModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	items := []list.Item{
		menuItem{title: "Ver Status Detalhado", desc: "Mostra um status detalhado das alterações"},
		menuItem{title: "Ver Info do Usuário", desc: "Exibe seu nome, email e conta do GitHub"},
		menuItem{title: "Fazer Commit", desc: "Inicia o processo para commitar as alterações"},
		menuItem{title: "Pull (Rebase)", desc: "Executa 'git pull --rebase'"},
		menuItem{title: "Push", desc: "Executa 'git push' para o remote atual"},
		menuItem{title: "Criar Repositório no GitHub", desc: "Cria um repo no GitHub e conecta o 'origin' (requer 'gh' CLI)"},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = selectedItemStyle
	delegate.Styles.SelectedDesc = selectedItemStyle.Copy().Faint(true)
	delegate.Styles.NormalTitle = itemStyle
	delegate.Styles.NormalDesc = itemStyle.Copy().Faint(true)

	menuList := list.New(items, delegate, 0, 0)
	menuList.Title = "O que você gostaria de fazer?"
	menuList.SetShowHelp(false)

	return model{
		state:    menuView,
		menu:     menuList,
		spinner:  s,
		viewport: viewport.New(0, 0),
	}
}

func (m model) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update é a função principal que lida com todas as mensagens e eventos.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.menu.SetSize(msg.Width-2, msg.Height-4)
		m.viewport.Width = msg.Width - appStyle.GetHorizontalFrameSize()
		m.viewport.Height = msg.Height - appStyle.GetVerticalFrameSize() - 5
		return m, nil

	case commandFinishedMsg:
		m.state = commandOutputView
		m.output = msg.output
		if msg.err != nil {
			m.output = msg.output // Git já coloca o erro na saída
		}
		m.viewport.SetContent(m.output)
		m.viewport.GotoTop()
		return m, nil

	case tea.KeyMsg:
		if msg.Type == tea.KeyEsc {
			if m.state == commandOutputView || m.state == commitInputView {
				if m.state == commitInputView {
					m.textarea.Reset()
				}
				m.state = menuView
				return m, nil
			}
		}
	}

	switch m.state {
	case menuView:
		cmd = m.handleMenuUpdate(msg)
	case commandOutputView:
		m.viewport, cmd = m.viewport.Update(msg)
	case commitInputView:
		cmd = m.handleCommitInputUpdate(msg)
	case loadingView:
		m.spinner, cmd = m.spinner.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renderiza a UI baseada no estado atual.
func (m model) View() string {
	var s strings.Builder
	s.WriteString(titleStyle.Render("📊 DASHBOARD GIT INTERATIVO") + "\n\n")

	switch m.state {
	case menuView:
		s.WriteString(m.menu.View())
	case loadingView:
		s.WriteString(m.spinner.View() + " Executando comando...")
	case commandOutputView:
		s.WriteString(viewportStyle.Render(m.viewport.View()))
		s.WriteString(helpStyle.Render("↑/↓ para rolar | Esc para voltar ao menu"))
	case commitInputView:
		s.WriteString("Digite a mensagem do commit (Ctrl+D para finalizar, Esc para cancelar):\n")
		s.WriteString(m.textarea.View())
	}

	return appStyle.Render(s.String())
}

// handleMenuUpdate agora inclui a lógica para as novas opções
func (m *model) handleMenuUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return tea.Quit
		case "enter":
			item, ok := m.menu.SelectedItem().(menuItem)
			if ok {
				m.state = loadingView
				switch item.title {
				case "Ver Status Detalhado":
					return func() tea.Msg {
						output, err := git.GetDetailedStatus()
						return commandFinishedMsg{output: output, err: err}
					}
				case "Ver Info do Usuário":
					return func() tea.Msg {
						output, err := git.GetGitUserInfo()
						return commandFinishedMsg{output: output, err: err}
					}
				case "Pull (Rebase)":
					return runGitCommand("pull", "--rebase")
				case "Push":
					return runGitCommand("push")
				case "Fazer Commit":
					m.state = commitInputView
					m.textarea = textarea.New()
					m.textarea.Placeholder = "Adicione as alterações com 'git add .' antes de commitar..."
					m.textarea.Focus()
					return textarea.Blink
				case "Criar Repositório no GitHub":
					return runGitCommand("gh", "repo", "create", "DASHBOARD-GO", "--public", "--source=.", "--remote=origin")
				}
			}
		}
	}
	m.menu, cmd = m.menu.Update(msg)
	return cmd
}

// handleCommitInputUpdate lida com a lógica da tela de input de commit
func (m *model) handleCommitInputUpdate(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlD:
			m.state = loadingView
			_, err := git.RunGitCommand("add", ".")
			if err != nil {
				return func() tea.Msg {
					return commandFinishedMsg{err: fmt.Errorf("erro ao executar 'git add .': %w", err)}
				}
			}
			return runGitCommand("commit", "-m", m.textarea.Value())
		}
	}
	m.textarea, cmd = m.textarea.Update(msg)
	return cmd
}

// runGitCommand é uma função helper que cria um tea.Cmd para executar um comando git
func runGitCommand(args ...string) tea.Cmd {
	return func() tea.Msg {
		output, err := git.RunGitCommand(args...)
		return commandFinishedMsg{output: output, err: err}
	}
}
