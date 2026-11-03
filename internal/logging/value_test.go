package logging

import "testing"

func TestValue(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"plain", `"plain"`},
		{"line\nbreak", `"line\nbreak"`},
		{"carriage\rreturn", `"carriage\rreturn"`},
		{`with "quotes"`, `"with \"quotes\""`},
		{"unicode é", `"unicode é"`},
		{"", `""`},
	}
	for _, c := range cases {
		if got := Value(c.in); got != c.want {
			t.Errorf("Value(%q)=%s want %s", c.in, got, c.want)
		}
	}
}
