package main

import (
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func TestRootCommand(t *testing.T) {
	// Reset viper for clean state
	viper.Reset()

	// Set up a temporary config file to avoid loading real config
	tempDir := t.TempDir()
	configFile := tempDir + "/config.yaml"

	// Create a minimal config file
	configContent := `
sonarr:
  - name: "test"
    baseurl: "http://localhost:8989"
    apikey: "test-key"
triggersearch: false
batchsize: 1
`

	err := os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test config: %v", err)
	}

	// Set config file path
	viper.SetConfigFile(configFile)

	// Test that root command can be created and has expected properties
	if rootCmd.Use != "score-checker" {
		t.Errorf("expected Use to be 'score-checker', got '%s'", rootCmd.Use)
	}

	if rootCmd.Short != "Check Sonarr/Radarr episodes/movies for low custom format scores" {
		t.Errorf("unexpected Short description: %s", rootCmd.Short)
	}

	// Test that daemon command exists
	daemonCmd := findCommand(rootCmd, "daemon")
	if daemonCmd == nil {
		t.Error("daemon command not found")
	} else {
		if daemonCmd.Use != "daemon" {
			t.Errorf("expected daemon Use to be 'daemon', got '%s'", daemonCmd.Use)
		}
	}
}

func TestCommandFlags(t *testing.T) {
	// Test that required flags exist
	flags := []string{"triggersearch", "batchsize", "interval"}

	for _, flagName := range flags {
		flag := rootCmd.PersistentFlags().Lookup(flagName)
		if flag == nil {
			t.Errorf("flag '%s' not found", flagName)
		}
	}

	// Test flag defaults
	triggerSearchFlag := rootCmd.PersistentFlags().Lookup("triggersearch")
	if triggerSearchFlag != nil && triggerSearchFlag.DefValue != "false" {
		t.Errorf("expected triggersearch default to be 'false', got '%s'", triggerSearchFlag.DefValue)
	}

	batchSizeFlag := rootCmd.PersistentFlags().Lookup("batchsize")
	if batchSizeFlag != nil && batchSizeFlag.DefValue != "5" {
		t.Errorf("expected batchsize default to be '5', got '%s'", batchSizeFlag.DefValue)
	}

	intervalFlag := rootCmd.PersistentFlags().Lookup("interval")
	if intervalFlag != nil && intervalFlag.DefValue != "1h" {
		t.Errorf("expected interval default to be '1h', got '%s'", intervalFlag.DefValue)
	}
}

func TestInitFunction(t *testing.T) {
	// Test that init function sets up commands correctly
	// The init function should have already run, so we verify its effects

	// Check that daemon command was added
	if len(rootCmd.Commands()) == 0 {
		t.Error("expected root command to have subcommands")
	}

	found := false
	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "daemon" {
			found = true
			break
		}
	}

	if !found {
		t.Error("daemon command was not added by init function")
	}
}

// Helper function to find a command by name
func findCommand(parent *cobra.Command, name string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		if cmd.Use == name {
			return cmd
		}
	}
	return nil
}

// Test the main function indirectly by testing command execution
func TestMainCommandExecution(t *testing.T) {
	// This test verifies that the command structure is set up correctly
	// We can't easily test main() directly without refactoring, but we can
	// test that the commands are properly configured

	// Reset args to avoid interference from actual command line args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	// Test root command setup
	if rootCmd == nil {
		t.Fatal("rootCmd is nil")
	}

	// Test that all expected persistent flags are bound to viper
	expectedBindings := []string{"triggersearch", "batchsize", "interval"}

	for _, binding := range expectedBindings {
		// We can't easily test viper bindings directly, but we can verify
		// the flags exist and have been set up
		flag := rootCmd.PersistentFlags().Lookup(binding)
		if flag == nil {
			t.Errorf("expected flag %s to exist", binding)
		}
	}
}
