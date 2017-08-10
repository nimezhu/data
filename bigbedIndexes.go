package data

import (
	"bytes"
	"io/ioutil"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/nimezhu/indexed/bbi"
	"github.com/nimezhu/netio"
)

func (m *BigBedManager) SaveIndexes(fn string) error {
	err := m.saveIndexes(fn, false)
	return err
}
func (m *BigBedManager) UpdateIndexes(fn string) error {
	err := m.saveIndexes(fn, true)
	return err
}
func (m *BigBedManager) saveIndexes(fn string, update bool) error {
	bucketName := m.dbname
	bucketNameIdx := "_idx"
	bucketNameFmt := "_format"
	db, err := bolt.Open(fn, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return err
	}
	defer db.Close()
	tx, err := db.Begin(true)
	if err != nil {
		return err
	}
	defer tx.Commit()

	bucketFormat, err := tx.CreateBucketIfNotExists([]byte(bucketNameFmt))
	if err != nil {
		return err
	}
	bucketFormat.Put([]byte(m.dbname), []byte("bigbed"))
	bucketIdx, err := tx.CreateBucketIfNotExists([]byte(bucketNameIdx))
	if err != nil {
		return err
	}
	bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
	if err != nil {
		return err
	}
	for k, uri := range m.uriMap {
		v := bucketIdx.Get([]byte(uri))
		if v == nil || update {
			tmpfile, err := ioutil.TempFile("", "temp") //TODO use buffer as WriteSeeker?
			if err != nil {
				return err
			}

			m.dataMap[k].Reader.WriteIndex(tmpfile)
			tmpfile.Seek(0, 0)
			buffer, err := ioutil.ReadAll(tmpfile)
			if err != nil {
				return err
			}
			tmpfile.Close()
			os.Remove(tmpfile.Name())
			bucketIdx.Put([]byte(uri), []byte(buffer))
		}
		bucket.Put([]byte(k), []byte(uri))
	}
	return nil
}
func (m *BigBedManager) LoadIndexes(fn string) error {
	db, err := bolt.Open(fn, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		panic(err)
		return err
	}
	defer db.Close()
	db.View(func(tx *bolt.Tx) error {
		bIdx := tx.Bucket([]byte("_idx"))
		b := tx.Bucket([]byte(m.dbname))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			m.uriMap[string(k)] = string(v)
			idx := bIdx.Get(v)
			reader, err := netio.NewReadSeeker(string(v))
			if err != nil {
				panic(err)
				return err
			}
			bwf := bbi.NewBbiReader(reader)
			bwf.ReadIndex(bytes.NewReader(idx))
			bwr := bbi.NewBigBedReader(bwf)
			m.dataMap[string(k)] = bwr
		}
		return nil
	})
	return nil
}
