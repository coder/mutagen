package logging

import (
	"bytes"
	"strings"
	"testing"
)

// scpErrorOutput is representative of the stderr that scp produces when the
// OpenSSH configuration diverts it onto a broken ProxyCommand. The first line
// is a symptom and the second line is the cause, which makes it a good test of
// whether the logger preserves everything it is given.
const scpErrorOutput = "unable to run SCP process: /bin/sh: 1: Syntax error: Bad for loop variable\n" +
	"/bin/sh: 1: exec: /c/Program Files/Coder/bin/coder.exe: not found\n" +
	"scp: Connection closed"

func TestLoggerPreservesMultiLineMessages(t *testing.T) {
	buffer := &bytes.Buffer{}
	NewLogger(LevelInfo, buffer).Sublogger("sync").Sublogger("sync_IxTbhpj3").Info(scpErrorOutput)

	output := buffer.String()
	for _, expected := range []string{
		"Bad for loop variable",
		"coder.exe: not found",
		"scp: Connection closed",
	} {
		if !strings.Contains(output, expected) {
			t.Errorf("output does not contain %q:\n%s", expected, output)
		}
	}

	// Every line needs the level and scope prefix, otherwise continuation
	// lines cannot be attributed to a session when reading a log.
	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 log lines, got %d:\n%s", len(lines), output)
	}
	for _, line := range lines {
		if !strings.Contains(line, "[I] [sync.sync_IxTbhpj3]") {
			t.Errorf("line missing level and scope prefix: %q", line)
		}
	}
}

func TestLoggerNeutralizesCarriageReturns(t *testing.T) {
	buffer := &bytes.Buffer{}
	NewLogger(LevelInfo, buffer).Info("progress\roverwritten")

	output := buffer.String()
	if strings.ContainsRune(output, '\r') {
		t.Errorf("output contains a raw carriage return: %q", output)
	}
	// The text after the carriage return must survive, escaped rather than
	// discarded.
	if !strings.Contains(output, `progress\roverwritten`) {
		t.Errorf("text after carriage return was not preserved: %q", output)
	}
}
