package leveldb

import (
	"fmt"
	"github.com/azer/logger"
	driver "github.com/syndtr/goleveldb/leveldb"
)

var log = logger.New("persistent-pool/leveldb")

type LevelDB struct {
	Client   *driver.DB
	Filename string
}

func New(filename string) (LevelDB, error) {
	level := LevelDB{
		Filename: filename,
	}

	client, err := driver.OpenFile(filename, nil)
	if err != nil {
		log.Error("Can not open LevelDB storage", logger.Attrs{
			"filename": filename,
		})

		return level, err
	}

	level.Client = client

	return level, nil
}

func (db *LevelDB) Close() error {
	return db.Client.Close()
}

func (db *LevelDB) Key(poolName string) []byte {
	return []byte(fmt.Sprintf("persistent-pool:%s", poolName))
}

func (db *LevelDB) Load(poolName string) ([]byte, error) {
	return db.Client.Get(db.Key(poolName), nil)
}

func (db *LevelDB) Write(poolName string, contents []byte) error {
	return db.Client.Put(db.Key(poolName), contents, nil)
}
