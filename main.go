package main

import (
	"os/exec"
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	bm "github.com/charmbracelet/wish/bubbletea"
)
// ─── Styles ──────────────────────────────────────────────────────────────────

var (
	accent   = lipgloss.Color("#f0a029")
	muted    = lipgloss.Color("#26b2ca")
	white    = lipgloss.Color("#F9FAFB")
	green    = lipgloss.Color("#10B981")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent).
			PaddingBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(muted).
			Italic(true)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(white).
			Background(accent).
			PaddingLeft(2).
			PaddingRight(2)

	normalStyle = lipgloss.NewStyle().
			Foreground(white).
			PaddingLeft(2).
			PaddingRight(2)

	dimStyle = lipgloss.NewStyle().
			Foreground(muted)

	contentStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(accent).
			Padding(1, 2).
			MarginTop(1)

	tagStyle = lipgloss.NewStyle().
			Foreground(green).
			Bold(true)

	linkStyle = lipgloss.NewStyle().
			Foreground(accent).
			Underline(true)
)

// ─── Data ─────────────────────────────────────────────────────────────────────

type Project struct {
	Name        string
	Description string
	Tags        []string
	URL         string
}

type Page int

const (
	MenuPage Page = iota
	AboutPage
	ProjectsPage
	ContactPage
)

var menuItems = []string{"About", "Projects", "Contact", "Quit"}

var projects = []Project{
	{
		Name:        "Terminal Portfolio",
		Description: "This website! An SSH-accessible TUI portfolio built with Go, Wish, and Bubble Tea.",
		Tags:        []string{"Go", "TUI", "SSH"},
		URL:         "github.com/martinfitzgerald1-beep/tui-portfolio",
	},
	{
		Name:        "ITH correction",
		Description: `The project proposes a digital self-tuning method
to automatically adjust the inductor current threshold (ITH)
in a peak current mode controlled converter.
This means the maximum current limit is well defined
and performance is more robust.`,
		Tags:        []string{"RTL", "Verilog", "Virtuoso"},
		URL:         "github.com/martinfitzgerald1-beep/ith_correction",
	},
	{
		Name:        "Chess",
		Description: `A Java chess game playable directly in the terminal.`,
		Tags:        []string{"Java"},
		URL:         "github.com/martinfitzgerald1-beep/chess",
	},
}	

// ─── Model ────────────────────────────────────────────────────────────────────

type model struct {
	page       Page
	cursor     int
	projCursor int
	width      int
	height     int
}

func initialModel() model {
	return model{page: MenuPage, cursor: 0, projCursor: 0}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			if m.page == MenuPage {
				return m, tea.Quit
			}
			m.page = MenuPage
			m.cursor = 0

		case "esc":
			if m.page != MenuPage {
				m.page = MenuPage
				m.cursor = 0
			}

		case "up", "k":
			if m.page == MenuPage && m.cursor > 0 {
				m.cursor--
			} else if m.page == ProjectsPage && m.projCursor > 0 {
				m.projCursor--
			}

		case "down", "j":
			if m.page == MenuPage && m.cursor < len(menuItems)-1 {
				m.cursor++
			} else if m.page == ProjectsPage && m.projCursor < len(projects)-1 {
				m.projCursor++
			}

		case "enter", " ":
					if m.page == MenuPage {
						switch m.cursor {
						case 0:
							m.page = AboutPage
						case 1:
							m.page = ProjectsPage
						case 2:
							m.page = ContactPage
						case 3:
							return m, tea.Quit
						}
					}
		}
	}

	return m, nil
}

// ─── Views ────────────────────────────────────────────────────────────────────

func (m model) View() string {
	switch m.page {
	case AboutPage:
		return m.aboutView()
	case ProjectsPage:
		return m.projectsView()
	case ContactPage:
		return m.contactView()
	default:
		return m.menuView()
	}
}

func (m model) menuView() string {
	header := titleStyle.Render("Martin Fitzgerald") + "\n" +
		subtitleStyle.Render("Engineer · Builder · Human")

	menu := "\n"
	for i, item := range menuItems {
		if i == m.cursor {
			menu += selectedStyle.Render("▸ "+item) + "\n"
		} else {
			menu += normalStyle.Render("  "+item) + "\n"
		}
	}

	footer := "\n" + dimStyle.Render("↑↓ navigate · enter select · q quit")

	return lipgloss.JoinVertical(lipgloss.Left,
		"\n"+header,
		menu,
		footer,
	)
}

func (m model) aboutView() string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("About Me"),
		"",
		"Hi! I'm an electronic engineer based in Dublin.",
		"I love building tools, tinkering with systems,",
		"and apparently making people navigate my portfolio",
		"through a terminal.",
		"",
		"I am a 4th year electronic engineering student in UCD.",
		"I have an interest in mixed-signal engineering and software.",
		dimStyle.Render("esc to go back"),
	)

	return "\n" + contentStyle.Render(content)
}

func (m model) projectsView() string {
	title := titleStyle.Render("Projects")

	list := ""
	for i, p := range projects {
		if i == m.projCursor {
			list += selectedStyle.Render("▸ "+p.Name) + "\n"
		} else {
			list += normalStyle.Render("  "+p.Name) + "\n"
		}
	}

	proj := projects[m.projCursor]
	tags := ""
	for _, t := range proj.Tags {
		tags += tagStyle.Render("["+t+"] ")
	}

	detail := contentStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render(proj.Name),
			"",
			proj.Description,
			"",
			tags,
			"",
			linkStyle.Render(proj.URL),
		),
	)

	footer := dimStyle.Render("↑↓ select · esc back")

	return "\n" + lipgloss.JoinVertical(lipgloss.Left,
		title,
		list,
		detail,
		footer,
	)
}

func (m model) contactView() string {
	content := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render("Contact"),
		"",
		"The best way to reach me:",
		"",
		"  Email   "+linkStyle.Render("martin.fitzgerald1@ucd.ie"),
		"  GitHub  "+linkStyle.Render("github.com/martinfitzgerald1-beep"),
		"",
		dimStyle.Render("esc to go back"),
	)

	return "\n" + contentStyle.Render(content)
}

// ─── SSH Server ───────────────────────────────────────────────────────────────

func launchChess() tea.Cmd {
    cmd := exec.Command("java", "-jar", "chess_game.jar") // no space in filename
    return tea.ExecProcess(cmd, func(err error) tea.Msg {
        return nil
    })
}

func main() {
	srv, err := wish.NewServer(
		wish.WithAddress(":2222"), // change to :443 in production (requires root)
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bm.Middleware(func(s ssh.Session) (tea.Model, []tea.ProgramOption) {
				return initialModel(), []tea.ProgramOption{tea.WithAltScreen()}
			}),
		),
	)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Println("SSH server listening on :2222")
	log.Println("Connect with: ssh localhost -p 2222")

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("server error: %v", err)
		}
	}()

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
