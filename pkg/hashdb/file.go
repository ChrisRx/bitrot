package hashdb

import (
	"encoding/binary"

	"github.com/pkg/errors"
)

type File struct {
	ModTime int64
	Hash    []byte
}

func (f File) Marshal() ([]byte, error) {
	data := make([]byte, 40)
	binary.BigEndian.PutUint64(data, uint64(f.ModTime))
	_ = copy(data[8:], f.Hash[:32])
	return data, nil
}

func (f *File) Unmarshal(data []byte) error {
	if len(data) != 40 {
		return errors.Errorf("invalid data size: %d", len(data))
	}
	f.ModTime = int64(binary.BigEndian.Uint64(data[:8]))
	f.Hash = data[8:]
	return nil
}
