package ssh

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/mutagen-io/mutagen/pkg/filesystem"
)

func TestCopy(t *testing.T) {
	// If localhost SSH support isn't available, then skip this test.
	if os.Getenv("MUTAGEN_TEST_SSH") != "true" {
		t.Skip()
	}

	// Compute source path.
	source := filepath.Join(t.TempDir(), "source")

	// Create contents.
	contents := []byte{0, 1, 2, 3, 4, 5, 6}

	// Attempt to write to a temporary file.
	if err := filesystem.WriteFileAtomic(source, contents, 0600); err != nil {
		t.Fatal("atomic file write failed:", err)
	}

	// Grab our username.
	user, err := user.Current()
	if err != nil {
		t.Fatal("unable to query user data:", err)
	}

	// Create a transport.
	transport := &sshTransport{
		user: user.Username,
		host: "localhost",
		port: 22,
	}

	// Compute the destination path.
	// HACK: Technically agent.Transport implementations only need to support
	// remote destination paths that are file names relative to the home
	// directory. For testing, however, we don't want to copy into the home
	// directory, and since we know our Copy implementation can support
	// arbitrary remote paths, we use one.
	destination := filepath.Join(t.TempDir(), "destination")

	// Copy the file.
	if err := transport.Copy(source, destination); err != nil {
		t.Fatal("unable to copy file:", err)
	}

	// Verify that the file exists.
	if _, err := os.Lstat(destination); err != nil {
		t.Error("unable to verify that destination exists")
	}
}

func TestCopyHomeRelativePathUsesSSHCommand(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell command test is POSIX-only")
	}

	temporaryDirectory := t.TempDir()
	fakeSSH := filepath.Join(temporaryDirectory, "ssh")
	fakeSCP := filepath.Join(temporaryDirectory, "scp")
	argumentsPath := filepath.Join(temporaryDirectory, "arguments")
	copiedPath := filepath.Join(temporaryDirectory, "copied")
	sourcePath := filepath.Join(temporaryDirectory, "source")

	contents := []byte("agent contents")
	if err := os.WriteFile(sourcePath, contents, 0600); err != nil {
		t.Fatal("unable to write source file:", err)
	}

	sshScript := `#!/bin/sh
{
	for argument do
		printf '%s\n' "$argument"
	done
} > "$MUTAGEN_TEST_COPY_ARGUMENTS"
cat > "$MUTAGEN_TEST_COPY_OUTPUT"
`
	if err := os.WriteFile(fakeSSH, []byte(sshScript), 0700); err != nil {
		t.Fatal("unable to write fake ssh:", err)
	}

	scpScript := `#!/bin/sh
echo scp invoked > "$MUTAGEN_TEST_COPY_ARGUMENTS"
exit 97
`
	if err := os.WriteFile(fakeSCP, []byte(scpScript), 0700); err != nil {
		t.Fatal("unable to write fake scp:", err)
	}

	t.Setenv("MUTAGEN_SSH_PATH", temporaryDirectory)
	t.Setenv("MUTAGEN_SSH_CONFIG_PATH", "")
	t.Setenv("MUTAGEN_TEST_COPY_ARGUMENTS", argumentsPath)
	t.Setenv("MUTAGEN_TEST_COPY_OUTPUT", copiedPath)

	transport := &sshTransport{
		user: "coder",
		host: "example.com",
		port: 22,
	}
	remoteName := filesystem.HomeDirectorySpecial + "/.mutagen-agent-test"
	if err := transport.Copy(sourcePath, remoteName); err != nil {
		t.Fatal("unable to copy file:", err)
	}

	copied, err := os.ReadFile(copiedPath)
	if err != nil {
		t.Fatal("unable to read copied file:", err)
	} else if string(copied) != string(contents) {
		t.Error("copied file contents do not match")
	}

	arguments, err := os.ReadFile(argumentsPath)
	if err != nil {
		t.Fatal("unable to read fake ssh arguments:", err)
	}
	argumentsString := string(arguments)
	if strings.Contains(argumentsString, "scp invoked") {
		t.Fatal("copy invoked scp instead of ssh")
	}
	if !strings.Contains(argumentsString, "coder@example.com") {
		t.Error("ssh target was not present in arguments")
	}
	if !strings.Contains(argumentsString, "umask 077 && cat > \"$HOME\"/'.mutagen-agent-test'") {
		t.Error("home-relative copy command was not present in arguments")
	}
}

func TestHomeRelativePOSIXDestination(t *testing.T) {
	if destination, ok := homeRelativePOSIXDestination("relative"); ok {
		t.Errorf("relative path converted unexpectedly: %s", destination)
	}

	destination, ok := homeRelativePOSIXDestination("~/.mutagen-agent-test")
	if !ok {
		t.Fatal("home-relative path was not converted")
	}
	expected := `"$HOME"/'.mutagen-agent-test'`
	if destination != expected {
		t.Errorf("destination mismatch: expected %s, got %s", expected, destination)
	}

	destination, ok = homeRelativePOSIXDestination("~/path/with'quote")
	if !ok {
		t.Fatal("quoted path was not converted")
	}
	expected = `"$HOME"/'path/with'\''quote'`
	if destination != expected {
		t.Errorf("quoted destination mismatch: expected %s, got %s", expected, destination)
	}
}

func TestCommandOutput(t *testing.T) {
	// If localhost SSH support isn't available, then skip this test.
	if os.Getenv("MUTAGEN_TEST_SSH") != "true" {
		t.Skip()
	}

	// Compute a command to run.
	command := "env"
	if runtime.GOOS == "windows" {
		command = "cmd /c set"
	}

	// Compute expected output content.
	content := "HOME="
	if runtime.GOOS == "windows" {
		content = "PROCESSOR_ARCHITECTURE="
	}

	// Grab our username.
	user, err := user.Current()
	if err != nil {
		t.Fatal("unable to query user data:", err)
	}

	// Create a transport.
	transport := &sshTransport{
		user: user.Username,
		host: "localhost",
		port: 22,
	}

	// Attempt to execute the command.
	// TODO: Should we also verify that an extracted HOME/USERPROFILE value
	// matches the expected home directory since we've already queried the user?
	if command, err := transport.Command(command); err != nil {
		t.Fatal("unable to create command:", err)
	} else if output, err := command.Output(); err != nil {
		t.Fatal("unable to run command:", err)
	} else if !strings.Contains(string(output), content) {
		t.Error("output does not contain expected content")
	} else if !utf8.Valid(output) {
		t.Error("output not in UTF-8 encoding")
	}
}
