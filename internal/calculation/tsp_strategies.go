package calculation

import "github.com/shopspring/decimal"

// TSPWithdrawalStrategy defines the interface for withdrawal strategies.
type TSPWithdrawalStrategy interface {
	CalculateWithdrawal(currentBalance decimal.Decimal, year int, targetIncome decimal.Decimal, age int, isRMDYear bool, rmdAmount decimal.Decimal) decimal.Decimal
	GetStrategyName() string
}

// FourPercentRule implements the 4% rule (inflation adjusted).
type FourPercentRule struct {
	InitialWithdrawalPercent decimal.Decimal
	InflationRate            decimal.Decimal
	InitialBalance           decimal.Decimal
	FirstWithdrawalAmount    decimal.Decimal
}

func NewFourPercentRule(initialBalance, inflationRate decimal.Decimal) *FourPercentRule {
	initialWithdrawal := initialBalance.Mul(decimal.NewFromFloat(0.04))
	return &FourPercentRule{
		InitialWithdrawalPercent: decimal.NewFromFloat(0.04),
		InflationRate:            inflationRate,
		InitialBalance:           initialBalance,
		FirstWithdrawalAmount:    initialWithdrawal,
	}
}

func (fpr *FourPercentRule) CalculateWithdrawal(currentBalance decimal.Decimal, year int, _ decimal.Decimal, _ int, isRMDYear bool, rmdAmount decimal.Decimal) decimal.Decimal {
	var withdrawal decimal.Decimal
	if year == 1 {
		withdrawal = fpr.FirstWithdrawalAmount
	} else {
		inflationFactor := decimal.NewFromFloat(1).Add(fpr.InflationRate)
		withdrawal = fpr.FirstWithdrawalAmount.Mul(inflationFactor.Pow(decimal.NewFromInt(int64(year - 1))))
	}
	if isRMDYear && withdrawal.LessThan(rmdAmount) {
		return rmdAmount
	}
	if withdrawal.GreaterThan(currentBalance) {
		return currentBalance
	}
	return withdrawal
}

func (fpr *FourPercentRule) GetStrategyName() string {
	return "4_percent_rule"
}

// NeedBasedWithdrawal withdraws a fixed monthly amount (with RMD safeguards).
type NeedBasedWithdrawal struct {
	TargetMonthlyWithdrawal decimal.Decimal
}

func NewNeedBasedWithdrawal(targetMonthly decimal.Decimal) *NeedBasedWithdrawal {
	return &NeedBasedWithdrawal{TargetMonthlyWithdrawal: targetMonthly}
}

func (nbw *NeedBasedWithdrawal) CalculateWithdrawal(currentBalance decimal.Decimal, _ int, _ decimal.Decimal, _ int, isRMDYear bool, rmdAmount decimal.Decimal) decimal.Decimal {
	withdrawal := nbw.TargetMonthlyWithdrawal.Mul(decimal.NewFromInt(12))
	if withdrawal.LessThan(decimal.Zero) {
		withdrawal = decimal.Zero
	}
	if isRMDYear && withdrawal.LessThan(rmdAmount) {
		withdrawal = rmdAmount
	}
	if withdrawal.GreaterThan(currentBalance) {
		return currentBalance
	}
	return withdrawal
}

func (nbw *NeedBasedWithdrawal) GetStrategyName() string {
	return "need_based"
}

// VariablePercentageWithdrawal withdraws a configurable percentage of the balance.
type VariablePercentageWithdrawal struct {
	WithdrawalRate decimal.Decimal
}

func NewVariablePercentageWithdrawal(withdrawalRate decimal.Decimal) *VariablePercentageWithdrawal {
	return &VariablePercentageWithdrawal{WithdrawalRate: withdrawalRate}
}

func (vpw *VariablePercentageWithdrawal) CalculateWithdrawal(currentBalance decimal.Decimal, _ int, _ decimal.Decimal, _ int, isRMDYear bool, rmdAmount decimal.Decimal) decimal.Decimal {
	withdrawal := currentBalance.Mul(vpw.WithdrawalRate)
	if isRMDYear && withdrawal.LessThan(rmdAmount) {
		if rmdAmount.GreaterThan(currentBalance) {
			return currentBalance
		}
		return rmdAmount
	}
	if withdrawal.GreaterThan(currentBalance) {
		return currentBalance
	}
	return withdrawal
}

func (vpw *VariablePercentageWithdrawal) GetStrategyName() string {
	return "variable_percentage"
}

// FixedAnnuity implements a fixed immediate annuity with guaranteed lifetime income.
// This strategy models converting TSP balance into an annuity contract that provides
// fixed monthly payments for life, with optional COLA adjustments and survivor benefits.
type FixedAnnuity struct {
	InitialPremium        decimal.Decimal // Amount invested in annuity (portion of TSP)
	AnnualPayoutRate      decimal.Decimal // Annual payout as percentage of premium (e.g., 0.055 for 5.5%)
	MonthlyPayment        decimal.Decimal // Calculated fixed monthly payment
	AnnualPayment         decimal.Decimal // Calculated fixed annual payment
	COLARate              decimal.Decimal // Annual cost-of-living adjustment rate (0 for no COLA)
	HasSurvivorBenefit    bool            // Whether payments continue to survivor
	SurvivorPayoutPercent decimal.Decimal // Percentage paid to survivor (e.g., 0.5 for 50%, 1.0 for 100%)
	GuaranteedYears       int             // Minimum guaranteed payment period (e.g., 10 years certain)
}

// NewFixedAnnuity creates a new fixed annuity withdrawal strategy.
// premium: The TSP balance amount converted to annuity
// annualPayoutRate: The annual payout rate (e.g., 0.055 for 5.5% annual payout)
// colaRate: Annual COLA adjustment (e.g., 0.02 for 2% annual increase, 0 for fixed payment)
// survivorPercent: Percentage paid to survivor (1.0 = 100%, 0.5 = 50%, 0 = no survivor benefit)
// guaranteedYears: Minimum years of guaranteed payments (e.g., 10 for "10 years certain")
func NewFixedAnnuity(premium, annualPayoutRate, colaRate, survivorPercent decimal.Decimal, guaranteedYears int) *FixedAnnuity {
	annualPayment := premium.Mul(annualPayoutRate)
	monthlyPayment := annualPayment.Div(decimal.NewFromInt(12))

	return &FixedAnnuity{
		InitialPremium:        premium,
		AnnualPayoutRate:      annualPayoutRate,
		MonthlyPayment:        monthlyPayment,
		AnnualPayment:         annualPayment,
		COLARate:              colaRate,
		HasSurvivorBenefit:    survivorPercent.GreaterThan(decimal.Zero),
		SurvivorPayoutPercent: survivorPercent,
		GuaranteedYears:       guaranteedYears,
	}
}

// CalculateWithdrawal returns the fixed annuity payment for the year.
// For fixed annuities, the payment is predetermined and doesn't depend on market performance.
// The currentBalance represents the notional "annuity value" but doesn't affect payments.
func (fa *FixedAnnuity) CalculateWithdrawal(currentBalance decimal.Decimal, year int, _ decimal.Decimal, _ int, isRMDYear bool, rmdAmount decimal.Decimal) decimal.Decimal {
	// Calculate payment with COLA adjustment if applicable
	var payment decimal.Decimal
	if year == 1 {
		payment = fa.AnnualPayment
	} else if fa.COLARate.GreaterThan(decimal.Zero) {
		// Apply COLA adjustment for subsequent years
		colaFactor := decimal.NewFromFloat(1).Add(fa.COLARate)
		payment = fa.AnnualPayment.Mul(colaFactor.Pow(decimal.NewFromInt(int64(year - 1))))
	} else {
		// Fixed payment with no COLA
		payment = fa.AnnualPayment
	}

	// Note: Annuity payments are NOT subject to RMD rules since the annuity
	// contract itself satisfies RMD requirements for the annuitized portion.
	// However, if there's remaining TSP balance not annuitized, that would need separate RMD.

	// For annuities, we don't cap by currentBalance because payments are guaranteed
	// by the insurance company regardless of account value
	return payment
}

func (fa *FixedAnnuity) GetStrategyName() string {
	return "fixed_annuity"
}

// GetMonthlyPayment returns the monthly payment amount for the current year
func (fa *FixedAnnuity) GetMonthlyPayment(year int) decimal.Decimal {
	annualPayment := fa.CalculateWithdrawal(decimal.Zero, year, decimal.Zero, 0, false, decimal.Zero)
	return annualPayment.Div(decimal.NewFromInt(12))
}
