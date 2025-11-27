package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rpgo/retirement-calculator/internal/calculation"
	"github.com/rpgo/retirement-calculator/internal/domain"
	"github.com/shopspring/decimal"
)

type CalculateRequest struct {
	Name        string  `json:"name"`
	BirthDate   string  `json:"birthDate"`
	HireDate    string  `json:"hireDate"`
	Salary      float64 `json:"salary"`
	High3       float64 `json:"high3"`
	TSPTrad     float64 `json:"tspTrad"`
	TSPRoth     float64 `json:"tspRoth"`
	TSPContrib  float64 `json:"tspContrib"`
	RetireDate  string  `json:"retireDate"`
	State       string  `json:"state"`
	SSAge       int     `json:"ssAge"`
	SSEstimate  float64 `json:"ssEstimate"`
	TSPStrategy string  `json:"tspStrategy"`
}

type CalculateResponse struct {
	RetirementDate      string  `json:"retirementDate"`
	FERSAnnual          float64 `json:"fersAnnual"`
	SupplementAnnual    float64 `json:"supplementAnnual"`
	TSPBalance          float64 `json:"tspBalance"`
	NetIncome           float64 `json:"netIncome"`
	TSPWithdrawal       float64 `json:"tspWithdrawal"`
	StrategyDescription string  `json:"strategyDescription"`
}

func HandleCalculate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Parse dates
	birthDate, _ := time.Parse("2006-01-02", req.BirthDate)
	hireDate, _ := time.Parse("2006-01-02", req.HireDate)
	retireDate, _ := time.Parse("2006-01-02", req.RetireDate)

	// Create configuration
	cfg := &domain.Configuration{
		PersonalDetails: map[string]domain.Employee{
			"person_a": {
				Name:                   req.Name,
				BirthDate:              birthDate,
				HireDate:               hireDate,
				CurrentSalary:          decimal.NewFromFloat(req.Salary),
				High3Salary:            decimal.NewFromFloat(req.High3),
				TSPBalanceTraditional:  decimal.NewFromFloat(req.TSPTrad),
				TSPBalanceRoth:         decimal.NewFromFloat(req.TSPRoth),
				TSPContributionPercent: decimal.NewFromFloat(req.TSPContrib),
				// Use user input for SS
				SSBenefitFRA:                   decimal.NewFromFloat(req.SSEstimate),
				SSBenefit62:                    decimal.NewFromFloat(req.SSEstimate * 0.70), // Rough estimate
				SSBenefit70:                    decimal.NewFromFloat(req.SSEstimate * 1.24), // Rough estimate
				FEHBPremiumPerPayPeriod:        decimal.NewFromInt(200),
				SurvivorBenefitElectionPercent: decimal.Zero,
			},
			"person_b": {
				Name:          "Spouse (Placeholder)",
				BirthDate:     birthDate,
				HireDate:      hireDate,
				CurrentSalary: decimal.Zero,
				High3Salary:   decimal.Zero,
				// Zero out spouse for simple calc
				TSPBalanceTraditional:          decimal.Zero,
				TSPBalanceRoth:                 decimal.Zero,
				TSPContributionPercent:         decimal.Zero,
				SSBenefitFRA:                   decimal.Zero,
				SSBenefit62:                    decimal.Zero,
				SSBenefit70:                    decimal.Zero,
				FEHBPremiumPerPayPeriod:        decimal.Zero,
				SurvivorBenefitElectionPercent: decimal.Zero,
			},
		},
		GlobalAssumptions: domain.GlobalAssumptions{
			InflationRate:           decimal.NewFromFloat(0.025),
			FEHBPremiumInflation:    decimal.NewFromFloat(0.05),
			TSPReturnPreRetirement:  decimal.NewFromFloat(0.07),
			TSPReturnPostRetirement: decimal.NewFromFloat(0.05),
			COLAGeneralRate:         decimal.NewFromFloat(0.02),
			ProjectionYears:         30,
			CurrentLocation: domain.Location{
				State: req.State,
			},
			FederalRules: domain.FederalRules{
				FEHBConfig: domain.FEHBConfig{
					PayPeriodsPerYear:           26,
					RetirementCalculationMethod: "same_as_active",
					RetirementPremiumMultiplier: decimal.NewFromFloat(1.0),
				},
			},
		},
		Scenarios: []domain.Scenario{
			{
				Name: "Web Scenario",
				PersonA: domain.RetirementScenario{
					EmployeeName:          "person_a",
					RetirementDate:        retireDate,
					SSStartAge:            req.SSAge,
					TSPWithdrawalStrategy: req.TSPStrategy,
					// Add required fields for specific strategies if needed
					TSPWithdrawalRate: func() *decimal.Decimal {
						if req.TSPStrategy == "variable_percentage" {
							d := decimal.NewFromFloat(0.04)
							return &d
						}
						return nil
					}(),
					AnnuityPayoutRate: func() *decimal.Decimal {
						if req.TSPStrategy == "fixed_annuity" {
							d := decimal.NewFromFloat(0.06) // Example rate
							return &d
						}
						return nil
					}(),
				},
				PersonB: domain.RetirementScenario{
					EmployeeName:          "person_b",
					RetirementDate:        retireDate,
					SSStartAge:            62,
					TSPWithdrawalStrategy: "4_percent_rule",
				},
			},
		},
	}

	// Run calculation
	engine := calculation.NewCalculationEngineWithConfig(cfg.GlobalAssumptions)
	results, err := engine.RunScenarios(cfg)
	if err != nil {
		http.Error(w, fmt.Sprintf("Calculation failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Extract results
	scenario := results.Scenarios[0]

	// Find first year of retirement to get pension values
	var fersAnnual, supplementAnnual, netIncome, tspWithdrawal decimal.Decimal
	for _, cf := range scenario.Projection {
		if cf.Date.After(retireDate) {
			fersAnnual = cf.PensionPersonA
			supplementAnnual = cf.FERSSupplementPersonA
			netIncome = cf.NetIncome
			tspWithdrawal = cf.TSPWithdrawalPersonA
			break
		}
	}

	resp := CalculateResponse{
		RetirementDate:      retireDate.Format("2006-01-02"),
		FERSAnnual:          fersAnnual.InexactFloat64(),
		SupplementAnnual:    supplementAnnual.InexactFloat64(),
		TSPBalance:          scenario.FinalTSPBalance.InexactFloat64(),
		NetIncome:           netIncome.InexactFloat64(),
		TSPWithdrawal:       tspWithdrawal.InexactFloat64(),
		StrategyDescription: req.TSPStrategy,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
