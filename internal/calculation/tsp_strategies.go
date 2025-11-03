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
