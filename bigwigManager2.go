package data

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gorilla/mux"
	"github.com/nimezhu/indexed/bbi"
	"github.com/nimezhu/netio"
)

/*
func (m *BigWigManager2) checkUri(uri string) (string, int) {
	return checkUri(uri, m.indexRoot)
}


func (m *BigWigManager2) SaveIdx(uri string) (int, error) {
	return saveIdx(uri, m.indexRoot)
}
*/

func (m *BigWigManager2) readBw(uri string) (*bbi.BigWigReader, error) {
	reader, err := netio.NewReadSeeker(uri)
	checkErr(err)
	bwf := bbi.NewBbiReader(reader)
	fn, mode := checkUri(uri, m.indexRoot)
	log.Println("load", uri, mode)
	if mode == 0 {
		bwf.InitIndex()
	} else if mode == 1 {
		bwf.InitIndex()
		go func() {
			os.MkdirAll(path.Dir(fn), 0700)
			f, err := os.Create(fn)
			if err != nil {
				log.Println("error in creating", err)
			}
			defer f.Close()
			err = bwf.WriteIndex(f)
			checkErr(err)
		}()
	} else if mode == 2 {
		f, err := os.Open(fn)
		defer f.Close()
		if err != nil {
			return nil, err
		}
		err = bwf.ReadIndex(f)
		if err != nil {
			return nil, err
		}
	}
	bw := bbi.NewBigWigReader(bwf)
	return bw, nil
}

/* BigWigManager2 implement DataManager Inteface
 * with buffered system
 */
type BigWigManager2 struct {
	uriMap    map[string]string
	bwMap     map[string]*bbi.BigWigReader
	dbname    string
	indexRoot string
	valueMap  map[string]map[string]interface{} //LONG LABELS
}

func (m *BigWigManager2) SetAttr(key string, value map[string]interface{}) error {
	m.valueMap[key] = value
	return nil
}
func (m *BigWigManager2) GetAttr(key string) (map[string]interface{}, bool) {
	v, ok := m.valueMap[key]
	return v, ok
}
func (m *BigWigManager2) Add(key string, reader io.ReadSeeker, uri string) error {
	m.uriMap[key] = uri
	bwf := bbi.NewBbiReader(reader)
	fn, mode := checkUri(uri, m.indexRoot)
	if mode == 0 {
		bwf.InitIndex()
	} else if mode == 1 {
		f, err := os.Open(fn)
		defer f.Close()
		if err != nil {
			return err
		}
		bwf.InitIndex()
		go func() { bwf.WriteIndex(f) }()
	} else if mode == 2 {
		f, err := os.Open(fn)
		defer f.Close()
		if err != nil {
			return err
		}
		err = bwf.ReadIndex(f)
		if err != nil {
			return err
		}
	}
	bw := bbi.NewBigWigReader(bwf)
	m.bwMap[key] = bw
	return nil
}
func (m *BigWigManager2) AddURI(uri string, key string) error {
	m.uriMap[key] = uri
	var err error
	m.bwMap[key], err = m.readBw(uri)
	return err
}
func (m *BigWigManager2) Del(key string) error {
	delete(m.uriMap, key)
	delete(m.bwMap, key)
	return nil
}
func (m *BigWigManager2) Get(key string) (string, bool) {
	v, ok := m.uriMap[key]
	return v, ok
}
func (m *BigWigManager2) Move(key1 string, key2 string) bool {
	v, ok1 := m.uriMap[key1]
	d, ok2 := m.bwMap[key1]
	if ok1 && ok2 {
		m.uriMap[key2] = v
		m.bwMap[key2] = d
		delete(m.uriMap, key1)
		delete(m.bwMap, key1)
		delete(m.valueMap, key1)
	}
	return ok1 && ok2
}
func (m *BigWigManager2) List() []string {
	keys := []string{}
	for k, _ := range m.uriMap {
		keys = append(keys, k)
	}
	return keys
}

/* TODO ls Add Attr */
func (m *BigWigManager2) ServeTo(router *mux.Router) {
	prefix := "/" + m.dbname
	sub := router.PathPrefix(prefix).Subrouter()
	sub.HandleFunc("/ls", func(w http.ResponseWriter, r *http.Request) {
		attr, ok := r.URL.Query()["attr"]

		if !ok || len(attr) < 1 || !(attr[0] == "1" || attr[0] == "true") {
			jsonHic, _ := json.Marshal(m.uriMap)
			w.Write(jsonHic)
		} else {
			jsonAttr, _ := json.Marshal(m.valueMap)
			w.Write(jsonAttr)
		}
		//not only uriMap ... but also attrs.

	})
	AddBwsHandle(sub, m.bwMap)
}

func NewBigWigManager2(uri string, dbname string, indexRoot string) *BigWigManager2 {
	//prefix := "/" + dbname
	uriMap := LoadURI(uri)
	bwmap := make(map[string]*bbi.BigWigReader)
	valueMap := make(map[string]map[string]interface{})
	for k, v := range uriMap {
		bwmap[k] = readBw(v)
	}
	m := BigWigManager2{
		uriMap,
		bwmap,
		dbname,
		indexRoot,
		valueMap,
	}
	//m.ServeTo(router)
	return &m
}

func InitBigWigManager2(dbname string, indexRoot string) *BigWigManager2 {
	uriMap := make(map[string]string)
	bwMap := make(map[string]*bbi.BigWigReader)
	valueMap := make(map[string]map[string]interface{})
	m := BigWigManager2{
		uriMap,
		bwMap,
		dbname,
		indexRoot,
		valueMap,
	}
	return &m
}
