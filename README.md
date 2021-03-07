# bitrot

A toy bitrot detector using [highwayhash](https://github.com/minio/highwayhash) for signatures and [badger](https://github.com/dgraph-io/badger) for signature storage.

## Building

```shell
make
```

This will place the binary in the `bin` folder.

## Usage

Scan the files you want to protect:

```shell
❯ bin/bitrot scan <path>
```
