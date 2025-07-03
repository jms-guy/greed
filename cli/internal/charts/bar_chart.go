package charts

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/NimbleMarkets/ntcharts/barchart"
	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jms-guy/greed/models"
	zone "github.com/lrstanley/bubblezone"
)

var defaultStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63")) // purple

var axisStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("3")) // yellow

var labelStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("63")) // purple

var blockStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("9")). // red
	Background(lipgloss.Color("9"))  // red

var blockStyle2 = lipgloss.NewStyle().
	Foreground(lipgloss.Color("2")). // green
	Background(lipgloss.Color("2"))  // red

type model struct {
	b1 barchart.Model
	lv []barchart.BarData
	zM *zone.Manager
}

func title(m *barchart.Model) string {
	return fmt.Sprintf("Max:%.1f, AutoMax:%t\nBarGap:%d, ShowAxis:%t\n", m.MaxValue(), m.AutoMaxValue, m.BarGap(), m.ShowAxis())
}

func legend(bd barchart.BarData) (r string) {
	r = "Legend\n"
	for _, bv := range bd.Values {
		r += "\n" + bv.Style.Render(fmt.Sprintf("%c %s", runes.FullBlock, bv.Name))
	}
	return
}

func (m model) Init() tea.Cmd {
	m.b1.Draw()
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	case tea.MouseMsg:
		if msg.Action == tea.MouseActionPress {
			switch {
			case m.zM.Get(m.b1.ZoneID()).InBounds(msg):
				return m, nil
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	s := "Same data values are pushed to all vertical bar charts, `q/ctrl+c` to quit\n"
	s += "Click bar segment to select and display values\n"
	s += lipgloss.JoinHorizontal(lipgloss.Top,
		defaultStyle.Render(title(&m.b1)+m.b1.View()),
		lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.JoinHorizontal(lipgloss.Top,
				defaultStyle.Render(legend(m.lv[0])),
			),
		),
	)
	return m.zM.Scan(s) // call zone Manager.Scan() at root model
}

func MakeIncomeChart(data []models.MonetaryData) error {

	var monthlyData []barchart.BarData
	width := 275
	height := 50
	
	for _, month := range data {
		i := strings.TrimPrefix(month.Income, "-")
		e := strings.TrimPrefix(month.Expenses, "-")
		income, err := strconv.ParseFloat(i, 64)
		if err != nil {
			return fmt.Errorf("error converting string value: %w", err)
		}
		expenses, err := strconv.ParseFloat(e, 64)
		if err != nil {
			return fmt.Errorf("error converting string value: %w", err)
		}

		m := barchart.BarData{
			Label: month.Date,
			Values: []barchart.BarValue{
				{Name: "Income", Value: income, Style: blockStyle},
				{Name: "Expenses", Value: expenses, Style: blockStyle2},
			},
		}

		monthlyData = append(monthlyData, m)
	}

	zoneManager := zone.New()

	m := model{
		barchart.New(width, height,
			barchart.WithZoneManager(zoneManager),
			barchart.WithDataSet(monthlyData),
			barchart.WithStyles(axisStyle, labelStyle)),
			monthlyData,
			zoneManager,
	}

	if _, err := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
		return fmt.Errorf("error creating graph: %w", err)
	}

	return nil
}