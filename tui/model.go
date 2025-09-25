package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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
		menuItem{title: "Ver Status", desc: "Executa 'git status' para ver as alterações"},
		menuItem{title: "Ver Log Recente", desc: "Mostra os 10 últimos commits"},
		menuItem{title: "Fazer Commit", desc: "Inicia o processo para commitar as alterações"},
		menuItem{title: "Pull (Rebase)", desc: "Executa 'git pull --rebase'"},
		menuItem{title: "Push", desc: "Executa 'git push' para o remote atual"},
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
			m.output = msg.err.Error()
		}
		m.viewport.SetContent(m.output)
		return m, nil

	case tea.KeyMsg:
		// ### INÍCIO DA CORREÇÃO ###
		// Lógica para voltar ao menu a partir de telas secundárias
		if msg.Type == tea.KeyEsc {
			// Verificamos se estamos em uma tela secundária
			if m.state == commandOutputView || m.state == commitInputView {
				// Apenas resetamos a textarea se estávamos na tela de commit
				if m.state == commitInputView {
					m.textarea.Reset()
				}
				m.state = menuView
				return m, nil
			}
		}
		// ### FIM DA CORREÇÃO ###
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

// handleMenuUpdate lida com a lógica da tela do menu principal
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
				case "Ver Status":
					return runGitCommand("status")
				case "Ver Log Recente":
					return runGitCommand("log", "-n", "10", "--graph", "--pretty=format:'%Cred%h%Creset -%C(yellow)%d%Creset %s %Cgreen(%cr) %C(bold blue)<%an>%Creset'")
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
