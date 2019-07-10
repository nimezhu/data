package data

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	path "path/filepath"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nimezhu/indexed/bbi"
	"github.com/nimezhu/netio"
)

type BigBedManager struct {
	uriMap    map[string]string
	dataMap   map[string]*bbi.BigBedReader
	dbname    string
	indexRoot string
	valueMap  map[string]map[string]interface{}
}

func (m *BigBedManager) readBw(uri string) (*bbi.BigBedReader, error) {
	reader, err := netio.NewReadSeeker(uri)
	checkErr(err)
	bwf := bbi.NewBbiReader(reader)
	fn, mode := checkUri(uri, m.indexRoot)
	//log.Println("  Load entry", uri, mode)
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
	bw := bbi.NewBigBedReader(bwf)
	return bw, nil
}
func (m *BigBedManager) SetAttr(key string, value map[string]interface{}) error {
	m.valueMap[key] = value
	return nil
}
func (m *BigBedManager) GetAttr(key string) (map[string]interface{}, bool) {
	v, ok := m.valueMap[key]
	return v, ok
}
func (m *BigBedManager) Add(key string, reader io.ReadSeeker, uri string) error {
	m.uriMap[key] = uri
	bwf := bbi.NewBbiReader(reader)
	fn, mode := checkUri(uri, m.indexRoot)
	log.Println("    Loading entry", key)
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
	bw := bbi.NewBigBedReader(bwf)
	m.dataMap[key] = bw
	return nil
}
func (bb *BigBedManager) AddURI(uri string, key string) error {
	bb.uriMap[key] = uri
	var err error
	log.Println("    Loading entry", key)
	bb.dataMap[key], err = bb.readBw(uri)
	return err
}

func (m *BigBedManager) Del(key string) error {
	delete(m.uriMap, key)
	delete(m.dataMap, key)
	delete(m.valueMap, key)
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
	sub := router.PathPrefix(prefix).Subrouter()
	sub.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {

		keys := []string{}
		for key, _ := range m.uriMap {
			keys = append(keys, key)
		}
		jsonHic, _ := json.Marshal(keys)
		w.Write(jsonHic)
	})
	sub.HandleFunc("/ls", func(w http.ResponseWriter, r *http.Request) {
		attr, ok := r.URL.Query()["attr"]

		if !ok || len(attr) < 1 || !(attr[0] == "1" || attr[0] == "true") {
			jsonHic, _ := json.Marshal(m.uriMap)
			w.Write(jsonHic)
		} else {
			jsonAttr, _ := json.Marshal(m.valueMap)
			w.Write(jsonAttr)
		}
	})

	sub.HandleFunc("/{id}/list", func(w http.ResponseWriter, r *http.Request) {

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

	sub.HandleFunc("/{id}/get/{chr}:{start}-{end}", func(w http.ResponseWriter, r *http.Request) {

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

	sub.HandleFunc("/{id}/getbin/{chr}:{start}-{end}/{binsize}", func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		id := params["id"]
		chr := params["chr"]
		start, _ := strconv.Atoi(params["start"])
		end, _ := strconv.Atoi(params["end"])
		binsize, _ := strconv.Atoi(params["binsize"])
		bw, ok := m.dataMap[id]
		if !ok {
			io.WriteString(w, id+" not found")
		} else {
			arr := []*bbi.BedBbiQueryType{}
			if iter, err := bw.QueryBin(chr, start, end, binsize); err == nil {
				for i := range iter {
					/*
						if i.To == 0 {
							break
						} //TODO DEBUG THIS FOR STREAM
					*/
					arr = append(arr, i)
					//io.WriteString(w, fmt.Sprintln(i.From, "\t", i.To, "\t", i.Sum))
				}
			}
			j, err := json.Marshal(arr)
			checkErr(err)
			io.WriteString(w, string(j))
		}
	})

}

func NewBigBedManager(uri string, dbname string, root string) *BigBedManager {
	uriMap := loadURI(uri)
	dataMap := make(map[string]*bbi.BigBedReader)
	valueMap := make(map[string]map[string]interface{})
	//dataList := []string{}
	m := BigBedManager{
		uriMap,
		dataMap,
		dbname,
		root,
		valueMap,
	}
	for k, v := range uriMap {
		m.AddURI(v, k)
		//dataList = append(dataList, k)
	}

	//m.ServeTo(router)
	return &m
}

func InitBigBedManager(dbname string, root string) *BigBedManager {
	uriMap := make(map[string]string)
	dataMap := make(map[string]*bbi.BigBedReader)
	valueMap := make(map[string]map[string]interface{})
	m := BigBedManager{
		uriMap,
		dataMap,
		dbname,
		root,
		valueMap,
	}
	return &m
}
