// interface
package kvstore

import "io"

type KVStore interface {
	// byte array can store all of the data
	// and hash func need byte array as input
	Put(key []byte, value []byte) error

	Get(key []byte) ([]byte, error)

	Delete(key []byte) error

	Exist(key []byte)
}

// there are database controller
type KVDatabase interface {
	KVStore   // this is a embeded interface
	io.Closer // Closer is a interface in io package so this embeded in KVDatabase
}
