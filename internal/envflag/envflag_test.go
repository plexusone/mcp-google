package envflag

import "testing"

func TestBool(t *testing.T) {
	tests := []struct {
		name  string
		given bool
		value bool
		env   string
		want  bool
	}{
		{"nothing set", false, false, "", false},
		{"environment turns it on", false, false, "true", true},
		{"environment other than true", false, false, "1", false},
		{"flag turns it on", true, true, "", true},
		{"flag on, environment off", true, true, "false", true},
		{"explicit false beats the environment", true, false, "true", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Bool(tt.given, tt.value, tt.env); got != tt.want {
				t.Errorf("Bool(%v, %v, %q) = %v, want %v", tt.given, tt.value, tt.env, got, tt.want)
			}
		})
	}
}
