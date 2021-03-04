package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/dustin/go-humanize"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	bytesRead, numFiles uint64
	dbPath              string

	rootCmd = &cobra.Command{
		Use:  "bitrot <path>",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := filepath.Abs(args[0])
			if err != nil {
				return err
			}
			db, err := OpenHashDB(dbPath, root)
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

				n, hash, err := db.Hash(f)
				if err != nil {
					return err
				}

				atomic.AddUint64(&numFiles, 1)
				atomic.AddUint64(&bytesRead, uint64(n))

				fi, err := f.Stat()
				if err != nil {
					return errors.Wrapf(err, "cannot read file info")
				}
				return db.Save(path, File{fi.ModTime().UnixNano(), hash})
			}); err != nil {
				return err
			}
			fmt.Printf("files: %d bytes: %s\n", numFiles, humanize.Bytes(bytesRead))
			return nil
		},
	}
)

func init() {
	rootCmd.Flags().StringVar(&dbPath, "dir", "data", "db path")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
