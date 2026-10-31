package main

import "testing"

func TestWriteEnabled(t *testing.T) {
	tests := []struct {
		name      string
		flagSet   bool
		flagValue bool
		env       string
		want      bool
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
			if got := writeEnabled(tt.flagSet, tt.flagValue, tt.env); got != tt.want {
				t.Errorf("writeEnabled(%v, %v, %q) = %v, want %v", tt.flagSet, tt.flagValue, tt.env, got, tt.want)
			}
		})
	}
}

// The flag, not just the helper, must carry the explicit false through
// applyEnvDefaults.
func TestApplyEnvDefaultsKeepsAnExplicitFlag(t *testing.T) {
	t.Setenv("MCP_GOOGLE_ENABLE_WRITE", "true")
	prev := enableWrite
	t.Cleanup(func() {
		enableWrite = prev
		if err := rootCmd.PersistentFlags().Set("enable-write", "false"); err != nil {
			t.Fatal(err)
		}
		rootCmd.PersistentFlags().Lookup("enable-write").Changed = false
	})

	enableWrite = false
	rootCmd.PersistentFlags().Lookup("enable-write").Changed = false
	applyEnvDefaults()
	if !enableWrite {
		t.Fatal("the environment did not enable write mode when the flag was absent")
	}

	if err := rootCmd.PersistentFlags().Set("enable-write", "false"); err != nil {
		t.Fatal(err)
	}
	applyEnvDefaults()
	if enableWrite {
		t.Error("MCP_GOOGLE_ENABLE_WRITE=true overrode an explicit --enable-write=false")
	}
}
