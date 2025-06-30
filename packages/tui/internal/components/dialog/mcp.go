package dialog

import (
	"fmt"

	"github.com/charmbracelet/bubbles/v2/key"
	tea "github.com/charmbracelet/bubbletea/v2"
	"github.com/sst/opencode/internal/app"
	"github.com/sst/opencode/internal/components/list"
	"github.com/sst/opencode/internal/components/modal"
	"github.com/sst/opencode/internal/layout"
	"github.com/sst/opencode/internal/styles"
	"github.com/sst/opencode/internal/theme"
	"github.com/sst/opencode/internal/util"
)

const (
	maxMcpDialogWidth    = 60
	numVisibleMcpServers = 8
)

type mcpServerItem struct {
	name    string
	enabled bool
	mcpType string
	details string
}

func (i mcpServerItem) Render(selected bool, width int) string {
	t := theme.CurrentTheme()
	baseStyle := styles.NewStyle()

	status := "✗ disabled"
	if i.enabled {
		status = "✓ enabled"
	}

	title := fmt.Sprintf("%s (%s)", i.name, status)
	description := fmt.Sprintf("%s - %s", i.mcpType, i.details)

	// Truncate if too long
	if len(title) > width-2 {
		title = title[:width-5] + "..."
	}
	if len(description) > width-2 {
		description = description[:width-5] + "..."
	}

	var itemStyle styles.Style
	if selected {
		itemStyle = baseStyle.
			Background(t.Primary()).
			Foreground(t.BackgroundElement()).
			Width(width).
			PaddingLeft(1)
	} else {
		itemStyle = baseStyle.
			Foreground(t.Text()).
			PaddingLeft(1)
	}

	content := fmt.Sprintf("%s\n  %s", title, description)
	return itemStyle.Render(content)
}

type McpServerToggleMsg struct {
	ServerName string
	Enabled    bool
}

type mcpDialog struct {
	width      int
	height     int
	modal      *modal.Modal
	app        *app.App
	mcpList    list.List[mcpServerItem]
	keymap     mcpKeymap
	mcpServers []mcpServerItem
}

type mcpKeymap struct {
	toggle key.Binding
	close  key.Binding
}

func newMcpKeymap() mcpKeymap {
	return mcpKeymap{
		toggle: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "toggle server"),
		),
		close: key.NewBinding(
			key.WithKeys("esc", "ctrl+c"),
			key.WithHelp("esc", "close"),
		),
	}
}

func (m *mcpDialog) Init() tea.Cmd {
	return nil
}

func (m *mcpDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keymap.toggle):
			selectedItem, idx := m.mcpList.GetSelectedItem()
			if idx >= 0 {
				// Toggle the server state
				selectedItem.enabled = !selectedItem.enabled

				// Update the list item
				m.mcpServers[idx] = selectedItem
				m.mcpList.SetItems(m.mcpServers)

				// Restore the selected index to prevent jumping to first item
				m.mcpList.SetSelectedIndex(idx)

				// Send toggle message to update backend
				cmds = append(cmds, func() tea.Msg {
					return McpServerToggleMsg{
						ServerName: selectedItem.name,
						Enabled:    selectedItem.enabled,
					}
				})
			}
		case key.Matches(msg, m.keymap.close):
			return m, util.CmdHandler(modal.CloseModalMsg{})
		}
	}

	// Update list
	updatedList, listCmd := m.mcpList.Update(msg)
	m.mcpList = updatedList.(list.List[mcpServerItem])
	cmds = append(cmds, listCmd)

	return m, tea.Batch(cmds...)
}

func (m *mcpDialog) View() string {
	if m.mcpList.IsEmpty() {
		t := theme.CurrentTheme()
		emptyStyle := styles.NewStyle().
			Foreground(t.TextMuted()).
			Width(maxMcpDialogWidth)
		return emptyStyle.Render("No MCP servers configured")
	}
	return m.mcpList.View()
}

func (m *mcpDialog) Render(background string) string {
	return m.modal.Render(m.View(), background)
}

func (m *mcpDialog) Close() tea.Cmd {
	return nil
}

type McpDialog interface {
	layout.Modal
}

func NewMcpDialog(app *app.App) McpDialog {
	// Create list items from MCP config
	var mcpServers []mcpServerItem

	// For now, create some dummy MCP servers as example
	// This will be replaced when we have proper MCP config integration
	mcpServers = append(mcpServers, mcpServerItem{
		name:    "example-local",
		enabled: true,
		mcpType: "local",
		details: "node server.js",
	})
	mcpServers = append(mcpServers, mcpServerItem{
		name:    "example-remote",
		enabled: false,
		mcpType: "remote",
		details: "https://api.example.com/mcp",
	})

	mcpList := list.NewListComponent(mcpServers, numVisibleMcpServers, "No MCP servers available", true)
	mcpList.SetMaxWidth(maxMcpDialogWidth)

	return &mcpDialog{
		app:        app,
		mcpList:    mcpList,
		mcpServers: mcpServers,
		modal: modal.New(
			modal.WithTitle("MCP Server Control"),
			modal.WithMaxWidth(maxMcpDialogWidth+4),
		),
		keymap: newMcpKeymap(),
	}
}
