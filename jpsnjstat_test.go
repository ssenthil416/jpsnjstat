package jpsnjstat

import (
	"testing"

	"github.com/influxdata/telegraf/testutil"
)

func TestJpsnJstatValue(t *testing.T) {
	jj := Jpsnjstat{}

	var acc testutil.Accumulator
	acc.GatherError(jj.Gather)

	fields := map[string]interface{}{
		"s0c": 333.0,
		"s1c": 444.0,
	}

	acc.AssertContainsTaggedFields(t, "Jpsnjstat", fields, nil)
}
