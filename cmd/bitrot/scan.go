package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/ChrisRx/bitrot/pkg/hash"
	"github.com/ChrisRx/bitrot/pkg/hashdb"
)

var (
	bytesRead, numFiles uint64

	scanCommand = &cobra.Command{
		Use:  "scan <path>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}
			db, err := hashdb.Open(dbPath)
			if err != nil {
				return err
			}
			defer db.Close()

			if err := WalkFiles(root, func(path string) error {
				f, err := os.Open(path)
				if err != nil {
					return err
				}
				defer f.Close()

				n, hash, err := hash.HashFile(f)
				if err != nil {
					return err
				}

				atomic.AddUint64(&numFiles, 1)
				atomic.AddUint64(&bytesRead, uint64(n))

				fi, err := f.Stat()
				if err != nil {
					return errors.Wrapf(err, "cannot read file info")
				}
				return db.Save(path, hashdb.File{
					ModTime: fi.ModTime().UnixNano(),
					Hash:    hash,
				})
			}); err != nil {
				return err
			}
			fmt.Printf("files: %d bytes: %s\n", numFiles, humanize.Bytes(bytesRead))
			return nil
		},
	}
)
