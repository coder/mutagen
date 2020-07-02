package main

import (
	"os"

	"github.com/pkg/errors"

	"github.com/mutagen-io/mutagen/pkg/command"
	"github.com/mutagen-io/mutagen/pkg/filesystem/locking"
)

func main() {
	// Validate arguments and extract the lock path.
	if len(os.Args) != 2 {
		command.Fatal(errors.New("invalid number of arguments"))
	} else if os.Args[1] == "" {
		command.Fatal(errors.New("empty lock path"))
	}
	path := os.Args[1]

	// Create a locker and attempt to acquire the lock.
	if locker, err := locking.NewLocker(path, 0600); err != nil {
		command.Fatal(errors.New("unable to create filesystem locker"))
	} else if err = locker.Lock(false); err != nil {
		command.Fatal(errors.Wrap(err, "lock acquisition failed"))
	} else if err = locker.Unlock(); err != nil {
		command.Fatal(errors.Wrap(err, "lock release failed"))
	} else if err = locker.Close(); err != nil {
		command.Fatal(errors.Wrap(err, "locker closure failed"))
	}
}
