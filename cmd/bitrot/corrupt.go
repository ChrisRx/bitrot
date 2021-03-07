package main

import (
	"math/rand"
	"os"
	"syscall"
	"time"

	"github.com/ChrisRx/bitrot/pkg/prompt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var corruptCommand = &cobra.Command{
	Use:  "corrupt <path>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		fi, err := os.Stat(args[0])
		if err != nil {
			return err
		}
		if fi.Size() == 0 {
			return errors.Errorf("file %q is empty", args[0])
		}
		mtime := fi.ModTime()
		stat := fi.Sys().(*syscall.Stat_t)
		atime := time.Unix(int64(stat.Atim.Sec), int64(stat.Atim.Nsec))

		if !prompt.Confirmf("This will flip a random bit in the following file:\n\t%s\nAre you sure you want to proceed?", args[0]) {
			return nil
		}

		f, err := os.OpenFile(args[0], os.O_RDWR, 0666)
		if err != nil {
			return err
		}

		// select a random bit in the file
		offset := rand.Int63n(fi.Size())
		bit := rand.Intn(8)

		buf := make([]byte, 1)
		if _, err := f.ReadAt(buf, offset); err != nil {
			return err
		}

		// flip
		buf[0] = uint8(buf[0]) ^ (1 << bit)
		if _, err := f.WriteAt(buf, offset); err != nil {
			return err
		}
		if err := os.Chtimes(args[0], atime, mtime); err != nil {
			return err
		}
		return nil
	},
}
