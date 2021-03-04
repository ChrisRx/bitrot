package main

import (
	"io/fs"
	"log"
	"path/filepath"
	"runtime"
	"sync"
)

func WalkFiles(root string, visit func(string) error) (err error) {
	var wg sync.WaitGroup
	wait := make(chan struct{}, runtime.NumCPU())
	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check if non-regular file (e.g. Directory, SymLink, etc).
		if d.Type()&fs.ModeType != 0 {
			return nil
		}

		wait <- struct{}{}
		wg.Add(1)
		go func() {
			defer func() {
				<-wait
				wg.Done()
			}()
			if err := visit(path); err != nil {
				log.Println(err)
				return
			}
		}()

		return nil
	}); err != nil {
		return err
	}
	wg.Wait()
	return nil
}
