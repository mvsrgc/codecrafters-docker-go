package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
)

func main() {
	command := os.Args[3]
	args := os.Args[4:]

	var exitCode int

	// Create a temporary directory that will become the root of our
	// executable.
	tempDir, err := os.MkdirTemp("", "mychroot")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(tempDir)

	// Ensure the temp directory has the correct permissions.
	if err := os.Chmod(tempDir, 0755); err != nil {
		panic(err)
	}

	//linkerPath := "/lib64/ld-linux-x86-64.so.2"
	//newLinkerPath := filepath.Join(tempDir, "lib64/ld-linux-x86-64.so.2")
	//if err := os.MkdirAll(filepath.Dir(newLinkerPath), 0755); err != nil {
	//	panic(err)
	//}
	//if err := copyFile(linkerPath, newLinkerPath); err != nil {
	//	panic(err)
	//}
	//if err := os.Chmod(newLinkerPath, 0755); err != nil {
	//	panic(err)
	//}

	//libDir := filepath.Join(tempDir, "lib", "x86_64-linux-gnu")
	//if err := os.MkdirAll(libDir, 0755); err != nil {
	//	panic(err)
	//}

	//// Copy libc.so.6
	//libcPath := "/lib/x86_64-linux-gnu/libc.so.6"
	//newLibcPath := filepath.Join(libDir, "libc.so.6")
	//if err := copyFile(libcPath, newLibcPath); err != nil {
	//	panic(err)
	//}

	binary := command
	newBinaryPath := filepath.Join(tempDir, filepath.Base(binary))

	// Copy the binary from the original location to the new root.
	if err := copyFile(binary, newBinaryPath); err != nil {
		panic(err)
	}

	// Ensure the new binary has the correct permissions.
	if err := os.Chmod(newBinaryPath, 0755); err != nil {
		panic(err)
	}

	// Enter the chroot.
	if err := syscall.Chroot(tempDir); err != nil {
		panic(err)
	}

	// Change to the root directory after chroot.
	if err := os.Chdir("/"); err != nil {
		panic(err)
	}

	// Ensure the binary exists and has the correct permissions.
	if _, err := os.Stat("/" + filepath.Base(binary)); err != nil {
		panic(err)
	}

	// Execute the command inside the chroot.
	cmd := exec.Command("/"+filepath.Base(binary), args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	}

	os.Exit(exitCode)
}

func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
