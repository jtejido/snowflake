package multi

import (
	"reflect"
	"testing"
	"time"
)

func TestMarshalJSON(t *testing.T) {
	id := MultiWorkerID(13587)
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
		expectedId  MultiWorkerID
		expectedErr error
	}{
		{`"13587"`, 13587, nil},
		{`1`, 0, JSONSyntaxError{[]byte(`1`)}},
		{`"invalid`, 0, JSONSyntaxError{[]byte(`"invalid`)}},
	}

	for _, tc := range tt {
		var id MultiWorkerID
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
	var x, y MultiWorkerID
	node, _ := New(1, 1)
	for i := 0; i < 10; i++ {
		x, _ = node.NextID()

		y, _ = node.GetID(time.Unix(0, x.Time()*1000000), x.Sequence())

		if x != y {
			t.Errorf("Expected %v, got %v", x, y)
		}
	}
}

func BenchmarkNextID(b *testing.B) {

	node, _ := New(1, 1)

	b.ReportAllocs()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		node.NextID()
	}
}

func BenchmarkGetID(b *testing.B) {
	var x MultiWorkerID
	node, _ := New(1, 1)
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
	node, _ := New(1, 1)
	id, _ := node.NextID()
	bytes, _ := id.MarshalJSON()

	var id2 MultiWorkerID

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = id2.UnmarshalJSON(bytes)
	}
}

func BenchmarkMarshal(b *testing.B) {

	node, _ := New(1, 1)
	id, _ := node.NextID()

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_, _ = id.MarshalJSON()
	}
}
