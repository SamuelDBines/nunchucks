package nunchucks

import (
	"reflect"
	"testing"
)

func TestPipeTo(t *testing.T) {
	add := func(v any, args ...any) any {
		return v.(int) + args[0].(int)
	}

	got := PipeOf(2).To(add, 3).To(add, 5).Value
	if got != 10 {
		t.Fatalf("want 10, got %v", got)
	}
}

func TestCombinerParsesArgs(t *testing.T) {
	join := func(v any, args ...any) any {
		out := []any{v}
		out = append(out, args...)
		return out
	}

	steps := []PipelineStep{
		SA(join, "1", "true", "[2,3]", `{"x":1}`),
	}
	got := Combiner(steps)("start")
	want := []any{"start", 1, true, []any{float64(2), float64(3)}, map[string]any{"x": float64(1)}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected\nwant: %#v\n got: %#v", want, got)
	}
}

func TestParseVar(t *testing.T) {
	cases := []struct {
		in   any
		want any
	}{
		{" 42 ", 42},
		{"false", false},
		{"null", nil},
		{"hello", "hello"},
	}

	for _, tc := range cases {
		got := ParseVar(tc.in)
		if !reflect.DeepEqual(got, tc.want) {
			t.Fatalf("in=%#v want=%#v got=%#v", tc.in, tc.want, got)
		}
	}
}
