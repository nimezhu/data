package data

import (
	"fmt"

	"github.com/boltdb/bolt"
	"github.com/nimezhu/netio"
)

//TODO manage boltdb _idx _format
// not load all into memory.
// try to manage hic bigwig and bigbed
// Data Manager API
/*
   type DataManager interface {
       AddURI(uri string, key string) error
       Del(string) error
       ServeTo(*mux.Router)
       List() []string
       Get(string) (string, bool)
       Move(key1 string, key2 string) bool
     }
*/
var bucketNameIdx = "_idx"

type IndexesManager struct {
	db *bolt.DB
}

//TODO
func getFormat(uri string) (string, error) {
	f, err := netio.NewReadSeeker(uri)
	fmt.Println("TODO", f, err)
	if err != nil {
		return "", err
	}
	return "unknown", nil
}

//TODO
func getIndex(uri string) ([]byte, error) {
	format, err := getFormat(uri)
	fmt.Println("TODO", format, err)
	switch format {
	case "bigwig":
		fmt.Println("TODO")
	case "bigbed":
		fmt.Println("TODO")
	case "hic":
		fmt.Println("TODO")
	default:
		fmt.Println("UNKNOWN FORMAT")
	}
	return []byte{}, nil
}

/* AddURI: Save URI to Index and name to key in total */
func (m *IndexesManager) AddURI(uri string, key string) error {
	tx, err := m.db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Commit()
	bucketIdx, err := tx.CreateBucketIfNotExists([]byte(bucketNameIdx))
	if err != nil {
		return err
	}
	//format := getFormat(uri)
	index, _ := getIndex(uri)
	bucketIdx.Put([]byte(uri), index)
	return nil
}

/* TODO
func NewIndexesManager() *IndexesManager {

}
*/
