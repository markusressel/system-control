package daemon

import "testing"

func TestAccumulator(t *testing.T) {
	var a Accumulator
	a.Add("x", 0.1)
	a.Add("x", 0.1)
	a.Add("y", -1)

	if got := a.Take("x"); got < 0.19 || got > 0.21 {
		t.Errorf("x = %v", got)
	}
	if got := a.Take("x"); got != 0 {
		t.Errorf("taken twice: %v", got)
	}
	if got := a.Take("y"); got != -1 {
		t.Errorf("y = %v", got)
	}
}

func TestAccumulator_Withdraw(t *testing.T) {
	var a Accumulator
	first := a.Add("x", 0.1)
	a.Add("x", 0.2)
	if !a.Withdraw(first) {
		t.Fatal("a pending change could not be withdrawn")
	}
	if got := a.Take("x"); got < 0.19 || got > 0.21 {
		t.Errorf("x = %v after withdrawing 0.1, want 0.2", got)
	}

	taken := a.Add("x", 0.3)
	a.Take("x")
	if a.Withdraw(taken) {
		t.Error("a change that was taken was withdrawn")
	}
	if got := a.Take("x"); got != 0 {
		t.Errorf("x = %v, want 0", got)
	}
}
