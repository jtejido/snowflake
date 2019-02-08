package snowflake

import (
	"testing"
	"time"
)

func TestTilNextMillis(t *testing.T) {
	now := time.Now()
	ms := now.UnixNano() / 1000000
	v := TilNextMillis(ms)
	expected := now.Add(time.Millisecond).UnixNano() / 1000000 // add 1ms
	if expected != v {
		t.Errorf("Got %v, expected %v", v, expected)
	}
}
