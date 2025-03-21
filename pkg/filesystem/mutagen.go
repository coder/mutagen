package filesystem

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mutagen-io/mutagen/pkg/mutagen"
)

const (
	// MutagenDataDirectoryName is the name of the global Mutagen data directory
	// inside the user's home directory.
	MutagenDataDirectoryName = ".mutagen"

	// MutagenDataDirectoryDevelopmentName is the name of the global Mutagen
	// data directory inside the user's home directory for development builds of
	// Mutagen.
	MutagenDataDirectoryDevelopmentName = ".mutagen-dev"

	// MutagenGlobalConfigurationName is the name of the global Mutagen
	// configuration file inside the user's home directory.
	MutagenGlobalConfigurationName = ".mutagen.yml"

	// MutagenDaemonDirectoryName is the name of the daemon storage directory
	// within the Mutagen data directory.
	MutagenDaemonDirectoryName = "daemon"

	// MutagenAgentsDirectoryName is the name of the agent storage directory
	// within the Mutagen data directory.
	MutagenAgentsDirectoryName = "agents"

	// MutagenSynchronizationSessionsDirectoryName is the name of the
	// synchronization session storage directory within the Mutagen data
	// directory.
	MutagenSynchronizationSessionsDirectoryName = "sessions"

	// MutagenSynchronizationCachesDirectoryName is the name of the
	// synchronization cache storage directory within the Mutagen data
	// directory.
	MutagenSynchronizationCachesDirectoryName = "caches"

	// MutagenSynchronizationArchivesDirectoryName is the name of the
	// synchronization archive storage directory within the Mutagen data
	// directory.
	MutagenSynchronizationArchivesDirectoryName = "archives"

	// MutagenSynchronizationStagingDirectoryName is the name of the
	// synchronization staging storage directory within the Mutagen data
	// directory.
	MutagenSynchronizationStagingDirectoryName = "staging"

	// MutagenForwardingDirectoryName is the name of the forwarding data
	// directory within the Mutagen data directory.
	MutagenForwardingDirectoryName = "forwarding"

	// MutagenIODirectoryName is the name of the mutagen.io data directory
	// within the Mutagen data directory. This directory is no longer used. It
	// was originally added to support the tunnel transport, but that no longer
	// exists. This constant is left in-place as a reminder to avoid using this
	// particular name for future functionality (at least without taking care to
	// avoid collision with any existing files).
	MutagenIODirectoryName = "mutagen.io"

	// MutagenLicensingDirectoryName is the name of the licensing data directory
	// within the Mutagen data directory.
	MutagenLicensingDirectoryName = "licensing"

	HomeDirectorySpecial = "~"
)

// Mutagen computes (and optionally creates) subdirectories inside the Mutagen
// data directory.
func Mutagen(create bool, pathComponents ...string) (string, error) {
	// Check if a data directory path has been explicitly specified. If not,
	// compute it using the standard procedure. Also track whether or not we
	// need to mark the directory as hidden on creation.
	mutagenDataDirectoryPath, ok := os.LookupEnv("MUTAGEN_DATA_DIRECTORY")
	var hide bool
	if ok {
		// Validate the provided path.
		if mutagenDataDirectoryPath == "" {
			return "", errors.New("provided data directory path is empty")
		} else if !filepath.IsAbs(mutagenDataDirectoryPath) {
			return "", errors.New("provided data directory path is not absolute")
		}
	} else {
		// Compute the path to the user's home directory.
		homeDirectory, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("unable to compute path to home directory: %w", err)
		}

		// Compute the path to the Mutagen data directory.
		if !mutagen.DevelopmentModeEnabled {
			mutagenDataDirectoryPath = filepath.Join(homeDirectory, MutagenDataDirectoryName)
		} else {
			mutagenDataDirectoryPath = filepath.Join(homeDirectory, MutagenDataDirectoryDevelopmentName)
		}

		// Flag the directory for hiding.
		hide = true
	}

	// Compute the target path.
	result := filepath.Join(mutagenDataDirectoryPath, filepath.Join(pathComponents...))

	// Handle directory creation, if requested.
	//
	// TODO: Should we iterate through each component and ensure the user hasn't
	// changed the directory permissions? MkdirAll won't reset them. But I
	// suppose the user may have changed them for whatever reason (though I
	// can't imagine any).
	if create {
		// Create the directory.
		if err := os.MkdirAll(result, 0700); err != nil {
			return "", fmt.Errorf("unable to create subpath: %w", err)
		}

		// Mark the directory as hidden, if necessary.
		if hide {
			if err := MarkHidden(mutagenDataDirectoryPath); err != nil {
				return "", fmt.Errorf("unable to hide Mutagen data directory: %w", err)
			}
		}
	}

	// Success.
	return result, nil
}
