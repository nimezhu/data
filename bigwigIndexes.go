package data

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/nimezhu/indexed/bbi"
	"github.com/nimezhu/netio"
)

func (m *BigWigManager) Query(chr string, start int, end int) { //TODO REMOVE;
	for k, v := range m.bwMap {
		fmt.Println("DB:", k)
		binsize := v.GetBinsize(end-start, 500)
		fmt.Println("BINSIZE:", binsize)
		if iter, err := v.QueryBin(chr, start, end, binsize); err == nil {
			for i := range iter {
				fmt.Println(i.From, "\t", i.To, "\t", i.Sum, "\t", float64(i.Valid))
				if i.To == 0 {
					break
				}
			}
		}
	}
}
func (m *BigWigManager) SaveIndexes(fn string) error {
	err := m.saveIndexes(fn, false)
	return err
}
func (m *BigWigManager) UpdateIndexes(fn string) error {
	err := m.saveIndexes(fn, true)
	return err
}
func (m *BigWigManager) saveIndexes(fn string, update bool) error {
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
	bucketFormat.Put([]byte(m.dbname), []byte("bigwig"))

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
		if v == nil || update == true {
			tmpfile, err := ioutil.TempFile("", "temp") //TODO use buffer as WriteSeeker?
			if err != nil {
				return err
			}
			m.bwMap[k].Reader.WriteIndex(tmpfile)
			tmpfile.Seek(0, 0)
			buffer, err := ioutil.ReadAll(tmpfile)
			if err != nil {
				return err
			}
			tmpfile.Close()
			os.Remove(tmpfile.Name())
			fmt.Println("PUT", uri)
			bucketIdx.Put([]byte(uri), []byte(buffer))
		}
		bucket.Put([]byte(k), []byte(uri))
	}
	return nil
}
func (m *BigWigManager) LoadIndex(db *bolt.DB, k string) error {
	db.View(func(tx *bolt.Tx) error {
		bIdx := tx.Bucket([]byte("_idx"))
		b := tx.Bucket([]byte(m.dbname))
		v := b.Get([]byte(k))
		m.uriMap[k] = string(v)
		idx := bIdx.Get(v)
		reader, err := netio.NewReadSeeker(string(v))
		if err != nil {
			return err
		}
		bwf := bbi.NewBbiReader(reader)
		bwf.ReadIndex(bytes.NewReader(idx))
		bwr := bbi.NewBigWigReader(bwf)
		m.bwMap[string(k)] = bwr
		return nil
	})
	return nil
}
func (m *BigWigManager) UnloadIndex(k string) error {
	_, ok := m.uriMap[k]
	if ok {
		delete(m.uriMap, k)
		delete(m.bwMap, k)
	}
	return nil
}
func (m *BigWigManager) LoadIndexes(fn string) error {
	db, err := bolt.Open(fn, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
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
				return err
			}
			bwf := bbi.NewBbiReader(reader)
			bwf.ReadIndex(bytes.NewReader(idx))
			bwr := bbi.NewBigWigReader(bwf)
			m.bwMap[string(k)] = bwr
		}
		return nil
	})
	return nil
}
