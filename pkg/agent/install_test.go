package agent

import (
	"os"
	"os/exec"
	"testing"
)

// NOTE: Unfortunately the Install() method can't be tested directly, but it is
// tested indirectly by integration tests.

type echoTransport struct {
	stdout string
	failCommand bool
}

func (t *echoTransport) Copy(_, _ string) error { return nil }

func (t *echoTransport) Command(_ string) (*exec.Cmd, error) {
	if t.failCommand {
		return exec.Command("false"), nil
	}
	return exec.Command("printf", "%s", t.stdout), nil
}

func (t *echoTransport) ClassifyError(_ *os.ProcessState, _ string) (bool, bool, error) {
	return false, false, nil
}

// TestRemoteHomeDirectory validates parsing and validation of the remote home
// directory returned over a transport.
func TestRemoteHomeDirectory(t *testing.T) {
	testCases := []struct {
		name        string
		stdout      string
		failCommand bool
		expected    string
		expectError bool
	}{
		{name: "simple", stdout: "/home/ubuntu", expected: "/home/ubuntu"},
		{name: "trailing newline", stdout: "/home/ubuntu\n", expected: "/home/ubuntu"},
		{name: "surrounding whitespace", stdout: "  /home/ubuntu \n", expected: "/home/ubuntu"},
		{name: "empty", stdout: "", expectError: true},
		{name: "unexpanded variable", stdout: "$HOME", expectError: true},
		{name: "relative", stdout: "home/ubuntu", expectError: true},
		{name: "command error", failCommand: true, expectError: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			transport := &echoTransport{stdout: testCase.stdout, failCommand: testCase.failCommand}
			home, err := remoteHomeDirectory(transport)
			if testCase.expectError {
				if err == nil {
					t.Fatalf("expected an error but got home %q", home)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if home != testCase.expected {
				t.Fatalf("expected home %q but got %q", testCase.expected, home)
			}
		})
	}
}
