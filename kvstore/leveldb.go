package kvstore

import "github.com/syndtr/goleveldb/leveldb"

type levelDB struct {
	db *leveldb.DB
}

func NewLevelDB(path string) *levelDB {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		panic(err)
	}
	return &levelDB{db: db}
}

// levelDB package is already implement the KVStore interface
// so we don't need to implement the Get and Put method
// we have to Nested the import package method
func (ldb *levelDB) Put(key, value []byte) error {
	return ldb.db.Put(key, value, nil)
}

func (ldb *levelDB) Get(key []byte) ([]byte, error) {
	return ldb.db.Get(key, nil)
}

func (ldb *levelDB) Delete(key []byte) error {
	return ldb.db.Delete(key, nil)
}

func (ldb *levelDB) Exist(key []byte) (bool, error) {
	return ldb.db.Has(key, nil)
}

func (ldb *levelDB) Closer() error {
	return ldb.db.Close()
}
