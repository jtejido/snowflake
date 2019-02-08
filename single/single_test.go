package single

import (
	"reflect"
	"testing"
	"time"
)

func TestMarshalJSON(t *testing.T) {
	id := SingleWorkerID(13587)
	expected := "\"13587\""

	bytes, err := id.MarshalJSON()
	if err != nil {
		t.Error("Unexpected error during MarshalJSON")
	}

	if string(bytes) != expected {
		t.Errorf("Got %s, expected %s", string(bytes), expected)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	tt := []struct {
		json        string
		expectedId  SingleWorkerID
		expectedErr error
	}{
		{`"13587"`, 13587, nil},
		{`1`, 0, JSONSyntaxError{[]byte(`1`)}},
		{`"invalid`, 0, JSONSyntaxError{[]byte(`"invalid`)}},
	}

	for _, tc := range tt {
		var id SingleWorkerID
		err := id.UnmarshalJSON([]byte(tc.json))
		if !reflect.DeepEqual(err, tc.expectedErr) {
			t.Errorf("Expected to get error '%s' decoding JSON, but got '%s'", tc.expectedErr, err)
		}

		if id != tc.expectedId {
			t.Errorf("Expected to get ID '%s' decoding JSON, but got '%s'", tc.expectedId, id)
		}
	}
}

func TestGetID(t *testing.T) {
	var x, y SingleWorkerID
	node, _ := New(1)
	for i := 0; i < 10; i++ {
		x, _ = node.NextID()

		y, _ = node.GetID(time.Unix(0, x.Time()*1000000), x.Sequence())

		if x != y {
			t.Errorf("Expected %v, got %v", x, y)
		}
	}
}

func BenchmarkNextID(b *testing.B) {

	node, _ := New(1)

	b.ReportAllocs()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		node.NextID()
	}
}

func BenchmarkGetID(b *testing.B) {
	var x SingleWorkerID
	node, _ := New(1)
	x, _ = node.NextID()
	time := time.Unix(0, x.Time()*1000000)
	seq := x.Sequence()

	b.ReportAllocs()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		node.GetID(time, seq)
	}
}

func BenchmarkUnmarshal(b *testing.B) {

	node, _ := New(1)
	id, _ := node.NextID()
	bytes, _ := id.MarshalJSON()

	var id2 SingleWorkerID

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = id2.UnmarshalJSON(bytes)
	}
}

func BenchmarkMarshal(b *testing.B) {

	node, _ := New(1)
	id, _ := node.NextID()

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = id.MarshalJSON()
	}
}
