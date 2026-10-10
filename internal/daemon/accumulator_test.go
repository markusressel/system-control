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
