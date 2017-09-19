package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	path "path/filepath"
	"strconv"

	"github.com/brentp/bix"
	"github.com/gorilla/mux"
	"github.com/nimezhu/data/image"
)

/*  Implement DataManager
type DataManager interface {
 AddURI(uri string, key string) error
 Del(string) error
 ServeTo(*mux.Router)
 List() []string
 Get(string) (string, bool)
 Move(key1 string, key2 string) bool
}
*/
/* File Directory
 * tabixdb.tsv		tabixFileId, tabixFileURI
 * tabixFiles.tsv.gz    // chr start end imageName
 * tabixFiles.tsv.gz.tbi
 * /images/*.png
 */
type TabixImageManager struct {
	uriMap  map[string]string
	dataMap map[string]*bix.Bix
	dbname  string
	root    string
}

func (T *TabixImageManager) AddURI(uri string, key string) error {
	b, err := bix.New(uri) //path to file now; not working for http, strict to current folder
	if err != nil {
		return err
	}
	T.uriMap[key] = uri
	T.dataMap[key] = b
	if T.root == "" {
		T.root = path.Dir(uri)
	}
	return nil
}
func (m *TabixImageManager) Del(key string) error {
	delete(m.uriMap, key)
	delete(m.dataMap, key)
	return nil
}
func (m *TabixImageManager) Get(key string) (string, bool) {
	v, ok := m.uriMap[key]
	return v, ok
}
func (m *TabixImageManager) Move(key1 string, key2 string) bool {
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

func (m *TabixImageManager) List() []string {
	keys := []string{}
	for k, _ := range m.uriMap {
		keys = append(keys, k)
	}
	return keys
}

func (T *TabixImageManager) ServeTo(router *mux.Router) {
	prefix := "/" + T.dbname
	router.HandleFunc(prefix+"/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		keys := []string{}
		for key, _ := range T.uriMap {
			keys = append(keys, key)
		}
		jsonHic, _ := json.Marshal(keys)
		w.Write(jsonHic)
	})
	router.HandleFunc(prefix+"/ls", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		jsonHic, _ := json.Marshal(T.uriMap)
		w.Write(jsonHic)
	})
	router.HandleFunc(prefix+"/{id}/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		params := mux.Vars(r)
		id := params["id"]
		a, ok := T.dataMap[id]
		if ok {
			jsonChr, _ := json.Marshal(a.Names())
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
		vals, _ := T.dataMap[id].Query(loc{chrom, s, e})
		i := 0
		for {
			v, err := vals.Next()
			if err != nil {
				break
			}
			//fmt.Println(v)
			i++
			io.WriteString(w, fmt.Sprint(v)) //TODO This In JavaScript For Link Images with Query Result.
			io.WriteString(w, "\n")
		}
	})
	image.AddTo(router, T.dbname+"/images", path.Join(T.root, "images")) //add images handler.
}
func InitTabixImageManager(dbname string) *TabixImageManager {
	uriMap := make(map[string]string) //NOT LoadURI. But Process Files or Load Tabix File
	dataMap := make(map[string]*bix.Bix)
	//dataList := []string{}

	//root := path.Dir(uri)
	m := TabixImageManager{
		uriMap,
		dataMap,
		dbname,
		"",
	}
	//m.ServeTo(router)
	return &m
}
func NewTabixImageManager(uri string, dbname string) *TabixImageManager {
	uriMap := LoadURI(uri) //NOT LoadURI. But Process Files or Load Tabix File
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
	root := path.Dir(uri)
	m := TabixImageManager{
		uriMap,
		dataMap,
		dbname,
		root,
	}
	//m.ServeTo(router)
	return &m
}
