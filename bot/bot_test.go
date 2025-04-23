package bot

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestXxx(t *testing.T) {
	Bootstrap("7839396085:AAGUb1np0LzYhnBhIAmhHULGnyO_ZYrMCVI", "7104844225")
	SendDepositMessage(1000, decimal.NewFromFloat(0.1))
	SendWithdrawMessage(1000, decimal.NewFromFloat(0.1))
}
