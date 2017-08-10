package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nimezhu/indexed/bbi"
	"github.com/nimezhu/netio"
)

/*  Implement
type DataManager interface {
 AddURI(uri string, key string) error
 Del(string) error
 ServeTo(*mux.Router)
 List() []string
 Get(string) (string, bool)
 Move(key1 string, key2 string) bool
}
*/

type BigBedManager struct {
	uriMap  map[string]string
	dataMap map[string]*bbi.BigBedReader
	dbname  string
}

func (bb *BigBedManager) AddURI(uri string, key string) error {
	reader, err := netio.NewReadSeeker(uri)
	if err != nil {
		return err
	}
	bwf := bbi.NewBbiReader(reader)
	err = bwf.InitIndex()
	if err != nil {
		return err
	}
	br := bbi.NewBigBedReader(bwf)
	bb.dataMap[key] = br
	bb.uriMap[key] = uri
	return nil
}

func (m *BigBedManager) Del(key string) error {
	delete(m.uriMap, key)
	delete(m.dataMap, key)
	return nil
}
func (m *BigBedManager) Get(key string) (string, bool) {
	v, ok := m.uriMap[key]
	return v, ok
}
func (m *BigBedManager) Move(key1 string, key2 string) bool {
	v, ok1 := m.uriMap[key1]
	d, ok2 := m.dataMap[key1]
	if ok1 && ok2 {
		m.uriMap[key2] = v
		m.dataMap[key2] = d
		delete(m.uriMap, key1)
		delete(m.dataMap, key1)
	}
	return ok1 && ok2
}

func (m *BigBedManager) List() []string {
	keys := []string{}
	for k, _ := range m.uriMap {
		keys = append(keys, k)
	}
	return keys
}

func (m *BigBedManager) ServeTo(router *mux.Router) {
	prefix := "/" + m.dbname
	router.HandleFunc(prefix+"/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		keys := []string{}
		for key, _ := range m.uriMap {
			keys = append(keys, key)
		}
		jsonHic, _ := json.Marshal(keys)
		w.Write(jsonHic)
	})
	router.HandleFunc(prefix+"/ls", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		jsonHic, _ := json.Marshal(m.uriMap)
		w.Write(jsonHic)
	})

	router.HandleFunc(prefix+"/{id}/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		params := mux.Vars(r)
		id := params["id"]
		a, ok := m.dataMap[id]
		if ok {
			jsonChr, _ := json.Marshal(a.Genome.Chrs)
			w.Write(jsonChr)
		} else {
			fmt.Println("can not find id", id)
		}
	})

	router.HandleFunc(prefix+"/{id}/get/{chr}:{start}-{end}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		params := mux.Vars(r)
		id := params["id"]
		chrom := params["chr"]
		s, _ := strconv.Atoi(params["start"])
		e, _ := strconv.Atoi(params["end"])
		vals, _ := m.dataMap[id].QueryRaw(chrom, s, e)
		for v := range vals {
			io.WriteString(w, m.dataMap[id].Format(v))
			io.WriteString(w, "\n")
		}
	})

}

func NewBigBedManager(uri string, dbname string) *BigBedManager {
	uriMap := LoadURI(uri)
	dataMap := make(map[string]*bbi.BigBedReader)
	//dataList := []string{}
	m := BigBedManager{
		uriMap,
		dataMap,
		dbname,
	}
	for k, v := range uriMap {
		m.AddURI(v, k)
		//dataList = append(dataList, k)
	}

	//m.ServeTo(router)
	return &m
}

func InitBigBedManager(dbname string) *BigBedManager {
	uriMap := make(map[string]string)
	dataMap := make(map[string]*bbi.BigBedReader)
	m := BigBedManager{
		uriMap,
		dataMap,
		dbname,
	}
	return &m
}
