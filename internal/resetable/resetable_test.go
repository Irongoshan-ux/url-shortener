package resetable

import "testing"

func TestResetableStruct_Reset(t *testing.T) {
	s := "hello"
	rs := &ResetableStruct{
		i:    42,
		str:  "world",
		strP: &s,
		s:    []int{1, 2, 3},
		m:    map[string]string{"a": "b"},
		child: &ResetableStruct{
			i:   7,
			str: "nested",
		},
	}

	rs.Reset()

	if rs.i != 0 {
		t.Fatalf("i: got %d, want 0", rs.i)
	}
	if rs.str != "" {
		t.Fatalf("str: got %q, want empty", rs.str)
	}
	if rs.strP == nil || *rs.strP != "" {
		t.Fatalf("strP: got %v, want pointer to empty string", rs.strP)
	}
	if len(rs.s) != 0 {
		t.Fatalf("s len: got %d, want 0", len(rs.s))
	}
	if len(rs.m) != 0 {
		t.Fatalf("m: got len %d, want 0", len(rs.m))
	}
	if rs.child == nil || rs.child.i != 0 || rs.child.str != "" {
		t.Fatalf("child not reset: %+v", rs.child)
	}
}
