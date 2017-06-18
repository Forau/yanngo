package nnutils_test

import (
	"github.com/Forau/yanngo/nnutils"

	"testing"
)

func TestDefaultTickTable(t *testing.T) {
	ticks := nnutils.DEFAULT_TICK_TABLE
	t.Log("Default tick table: ", ticks)

	ttu := nnutils.NewDefaultTickTableUtil()

	for _, testData := range []float64{
		4.2346634,
		0.2346634,
		-0.2346634,
		53453, 234234,
		9999999999999999941,
	} {
		newVal, tick, from, to, decimals := ttu.AsTick(testData)

		t.Logf("From %f -> %f: (%f)  %f -> %f (%d)",
			testData, newVal, tick, from, to, decimals)
	}

	p1 := 4.8543
	p1a3 := ttu.AddTicks(p1, 3)
	t.Logf("From %f add 3 -> %f", p1, p1a3)
	p1a30 := ttu.AddTicks(p1, 30)
	t.Logf("From %f add 30 -> %f", p1, p1a30)
	p1s30 := ttu.AddTicks(p1, -30)
	t.Logf("From %f sub 30 -> %f", p1, p1s30)

	t.Logf("From %f sub 5 -> %f", 1.1, ttu.AddTicks(1.1, -5))
	t.Logf("From %f add 5 -> %f", 1.1, ttu.AddTicks(1.1, 5))
	t.Logf("From %f add 5000 -> %f", 1.1, ttu.AddTicks(1.1, 5000))
	t.Logf("From %f sub 5000 -> %f", 1.1, ttu.AddTicks(1.1, -5000))
	t.Logf("From %f add 0 -> %f", 1.123456789, ttu.AddTicks(1.123456789, 0))

	t.Logf("Ticks from 0.123 to 3.21 -> %d", ttu.TicksBetween(0.123, 3.21))
	t.Logf("Ticks from 3.21 to 0.123 -> %d", ttu.TicksBetween(3.21, 0.123))
	t.Logf("Ticks from 3.2112312412512 to 0.12324623452345 -> %d", ttu.TicksBetween(3.2112312412512, 0.12324623452345))
	t.Logf("Ticks from 0.1 to 0.1 -> %d", ttu.TicksBetween(0.1, 0.1))
	t.Logf("Ticks from 0.01 to 0.02 -> %d", ttu.TicksBetween(0.01, 0.02))
	t.Logf("Ticks from 0.01 to 0.02 -> %d", ttu.TicksBetween(0.01, 0.02))
	t.Logf("Ticks from 50.0 to 49.9 -> %d", ttu.TicksBetween(50.0, 49.9))
	t.Logf("Ticks from 50.0 to 99999999 -> %d", ttu.TicksBetween(50.0, 99999999))

	t.Logf("From %f add 0 -> %f", 43.34567, ttu.AddTicks(43.34567, 0))
	t.Logf("As string: 43.34567 -> %s", ttu.ToString(43.34567))
}
