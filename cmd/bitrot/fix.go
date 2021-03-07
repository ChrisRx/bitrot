package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ChrisRx/bitrot/pkg/hash"
	"github.com/ChrisRx/bitrot/pkg/hashdb"
)

var fixCommand = &cobra.Command{
	Use:  "fix <path>",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		f, err := os.OpenFile(args[0], os.O_RDWR, 0666)
		if err != nil {
			return err
		}
		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			return err
		}
		path, err := filepath.Abs(args[0])
		if err != nil {
			return err
		}
		db, err := hashdb.Open(dbPath)
		if err != nil {
			return err
		}
		file, err := db.Get(path)
		if err != nil {
			return err
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			return err
		}
		_, h, err := hash.HashFile(bytes.NewReader(data))
		if err != nil {
			return err
		}
		if bytes.Equal(h, file.Hash) {
			fmt.Printf("The file is already good, homie\n")
			return nil
		}
		t := &twiddler{r: f}
		for i := 0; int64(i) < fi.Size(); i++ {
			for j := 0; j < 8; j++ {
				if err := t.Twiddle(int64(i), int64(j)); err != nil {
					return err
				}
				data, err := ioutil.ReadAll(t)
				if err != nil {
					return err
				}
				_, h, err := hash.HashFile(bytes.NewReader(data))
				if err != nil {
					return err
				}
				if _, err := t.Seek(0, os.SEEK_SET); err != nil {
					return err
				}
				if bytes.Equal(h, file.Hash) {
					fmt.Printf("%x: \n\t%q\n", h, data)
					buf := make([]byte, 1)
					if _, err := f.ReadAt(buf, int64(i)); err != nil {
						return err
					}
					buf[0] = buf[0] ^ (1 << j)
					if _, err := f.WriteAt(buf, int64(i)); err != nil {
						return err
					}
				}
			}
		}
		return nil
	},
}

type twiddler struct {
	r           io.ReadSeeker
	pos         int64
	offset, bit int64
}

func (t *twiddler) Read(p []byte) (n int, err error) {
	n, err = t.r.Read(p)
	if t.pos <= t.offset && t.offset < t.pos+int64(len(p)) {
		p[t.offset-t.pos] = uint8(p[t.offset-t.pos]) ^ (1 << t.bit)
	}
	t.pos += int64(n)
	return
}

func (t *twiddler) Reset() error {
	_, err := t.Seek(0, os.SEEK_SET)
	return err
}

func (t *twiddler) Seek(offset int64, whence int) (int64, error) {
	t.pos = offset
	return t.r.Seek(offset, whence)
}

func (t *twiddler) Twiddle(offset, bit int64) error {
	t.offset = offset
	t.bit = bit
	return t.Reset()
}
