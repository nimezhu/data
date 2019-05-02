package data

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nimezhu/indexed/bbi"
	"github.com/nimezhu/netio"
)

func readBw(uri string) *bbi.BigWigReader {
	reader, err := netio.NewReadSeeker(uri)
	checkErr(err)
	bwf := bbi.NewBbiReader(reader)
	bwf.InitIndex()
	//log.Println("in reading idx of", uri)
	bw := bbi.NewBigWigReader(bwf)
	return bw
}

/*BigWigManager implement DataManager Inteface */

type BigWigManager struct {
	uriMap map[string]string
	bwMap  map[string]*bbi.BigWigReader
	dbname string
}

func (m *BigWigManager) Add(key string, reader io.ReadSeeker, uri string) error {
	m.uriMap[key] = uri
	log.Println("adding in bigwig", key, uri)
	bwf := bbi.NewBbiReader(reader)
	bwf.InitIndex()
	//log.Println("in reading idx of", uri)
	bw := bbi.NewBigWigReader(bwf)
	m.bwMap[key] = bw
	return nil
}
func (m *BigWigManager) AddURI(uri string, key string) error {
	m.uriMap[key] = uri
	m.bwMap[key] = readBw(uri)
	log.Println("add uri", uri, key)
	return nil
}
func (m *BigWigManager) Del(key string) error {
	delete(m.uriMap, key)
	delete(m.bwMap, key)
	return nil
}
func (m *BigWigManager) Get(key string) (string, bool) {
	v, ok := m.uriMap[key]
	return v, ok
}
func (m *BigWigManager) Move(key1 string, key2 string) bool {
	v, ok1 := m.uriMap[key1]
	d, ok2 := m.bwMap[key1]
	if ok1 && ok2 {
		m.uriMap[key2] = v
		m.bwMap[key2] = d
		delete(m.uriMap, key1)
		delete(m.bwMap, key1)
	}
	return ok1 && ok2
}
func (m *BigWigManager) List() []string {
	keys := []string{}
	for k, _ := range m.uriMap {
		keys = append(keys, k)
	}
	return keys
}
func (m *BigWigManager) ServeTo(router *mux.Router) {
	prefix := "/" + m.dbname
	sub := router.PathPrefix(prefix).Subrouter()
	sub.HandleFunc("/ls", func(w http.ResponseWriter, r *http.Request) {

		jsonHic, _ := json.Marshal(m.uriMap)
		w.Write(jsonHic)
	})
	addBwsHandle(sub, m.bwMap)
}

func NewBigWigManager(uri string, dbname string) *BigWigManager {
	//prefix := "/" + dbname
	uriMap := LoadURI(uri)
	bwmap := make(map[string]*bbi.BigWigReader)
	for k, v := range uriMap {
		bwmap[k] = readBw(v)
	}
	m := BigWigManager{
		uriMap,
		bwmap,
		dbname,
	}
	//m.ServeTo(router)
	return &m
}

func InitBigWigManager(dbname string) *BigWigManager {
	uriMap := make(map[string]string)
	bwMap := make(map[string]*bbi.BigWigReader)
	m := BigWigManager{
		uriMap,
		bwMap,
		dbname,
	}
	return &m
}
