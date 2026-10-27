package ui

import (
	"strings"
	"testing"
)

func TestBrailleCanvas(t *testing.T) {
	c := NewBrailleCanvas(5, 2)
	// Plot a ramp: 0, 10, 20 ... 100
	vals := []float64{0, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100}
	c.RenderAreaGraph(vals, 100.0)

	out := c.String()
	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
	t.Logf("Rendered Braille output:\n%s", out)
}
