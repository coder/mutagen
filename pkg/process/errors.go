package process

import (
	"os/exec"
	"strings"
	"unicode/utf8"

	"github.com/mutagen-io/mutagen/pkg/platform/terminal"
)

const (
	// posixCommandNotFoundFragment is a fragment of the error output returned
	// on some POSIX shells when a command is not found. The capitalization of
	// the word "command" is inconsistent between shells, so only part of the
	// word is used.
	posixCommandNotFoundFragment = "ommand not found"
	// windowsInvalidCommandFragment is a fragment of the error output returned
	// on Windows systems when a command is not recognized.
	windowsInvalidCommandFragment = "is not recognized as an internal or external command"
	// windowsCommandNotFoundFragment is a fragment of the error output returned
	// on Windows systems when a command cannot be found.
	windowsCommandNotFoundFragment = "The system cannot find the path specified"
	// windowsPowershellCommandNotFoundFragment is a fragment of the error output
	// returned on Windows systems running Powershell when a command cannot be
	// found.
	// Different Windows versions use slightly different error messages.
	// i.e. "is not recognized as the name of a cmdlet, function, script file, or operable program."
	//      "is not recognized as a name of a cmdlet, function, script file, or executable program."
	windowsPowershellCommandNotFoundFragment = "cmdlet, function, script file, or"
)

// OutputIsPOSIXCommandNotFound returns whether or not a process' error output
// represents a command not found error on POSIX systems.
func OutputIsPOSIXCommandNotFound(output string) bool {
	return strings.Contains(output, posixCommandNotFoundFragment)
}

// OutputIsWindowsInvalidCommand returns whether or not a process' error output
// represents an invalid command error on Windows.
func OutputIsWindowsInvalidCommand(output string) bool {
	return strings.Contains(output, windowsInvalidCommandFragment)
}

// OutputIsWindowsCommandNotFound returns whether or not a process' error output
// represents a command not found error on Windows.
func OutputIsWindowsCommandNotFound(output string) bool {
	return strings.Contains(output, windowsCommandNotFoundFragment)

}

// OutputIsWindowsPowershellCommandNotFound returns whether or not a process' error
// output represents a command not found error from Windows running Powershell.
func OutputIsWindowsPowershellCommandNotFound(output string) bool {
	return strings.Contains(output, windowsPowershellCommandNotFoundFragment)
}

// ExtractExitErrorMessage is a utility function that will attempt to extract
// the Stderr portion of the specified error, assuming it is an
// os/exec.ExitError. If the error is not an os/exec.ExitError, or if the Stderr
// field is not UTF-8 encoded, or if the message in Stderr is empty after
// stripping surrounding white space, then an empty string is returned. This
// function will perform control character neutralization on any returned value.
func ExtractExitErrorMessage(err error) string {
	if exitErr, ok := err.(*exec.ExitError); ok && utf8.Valid(exitErr.Stderr) {
		return terminal.NeutralizeControlCharacters(strings.TrimSpace(string(exitErr.Stderr)))
	}
	return ""
}
