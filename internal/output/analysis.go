package output

import (
	"sort"

	"github.com/rpgo/retirement-calculator/internal/domain"
	"github.com/shopspring/decimal"
)

// Recommendation encapsulates the selection result of the best scenario.
type Recommendation struct {
	ScenarioName       string
	FirstRetirementNet decimal.Decimal
	NetIncomeChange    decimal.Decimal
	PercentageChange   decimal.Decimal
}

// AnalyzeScenarios determines the best retirement scenario based on long-term net income.
// Compares scenarios at year 2030 (or earliest year when both spouses are fully retired in all scenarios).
// This provides an apples-to-apples comparison avoiding distortions from partial retirement years.
func AnalyzeScenarios(results *domain.ScenarioComparison) Recommendation {
	baseline := results.BaselineNetIncome
	type ranked struct {
		name   string
		income decimal.Decimal
	}
	var ranks []ranked

	// Use 2030 net income for comparison (typically 5+ years into retirement for most scenarios)
	// This avoids comparing partial retirement years and provides meaningful long-term comparison
	for _, sc := range results.Scenarios {
		// Prefer the 2030 calendar year comparison if available
		income := sc.NetIncome2030
		if income.IsZero() {
			// Fallback: use first year where both are fully retired
			for _, y := range sc.Projection {
				if y.IsRetired && !y.PersonADeceased && !y.PersonBDeceased {
					income = y.NetIncome
					break
				}
			}
		}
		ranks = append(ranks, ranked{sc.Name, income})
	}

	if len(ranks) == 0 {
		return Recommendation{}
	}

	sort.Slice(ranks, func(i, j int) bool { return ranks[i].income.GreaterThan(ranks[j].income) })
	best := ranks[0]

	// Calculate change vs baseline (current working income)
	delta := best.income.Sub(baseline)
	pct := decimal.Zero
	if !baseline.IsZero() {
		pct = delta.Div(baseline).Mul(decimal.NewFromInt(100))
	}

	return Recommendation{
		ScenarioName:       best.name,
		FirstRetirementNet: best.income,
		NetIncomeChange:    delta,
		PercentageChange:   pct,
	}
}
