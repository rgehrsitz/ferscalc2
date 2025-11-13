package calculation

import (
	"testing"

	"github.com/rpgo/retirement-calculator/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestNewJerseyTaxCalculator(t *testing.T) {
	calc := NewNewJerseyTaxCalculator()

	t.Run("Working income - MFJ - no retirement", func(t *testing.T) {
		income := domain.TaxableIncome{
			Salary:             decimal.NewFromInt(100000),
			WageIncome:         decimal.NewFromInt(100000),
			FERSPension:        decimal.Zero,
			TSPWithdrawalsTrad: decimal.Zero,
			TaxableSSBenefits:  decimal.Zero,
			OtherTaxableIncome: decimal.Zero,
			InterestIncome:     decimal.Zero,
		}

		tax := calc.CalculateTax(income, false, "mfj", 45, 43)

		// Expected calculation:
		// First $20,000 at 1.4% = $280
		// Next $30,000 ($20,001-$50,000) at 1.75% = $525
		// Next $20,000 ($50,001-$70,000) at 2.45% = $490
		// Next $10,000 ($70,001-$80,000) at 3.5% = $350
		// Next $20,000 ($80,001-$100,000) at 5.525% = $1,105
		// Total = $2,750
		expected := decimal.NewFromInt(2750)
		assert.True(t, tax.Sub(expected).Abs().LessThan(decimal.NewFromInt(1)), 
			"Tax should be approximately $2,750, got %s", tax.StringFixed(2))
	})

	t.Run("Retirement income - MFJ age 62+ - under threshold", func(t *testing.T) {
		// Retirees age 62+ with total income $90,000 (under $100K threshold)
		// FERS pension: $40,000
		// TSP withdrawals: $30,000
		// Social Security: $24,000 (not taxed by NJ)
		// Total: $94,000 but SS excluded from taxable, so $70K for NJ purposes
		
		income := domain.TaxableIncome{
			Salary:             decimal.Zero,
			WageIncome:         decimal.Zero,
			FERSPension:        decimal.NewFromInt(40000),
			TSPWithdrawalsTrad: decimal.NewFromInt(30000),
			TaxableSSBenefits:  decimal.NewFromInt(24000), // This is federal taxable SS, NJ doesn't tax it
			OtherTaxableIncome: decimal.Zero,
			InterestIncome:     decimal.Zero,
		}

		tax := calc.CalculateTax(income, true, "mfj", 63, 65)

		// Retirement income = $40K (pension) + $30K (TSP) = $70K
		// Since total income ($70K) <= $100K threshold, full exclusion applies
		// Max exclusion for MFJ = $100K, so all $70K is excluded
		// Taxable income = $0
		// Tax = $0
		expected := decimal.Zero
		assert.True(t, tax.Equal(expected), "Tax should be $0 for fully excluded retirement income, got %s", tax.StringFixed(2))
	})

	t.Run("Retirement income - MFJ age 62+ - in phase-out range", func(t *testing.T) {
		// Retirees age 62+ with total income $110,000 (in $100K-$125K phase-out range)
		// FERS pension: $60,000
		// TSP withdrawals: $50,000
		// Social Security: $20,000 (not taxed by NJ, not counted toward threshold)
		// Total for NJ purposes: $110K
		
		income := domain.TaxableIncome{
			Salary:             decimal.Zero,
			WageIncome:         decimal.Zero,
			FERSPension:        decimal.NewFromInt(60000),
			TSPWithdrawalsTrad: decimal.NewFromInt(50000),
			TaxableSSBenefits:  decimal.NewFromInt(20000), // Not taxed by NJ
			OtherTaxableIncome: decimal.Zero,
			InterestIncome:     decimal.Zero,
		}

		tax := calc.CalculateTax(income, true, "mfj", 64, 66)

		// Retirement income = $60K (pension) + $50K (TSP) = $110K
		// Total income = $110K (in phase-out range $100K-$125K)
		// Exclusion percentage = 50%
		// Exclusion = min($110K, $100K max) * 50% = $50K
		// Taxable income = $110K - $50K = $60K
		// 
		// Tax calculation on $60K:
		// First $20,000 at 1.4% = $280
		// Next $30,000 ($20,001-$50,000) at 1.75% = $525
		// Next $10,000 ($50,001-$60,000) at 2.45% = $245
		// Total = $1,050
		expected := decimal.NewFromInt(1050)
		assert.True(t, tax.Sub(expected).Abs().LessThan(decimal.NewFromInt(1)), 
			"Tax should be approximately $1,050, got %s", tax.StringFixed(2))
	})

	t.Run("Retirement income - MFJ age 62+ - above limit", func(t *testing.T) {
		// Retirees age 62+ with total income $160,000 (above $150K limit)
		// No retirement income exclusion
		
		income := domain.TaxableIncome{
			Salary:             decimal.Zero,
			WageIncome:         decimal.Zero,
			FERSPension:        decimal.NewFromInt(80000),
			TSPWithdrawalsTrad: decimal.NewFromInt(80000),
			TaxableSSBenefits:  decimal.NewFromInt(20000), // Not taxed by NJ, not counted
			OtherTaxableIncome: decimal.Zero,
			InterestIncome:     decimal.Zero,
		}

		tax := calc.CalculateTax(income, true, "mfj", 67, 68)

		// Retirement income = $80K (pension) + $80K (TSP) = $160K
		// Total income = $160K (above $150K limit)
		// No exclusion applies
		// Taxable income = $160K
		// 
		// Tax calculation on $160K:
		// First $20,000 at 1.4% = $280
		// Next $30,000 ($20,001-$50,000) at 1.75% = $525
		// Next $20,000 ($50,001-$70,000) at 2.45% = $490
		// Next $10,000 ($70,001-$80,000) at 3.5% = $350
		// Next $70,000 ($80,001-$150,000) at 5.525% = $3,867.50
		// Next $10,000 ($150,001-$160,000) at 6.37% = $637
		// Total = $6,149.50
		expected := decimal.NewFromFloat(6149.50)
		assert.True(t, tax.Sub(expected).Abs().LessThan(decimal.NewFromInt(1)), 
			"Tax should be approximately $6,149.50, got %s", tax.StringFixed(2))
	})

	t.Run("Retirement income - Single age 62+ - under threshold", func(t *testing.T) {
		// Single retiree age 62+ with total income $70,000 (under $100K threshold)
		
		income := domain.TaxableIncome{
			Salary:             decimal.Zero,
			WageIncome:         decimal.Zero,
			FERSPension:        decimal.NewFromInt(40000),
			TSPWithdrawalsTrad: decimal.NewFromInt(30000),
			TaxableSSBenefits:  decimal.NewFromInt(15000), // Not taxed by NJ
			OtherTaxableIncome: decimal.Zero,
			InterestIncome:     decimal.Zero,
		}

		tax := calc.CalculateTax(income, true, "single", 65, 0)

		// Retirement income = $40K (pension) + $30K (TSP) = $70K
		// Since total income ($70K) <= $100K threshold, full exclusion applies
		// Max exclusion for single = $75K, so all $70K is excluded
		// Taxable income = $0
		// Tax = $0
		expected := decimal.Zero
		assert.True(t, tax.Equal(expected), "Tax should be $0 for fully excluded retirement income, got %s", tax.StringFixed(2))
	})

	t.Run("Social Security is not taxed by NJ", func(t *testing.T) {
		// Test that Social Security benefits are not included in NJ taxable income
		
		income := domain.TaxableIncome{
			Salary:             decimal.Zero,
			WageIncome:         decimal.Zero,
			FERSPension:        decimal.Zero,
			TSPWithdrawalsTrad: decimal.Zero,
			TaxableSSBenefits:  decimal.NewFromInt(30000), // Only SS income
			OtherTaxableIncome: decimal.Zero,
			InterestIncome:     decimal.Zero,
		}

		tax := calc.CalculateTax(income, true, "mfj", 67, 68)

		// Social Security is not taxed by NJ, so tax should be $0
		expected := decimal.Zero
		assert.True(t, tax.Equal(expected), "Tax should be $0 for SS-only income, got %s", tax.StringFixed(2))
	})

	t.Run("Retirement income - under age 62 - no exclusion", func(t *testing.T) {
		// Retiree under age 62 - no retirement income exclusion even if income is low
		
		income := domain.TaxableIncome{
			Salary:             decimal.Zero,
			WageIncome:         decimal.Zero,
			FERSPension:        decimal.NewFromInt(40000),
			TSPWithdrawalsTrad: decimal.NewFromInt(30000),
			TaxableSSBenefits:  decimal.NewFromInt(15000), // Not taxed by NJ
			OtherTaxableIncome: decimal.Zero,
			InterestIncome:     decimal.Zero,
		}

		tax := calc.CalculateTax(income, true, "mfj", 60, 58)

		// Retirement income = $40K (pension) + $30K (TSP) = $70K
		// Both under age 62, so no exclusion applies
		// Taxable income = $70K
		// 
		// Tax calculation on $70K:
		// First $20,000 at 1.4% = $280
		// Next $30,000 ($20,001-$50,000) at 1.75% = $525
		// Next $20,000 ($50,001-$70,000) at 2.45% = $490
		// Total = $1,295
		expected := decimal.NewFromInt(1295)
		assert.True(t, tax.Sub(expected).Abs().LessThan(decimal.NewFromInt(1)), 
			"Tax should be approximately $1,295 (no exclusion under age 62), got %s", tax.StringFixed(2))
	})
}

func TestComprehensiveTaxCalculatorWithNJ(t *testing.T) {
	t.Run("NJ state tax calculator is used when state is New Jersey", func(t *testing.T) {
		config := domain.StateLocalTaxConfig{
			State: "New Jersey",
			NewJerseyRetirementExclusionMFJ:    decimal.NewFromInt(100000),
			NewJerseyRetirementExclusionSingle: decimal.NewFromInt(75000),
			NewJerseyRetirementIncomeThreshold: decimal.NewFromInt(100000),
			NewJerseyRetirementIncomeLimit:     decimal.NewFromInt(150000),
		}

		federalRules := domain.FederalRules{
			StateLocalTaxConfig: config,
			FederalTaxConfig: domain.FederalTaxConfig{
				StandardDeductionMFJ:        decimal.NewFromInt(30000),
				AdditionalStandardDeduction: decimal.NewFromInt(1550),
			},
			FICATaxConfig: domain.FICATaxConfig{
				Year:                   2025,
				SocialSecurityWageBase: decimal.NewFromInt(176100),
				SocialSecurityRate:     decimal.NewFromFloat(0.062),
				MedicareRate:           decimal.NewFromFloat(0.0145),
				AdditionalMedicareRate: decimal.NewFromFloat(0.009),
				HighIncomeThresholdMFJ: decimal.NewFromInt(250000),
			},
		}

		calc := NewComprehensiveTaxCalculatorWithConfig(federalRules)
		assert.NotNil(t, calc.StateTaxCalc, "State tax calculator should be initialized")

		// Verify it's a NJ calculator by checking tax on retirement income
		income := domain.TaxableIncome{
			FERSPension:        decimal.NewFromInt(50000),
			TSPWithdrawalsTrad: decimal.NewFromInt(30000),
		}
		
		tax := calc.StateTaxCalc.CalculateTax(income, true, "mfj", 65, 65)
		// With NJ exclusion, this should be $0 (income under threshold)
		assert.True(t, tax.Equal(decimal.Zero), "NJ should exclude retirement income under threshold, got %s", tax.StringFixed(2))
	})

	t.Run("PA state tax calculator is used when state is Pennsylvania", func(t *testing.T) {
		config := domain.StateLocalTaxConfig{
			State:            "Pennsylvania",
			PennsylvaniaRate: decimal.NewFromFloat(0.0307),
		}

		federalRules := domain.FederalRules{
			StateLocalTaxConfig: config,
			FederalTaxConfig: domain.FederalTaxConfig{
				StandardDeductionMFJ:        decimal.NewFromInt(30000),
				AdditionalStandardDeduction: decimal.NewFromInt(1550),
			},
			FICATaxConfig: domain.FICATaxConfig{
				Year:                   2025,
				SocialSecurityWageBase: decimal.NewFromInt(176100),
				SocialSecurityRate:     decimal.NewFromFloat(0.062),
				MedicareRate:           decimal.NewFromFloat(0.0145),
				AdditionalMedicareRate: decimal.NewFromFloat(0.009),
				HighIncomeThresholdMFJ: decimal.NewFromInt(250000),
			},
		}

		calc := NewComprehensiveTaxCalculatorWithConfig(federalRules)
		assert.NotNil(t, calc.StateTaxCalc, "State tax calculator should be initialized")

		// Verify it's a PA calculator by checking tax on retirement income
		income := domain.TaxableIncome{
			FERSPension:        decimal.NewFromInt(50000),
			TSPWithdrawalsTrad: decimal.NewFromInt(30000),
		}
		
		tax := calc.StateTaxCalc.CalculateTax(income, true, "mfj", 65, 65)
		// PA should exclude all retirement income, so tax should be $0
		assert.True(t, tax.Equal(decimal.Zero), "PA should exclude all retirement income, got %s", tax.StringFixed(2))
	})
}
