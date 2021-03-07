package hash

import (
	"io"

	"github.com/minio/highwayhash"
)

// An empty key is given to highwayhash. This should be ok considering it is
// being used for its performance characteristics and NOT for its security
// features.
var key = make([]byte, 32)

func HashFile(r io.Reader) (n int64, _ []byte, err error) {
	hh, err := highwayhash.New(key)
	if err != nil {
		return
	}
	n, err = io.Copy(hh, r)
	if err != nil {
		return
	}
	return n, hh.Sum(nil), nil
}

func Sum256(data []byte) ([]byte, error) {
	hh, err := highwayhash.New(key)
	if err != nil {
		return nil, err
	}
	hh.Write(data)
	return hh.Sum(nil), err
}
