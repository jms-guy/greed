package charts

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/NimbleMarkets/ntcharts/barchart"
	"github.com/NimbleMarkets/ntcharts/canvas/runes"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jms-guy/greed/models"
	zone "github.com/lrstanley/bubblezone"
	"golang.org/x/term"
)

var defaultStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("63")) // purple

var axisStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("3")) // yellow

var labelStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("63")) // purple

var blockStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#2b3259")). 
	Background(lipgloss.Color("#2b3259"))  

var blockStyle2 = lipgloss.NewStyle().
	Foreground(lipgloss.Color("#3d0b18")). 
	Background(lipgloss.Color("#3d0b18"))  

type model struct {
	b1 barchart.Model
	lv []barchart.BarData
	zM *zone.Manager
	originalValues []struct {
        Income   float64
        Expenses float64
    }
}

func title(m *barchart.Model) string {
	return fmt.Sprintf("Max:%.1f\nBarGap:%d\n", m.MaxValue(), m.BarGap())
}

func legend(bd barchart.BarData) (r string) {
    r = "Legend\n"
    
    // Add standard entries
    r += "\n" + blockStyle.Render(string(runes.FullBlock)) + " Income"
    r += "\n" + blockStyle2.Render(string(runes.FullBlock)) + " Expenses"
    
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
    
    // Get the chart view first
    chartView := m.b1.View()
    
    // Create value labels for each bar
    var incomeLabels []string
    var expenseLabels []string
    
    // Estimate the axis area width
    axisWidth := 0
    for i, line := range strings.Split(chartView, "\n") {
        if i == m.b1.Canvas.Height()-2 { // At the origin Y position
            // Find the first non-space character after any axis labels
            for j, c := range line {
                if c != ' ' && j > axisWidth {
                    axisWidth = j
                    break
                }
            }
            break
        }
    }
    
    // Calculate total width available for bars
    barsAreaWidth := m.b1.Width() - axisWidth
    
    // Calculate individual bar width
    totalBars := len(m.lv)
    totalGaps := (totalBars - 1) * m.b1.BarGap()
    var singleBarWidth int
	if len(m.lv) > 0 {
		singleBarWidth = (barsAreaWidth - totalGaps) / len(m.lv)
	} else {
		singleBarWidth = 0 // or some default value
	}
    
    // Create value labels
	for i, _ := range m.lv {
		if i >= len(m.originalValues) {
			continue // Skip if index out of bounds
		}
		
		// Get original values from the parallel slice
		incomeValue := m.originalValues[i].Income
		expenseValue := m.originalValues[i].Expenses
        
        // Format the values - now horizontally
        incomeText := fmt.Sprintf("I:$%.2f", incomeValue)
        expenseText := fmt.Sprintf("E:$%.2f", expenseValue)
        
        // Center align the text within the bar width
        paddedIncome := lipgloss.NewStyle().
            Width(singleBarWidth).
            Align(lipgloss.Center).
            Render(incomeText)
            
        paddedExpense := lipgloss.NewStyle().
            Width(singleBarWidth).
            Align(lipgloss.Center).
            Render(expenseText)
            
        incomeLabels = append(incomeLabels, paddedIncome)
        expenseLabels = append(expenseLabels, paddedExpense)
    }
    
    // Join all value labels horizontally with appropriate gaps
    var incomeRow, expenseRow string
    incomeRow = strings.Repeat(" ", axisWidth) // Add padding for axis area
    expenseRow = strings.Repeat(" ", axisWidth)
    
    for i, _ := range incomeLabels {
        incomeRow += incomeLabels[i]
        expenseRow += expenseLabels[i]
        
        // Add gap between labels (except after the last one)
        if i < len(incomeLabels)-1 && m.b1.BarGap() > 0 {
            incomeRow += strings.Repeat(" ", m.b1.BarGap())
            expenseRow += strings.Repeat(" ", m.b1.BarGap())
        }
    }
    
    // Join the chart view with the values rows
    chartWithValues := chartView + "\n" + incomeRow + "\n" + expenseRow
    
    s += lipgloss.JoinHorizontal(lipgloss.Top,
        defaultStyle.Render(title(&m.b1) + chartWithValues),
        lipgloss.JoinVertical(lipgloss.Left,
            lipgloss.JoinHorizontal(lipgloss.Top,
                defaultStyle.Render(legend(m.lv[0])),
            ),
        ),
    )
    return m.zM.Scan(s)
}

func MakeIncomeChart(data []models.MonetaryData) error {
    if len(data) == 0 {
        fmt.Println("No data available to display")
        return nil
    }

    // Get terminal dimensions
    width, height, err := term.GetSize(int(os.Stdout.Fd()))
    if err != nil {
        // Fallback to reasonable defaults if we can't get terminal size
        width = 80
        height = 24
    }

    // Calculate chart dimensions based on terminal size
    // Leave some margin for borders and other UI elements
    chartWidth := width - 20  // Leave 20 columns for margins and other elements
    chartHeight := height / 2 // Use half the terminal height for the chart

    // Ensure minimum dimensions
    if chartWidth < 40 {
        chartWidth = 40 // Minimum width to display a readable chart
    }
    if chartHeight < 10 {
        chartHeight = 10 // Minimum height for a useful chart
    }

    // Create your bar chart with the dynamic dimensions
    var monthlyData []barchart.BarData
    var originalValues []struct {
        Income   float64
        Expenses float64
    }
    
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

        // Store original values in the parallel slice
        originalValues = append(originalValues, struct {
            Income   float64
            Expenses float64
        }{
            Income:   income,
            Expenses: expenses,
        })

        // Create stacked bar representation
        var barValues []barchart.BarValue
        
        if income >= expenses {
            // Income is higher or equal
            if expenses > 0 {
                barValues = append(barValues, barchart.BarValue{
                    Name:  "Expenses",
                    Value: expenses,
                    Style: blockStyle2,
                })
            }
            
            if income > expenses {
                barValues = append(barValues, barchart.BarValue{
                    Name:  "Income Surplus",
                    Value: income - expenses,
                    Style: blockStyle,
                })
            }
        } else {
            // Expenses are higher
            if income > 0 {
                barValues = append(barValues, barchart.BarValue{
                    Name:  "Income",
                    Value: income,
                    Style: blockStyle,
                })
            }
            
            barValues = append(barValues, barchart.BarValue{
                Name:  "Expense Surplus",
                Value: expenses - income,
                Style: blockStyle2,
            })
        }

        m := barchart.BarData{
            Label:  fmt.Sprintf("  %s", month.Date),
            Values: barValues,
        }

        monthlyData = append(monthlyData, m)
    }

    zoneManager := zone.New()

    m := model{
        barchart.New(chartWidth, chartHeight, // Use dynamic dimensions
            barchart.WithZoneManager(zoneManager),
            barchart.WithDataSet(monthlyData),
            barchart.WithStyles(axisStyle, labelStyle)),
        monthlyData,
        zoneManager,
        originalValues,
    }

    if _, err := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()).Run(); err != nil {
        return fmt.Errorf("error creating graph: %w", err)
    }

    return nil
}