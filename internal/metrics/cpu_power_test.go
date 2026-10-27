package metrics

import (
	"testing"
	"time"
)

func TestCPUPowerCollector(t *testing.T) {
	c := NewCPUPowerCollector()
	defer c.Close()

	time.Sleep(200 * time.Millisecond)
	pwr := c.QueryPower()
	t.Logf("CPUPowerCollector returned: %.2f Watts", pwr)

	fanRPM := QueryCPUFanRPM()
	t.Logf("QueryCPUFanRPM returned: %d RPM", fanRPM)
}
