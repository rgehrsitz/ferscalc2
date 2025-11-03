package calculation

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestFourPercentRule(t *testing.T) {
	strategy := NewFourPercentRule(decimal.NewFromInt(100000), decimal.NewFromFloat(0.02))

	t.Run("first year withdrawal equals initial amount", func(t *testing.T) {
		got := strategy.CalculateWithdrawal(decimal.NewFromInt(100000), 1, decimal.Zero, 65, false, decimal.Zero)
		want := decimal.NewFromInt(4000)
		if !got.Equal(want) {
			t.Fatalf("expected %s got %s", want, got)
		}
	})

	t.Run("second year increases with inflation", func(t *testing.T) {
		got := strategy.CalculateWithdrawal(decimal.NewFromInt(100000), 2, decimal.Zero, 66, false, decimal.Zero)
		want := decimal.NewFromInt(4000).Mul(decimal.NewFromFloat(1.02))
		if !got.Equal(want) {
			t.Fatalf("expected %s got %s", want, got)
		}
	})

	t.Run("RMD overrides smaller withdrawal", func(t *testing.T) {
		rmd := decimal.NewFromInt(6000)
		got := strategy.CalculateWithdrawal(decimal.NewFromInt(100000), 5, decimal.Zero, 75, true, rmd)
		if !got.Equal(rmd) {
			t.Fatalf("expected RMD %s got %s", rmd, got)
		}
	})

	t.Run("withdrawal capped by balance", func(t *testing.T) {
		got := strategy.CalculateWithdrawal(decimal.NewFromInt(2000), 1, decimal.Zero, 65, false, decimal.Zero)
		want := decimal.NewFromInt(2000)
		if !got.Equal(want) {
			t.Fatalf("expected %s got %s", want, got)
		}
	})
}

func TestNeedBasedWithdrawal(t *testing.T) {
	strategy := NewNeedBasedWithdrawal(decimal.NewFromInt(500))

	t.Run("withdraws target amount", func(t *testing.T) {
		got := strategy.CalculateWithdrawal(decimal.NewFromInt(100000), 1, decimal.Zero, 65, false, decimal.Zero)
		want := decimal.NewFromInt(6000)
		if !got.Equal(want) {
			t.Fatalf("expected %s got %s", want, got)
		}
	})

	t.Run("negative target treated as zero", func(t *testing.T) {
		negStrategy := NewNeedBasedWithdrawal(decimal.NewFromInt(-300))
		got := negStrategy.CalculateWithdrawal(decimal.NewFromInt(100000), 1, decimal.Zero, 65, false, decimal.Zero)
		if !got.Equal(decimal.Zero) {
			t.Fatalf("expected zero withdrawal got %s", got)
		}
	})

	t.Run("RMD overrides smaller target", func(t *testing.T) {
		rmd := decimal.NewFromInt(10000)
		got := strategy.CalculateWithdrawal(decimal.NewFromInt(100000), 1, decimal.Zero, 75, true, rmd)
		if !got.Equal(rmd) {
			t.Fatalf("expected RMD %s got %s", rmd, got)
		}
	})

	t.Run("withdrawal capped by balance", func(t *testing.T) {
		got := strategy.CalculateWithdrawal(decimal.NewFromInt(3000), 1, decimal.Zero, 65, false, decimal.Zero)
		want := decimal.NewFromInt(3000)
		if !got.Equal(want) {
			t.Fatalf("expected %s got %s", want, got)
		}
	})
}

func TestVariablePercentageWithdrawal(t *testing.T) {
	strategy := NewVariablePercentageWithdrawal(decimal.NewFromFloat(0.05))

	t.Run("withdraws percentage of balance", func(t *testing.T) {
		got := strategy.CalculateWithdrawal(decimal.NewFromInt(200000), 1, decimal.Zero, 65, false, decimal.Zero)
		want := decimal.NewFromInt(10000)
		if !got.Equal(want) {
			t.Fatalf("expected %s got %s", want, got)
		}
	})

	t.Run("RMD overrides smaller percentage", func(t *testing.T) {
		rmd := decimal.NewFromInt(15000)
		got := strategy.CalculateWithdrawal(decimal.NewFromInt(200000), 1, decimal.Zero, 75, true, rmd)
		if !got.Equal(rmd) {
			t.Fatalf("expected RMD %s got %s", rmd, got)
		}
	})

	t.Run("withdrawal capped by balance", func(t *testing.T) {
		highRate := NewVariablePercentageWithdrawal(decimal.NewFromFloat(0.90))
		got := highRate.CalculateWithdrawal(decimal.NewFromInt(1000), 1, decimal.Zero, 65, false, decimal.Zero)
		want := decimal.NewFromInt(900)
		if !got.Equal(want) {
			t.Fatalf("expected %s got %s", want, got)
		}

		got = highRate.CalculateWithdrawal(decimal.NewFromInt(1000), 1, decimal.Zero, 75, true, decimal.NewFromInt(2000))
		if !got.Equal(decimal.NewFromInt(1000)) {
			t.Fatalf("expected RMD clamped to balance got %s", got)
		}
	})
}
