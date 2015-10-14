package cmd

import "testing"

func TestCopy(t *testing.T) {
	if err := Copy("data string"); err != nil {
		t.Errorf("Copy error is %s", err)
	}
}

func TestName(t *testing.T) {
	cmd := New("ls")

	if cmd.Name != "ls" {
		t.Errorf("Expected .Name to be 'ls' but instead it was '%s'", cmd.Name)
	}
}

func TestString(t *testing.T) {
	cmd := New("ls").WithArgs("-a", "-l")

	if cmd.String() != "ls -a -l" {
		t.Errorf("Expected .String() to be 'ls -a -l' but got '%s'", cmd.String())
	}
}

func TestArgs(t *testing.T) {
	cmd := New("ls").WithArgs("-a", "-l")

	if len(cmd.Args) != 2 {
		t.Errorf("Expected 2 args, got %d", len(cmd.Args))
	}
}

func TestCombinedOutput(t *testing.T) {
	cmd := New("ls")

	_, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Expected output, got error '%s'", err)
	}
}

func TestRun(t *testing.T) {
	cmd := New("ls")

	err := cmd.Run()
	if err != nil {
		t.Errorf("Expected no errors, got error '%s'", err)
	}
}
