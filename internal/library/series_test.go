package library

import "testing"

func TestCleanSeriesName(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"[www.ebookism.net] everything_s eventual", "everything s eventual"},
		{"[www.ebookism.net] from a buick", "from a buick"},
		{"[www.ebookism.net] mile", "mile"},
		{"The Expanse", "The Expanse"},
		{"  spaced   out  ", "spaced out"},
		{"", ""},
		{"1", ""},
		{"1-s2", ""},
		{"ARN18147_R600", ""},
		{"ARN30018-AR", ""},
		{"ARN30494-AR", ""},
		{"ARN7779_AR600", ""},
		{"Honor Bound Series, #5", "Honor Bound Series"},
		{"Extreme_Privacy_Linux_Devices-2024", "Extreme Privacy Linux Devices"},
		{"Cyber-Reports-2019", "Cyber Reports"},
		{"Badge of Honor", "Badge of Honor"},
		{"Jack Ryan/John Clark", "Jack Ryan/John Clark"},
		{"Foo__Bar Series", "Foo Bar Series"},
		{"Alpha   Beta", "Alpha Beta"},
	}
	for _, tc := range tests {
		if got := CleanSeriesName(tc.in); got != tc.want {
			t.Errorf("CleanSeriesName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
