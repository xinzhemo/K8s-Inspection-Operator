package rules

import "testing"

func TestCase(t *testing.T) {
	type testCase struct {
		name     string
		val      int32
		op       string
		target   int32
		expected bool
	}
	cases := []testCase{
		{"greater than - true", 10, ">", 5, true},
		{"greater than - false", 3, ">", 5, false},
		{"equal - true", 5, "==", 5, true},
		{"equal - false", 3, "==", 5, false},
		{"less than - true", 3, "<", 5, true},
		{"less than - false", 10, "<", 5, false},
		{"unknown op", 10, "??", 5, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := compare(tc.val, tc.op, tc.target)
			if result != tc.expected {
				t.Errorf("compare(%d,%s,%d) = %v,want %v", tc.val, tc.op, tc.target, result, tc.expected)
			}
		})
	}
}
