package agent

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/google/uuid"

	"github.com/mutagen-io/mutagen/pkg/filesystem"
	"github.com/mutagen-io/mutagen/pkg/logging"
	"github.com/mutagen-io/mutagen/pkg/prompting"
)

// Install installs the current binary to the appropriate location for an agent
// binary with the current Mutagen version.
func Install() error {
	// Compute the destination.
	destination, err := installPath()
	if err != nil {
		return fmt.Errorf("unable to compute agent destination: %w", err)
	}

	// Compute the path to the current executable.
	executablePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("unable to determine executable path: %w", err)
	}

	// Relocate the current executable to the installation path.
	if err = filesystem.Rename(nil, executablePath, nil, destination, true); err != nil {
		return fmt.Errorf("unable to relocate agent executable: %w", err)
	}

	// Success.
	return nil
}

// install attempts to probe an endpoint and install the appropriate agent
// binary over the specified transport.
func install(logger *logging.Logger, transport Transport, prompter string, cmdExe bool) error {
	// Detect the target platform.
	goos, goarch, posix, err := probe(transport, prompter)
	if err != nil {
		return fmt.Errorf("unable to probe remote platform: %w", err)
	}

	// Find the appropriate agent binary. Ensure that it's cleaned up when we're
	// done with it.
	if err := prompting.Message(prompter, "Extracting agent..."); err != nil {
		return fmt.Errorf("unable to message prompter: %w", err)
	}
	agentExecutable, err := ExecutableForPlatform(goos, goarch, "")
	if err != nil {
		return fmt.Errorf("unable to get agent for platform: %w", err)
	}
	defer os.Remove(agentExecutable)

	// Copy the agent to the remote. We use a unique identifier for the
	// temporary destination. For Windows remotes, we add a ".exe" suffix, which
	// will automatically make the file executable on the remote (POSIX systems
	// are handled separately below). For POSIX systems, we add a dot prefix to
	// hide the executable.
	if err := prompting.Message(prompter, "Copying agent..."); err != nil {
		return fmt.Errorf("unable to message prompter: %w", err)
	}
	randomUUID, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("unable to generate UUID for agent copying: %w", err)
	}
	remoteFileName := BaseName + randomUUID.String()
	if goos == "windows" {
		remoteFileName += ".exe"
	}
	if posix {
		remoteFileName = "." + remoteFileName
	}
	fullRemotePath := remotePathFromHome(cmdExe, remoteFileName)
	// On POSIX remotes, the agent binary is copied with scp and then executed
	// over ssh using this same path. Historically that path is "~/"-prefixed and
	// relies on "~" resolving identically for both steps. That assumption breaks
	// when the remote SSH/SFTP working directory isn't the home directory (for
	// example, Coder workspaces configured with an explicit directory, or
	// devcontainers whose workspace folder differs from $HOME): scp resolves the
	// path relative to the working directory while the ssh exec expands "~" to
	// $HOME, so the freshly-copied binary can't be found. Resolve the absolute
	// home directory once and use it for both the copy and the invocation so they
	// agree regardless of the remote working directory. If resolution fails, fall
	// back to the previous "~"-relative behavior.
	if posix {
		if home, homeErr := remoteHomeDirectory(transport); homeErr == nil {
			fullRemotePath = path.Join(home, remoteFileName)
		} else {
			logger.Infof("unable to resolve remote home directory, using ~-relative agent path: %v", homeErr)
		}
	}

	if err = transport.Copy(agentExecutable, fullRemotePath); err != nil {
		return fmt.Errorf("unable to copy agent binary: %w", err)
	}

	// For cases where we're copying from a Windows system to a POSIX remote,
	// invoke "chmod +x" to add executability back to the copied binary. This is
	// necessary under the specified circumstances because as soon as the agent
	// binary is extracted from the bundle, it will lose its executability bit
	// since Windows can't preserve this. This will also be applied to Windows
	// POSIX remotes, but a "chmod +x" there will just be a no-op.
	if runtime.GOOS == "windows" && posix {
		if err := prompting.Message(prompter, "Setting agent executability..."); err != nil {
			return fmt.Errorf("unable to message prompter: %w", err)
		}
		executabilityCommand := fmt.Sprintf("chmod +x %s", fullRemotePath)
		if err := run(transport, executabilityCommand); err != nil {
			return fmt.Errorf("unable to set agent executability: %w", err)
		}
	}

	// Invoke the remote installation.
	if err := prompting.Message(prompter, "Installing agent..."); err != nil {
		return fmt.Errorf("unable to message prompter: %w", err)
	}
	installCommand := fmt.Sprintf("%s %s", fullRemotePath, CommandInstall)
	if err := run(transport, installCommand); err != nil {
		return fmt.Errorf("unable to invoke agent installation: %w", err)
	}

	// Success.
	return nil
}

// remoteHomeDirectory resolves the absolute path of the home directory on a
// POSIX remote by querying $HOME over the transport. It's used to construct an
// absolute agent installation path so that the scp copy and ssh execution steps
// agree even when the remote SSH/SFTP working directory isn't the home
// directory (e.g. Coder workspaces with a configured directory or devcontainer
// workspace folder).
func remoteHomeDirectory(transport Transport) (string, error) {
	out, err := output(transport, `echo "$HOME"`)
	if err != nil {
		return "", fmt.Errorf("unable to query remote home directory: %w", err)
	}
	home := strings.TrimSpace(string(out))
	if !strings.HasPrefix(home, "/") {
		return "", fmt.Errorf("invalid remote home directory: %q", home)
	}
	return home, nil
}
