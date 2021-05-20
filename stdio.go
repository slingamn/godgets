// Copyright (c) 2021 Shivaram Lingamneni
// released under the MIT license

package godgets

import (
	"os"
)

// StdioRWC wraps os.Stdin and os.Stdout to expose an io.ReadWriteCloser.
type StdioRWC struct{}

func (s StdioRWC) Read(buf []byte) (n int, err error) {
	return os.Stdin.Read(buf)
}

func (s StdioRWC) Write(buf []byte) (n int, err error) {
	return os.Stdout.Write(buf)
}

func (s StdioRWC) Close() (err error) {
	return nil
}
