package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/brentp/bix"
	"github.com/gorilla/mux"
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

type TabixManager struct {
	uriMap  map[string]string
	dataMap map[string]*bix.Bix
	dbname  string
}

type loc struct {
	chrom string
	start int
	end   int
}

func (s loc) Chrom() string {
	return s.chrom
}
func (s loc) Start() uint32 {
	return uint32(s.start)
}
func (s loc) End() uint32 {
	return uint32(s.end)
}

func (T *TabixManager) AddURI(uri string, key string) error {
	b, err := bix.New(uri) //path to file now; not working for http
	if err != nil {
		return err
	}
	T.uriMap[key] = uri
	T.dataMap[key] = b
	return nil
}
func (m *TabixManager) Del(key string) error {
	delete(m.uriMap, key)
	delete(m.dataMap, key)
	return nil
}
func (m *TabixManager) Get(key string) (string, bool) {
	v, ok := m.uriMap[key]
	return v, ok
}
func (m *TabixManager) Move(key1 string, key2 string) bool {
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

func (m *TabixManager) List() []string {
	keys := []string{}
	for k, _ := range m.uriMap {
		keys = append(keys, k)
	}
	return keys
}

func (T *TabixManager) ServeTo(router *mux.Router) {
	prefix := "/" + T.dbname
	router.HandleFunc(prefix+"/list", func(w http.ResponseWriter, r *http.Request) {
		
		keys := []string{}
		for key, _ := range T.uriMap {
			keys = append(keys, key)
		}
		jsonHic, _ := json.Marshal(keys)
		w.Write(jsonHic)
	})
	router.HandleFunc(prefix+"/ls", func(w http.ResponseWriter, r *http.Request) {
		
		jsonHic, _ := json.Marshal(T.uriMap)
		w.Write(jsonHic)
	})
	router.HandleFunc(prefix+"/{id}/list", func(w http.ResponseWriter, r *http.Request) {
		
		params := mux.Vars(r)
		id := params["id"]
		if _, ok := T.dataMap[id]; ok {
			//jsonChr, _ := json.Marshal(a.Names()) //TODO FIX
			//w.Write(jsonChr)
			w.Write([]byte("TODO CHRS"))
		} else {
			fmt.Println("can not find id", id)
		}
	})
	router.HandleFunc(prefix+"/{id}/get/{chr}:{start}-{end}", func(w http.ResponseWriter, r *http.Request) {
		
		params := mux.Vars(r)
		id := params["id"]
		chrom := params["chr"]
		s, _ := strconv.Atoi(params["start"])
		e, _ := strconv.Atoi(params["end"])
		vals, _ := T.dataMap[id].Query(loc{chrom, s, e})
		i := 0
		for {
			v, err := vals.Next()
			if err != nil {
				break
			}
			//fmt.Println(v)
			i++
			io.WriteString(w, fmt.Sprint(v))
			io.WriteString(w, "\n")
		}
	})
}

func NewTabixManager(uri string, dbname string) *TabixManager {
	uriMap := LoadURI(uri)
	dataMap := make(map[string]*bix.Bix)
	//dataList := []string{}
	for k, v := range uriMap {
		b, err := bix.New(v)
		if err == nil {
			dataMap[k] = b
			fmt.Println("loading ", k)
		} else {
			panic(err)
		}
		//dataList = append(dataList, k)
	}
	m := TabixManager{
		uriMap,
		dataMap,
		dbname,
	}
	//m.ServeTo(router)
	return &m
}
