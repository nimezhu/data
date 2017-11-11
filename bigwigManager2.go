package data

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	"github.com/nimezhu/indexed/bbi"
	"github.com/nimezhu/netio"
)

var (
	h1, _ = regexp.Compile("^http://")
	h2, _ = regexp.Compile("^https://")
)

/* 0 : local file
   1 : remote file without local index
   2 : remote file with local index
*/
func (m *BigWigManager2) checkUri(uri string) (string, int) {
	if h1.MatchString(uri) || h2.MatchString(uri) {
		var dir string
		if h1.MatchString(uri) {
			dir = strings.Replace(uri, "http://", "", 1)
		} else {
			dir = strings.Replace(uri, "https://", "", 1)
		}
		dir += ".index"
		fn := path.Join(m.indexRoot, dir)
		if _, err := os.Stat(fn); os.IsNotExist(err) {
			return fn, 1
		} else {
			return fn, 2
		}
	} else {
		return "", 0
	}
}

func (m *BigWigManager2) SaveIdx(uri string) (int, error) {
	fn, mode := m.checkUri(uri)
	if mode == 1 {
		reader, err := netio.NewReadSeeker(uri)
		if err != nil {
			return -1, err
		}
		bwf := bbi.NewBbiReader(reader)
		defer bwf.Close()
		os.MkdirAll(path.Dir(fn), 0700)
		f, err := os.Create(fn)
		if err != nil {
			log.Println("error in creating", err)
		}
		defer f.Close()
		err = bwf.WriteIndex(f)
		if err != nil {
			return -1, err
		}
	}
	return mode, nil
}
func (m *BigWigManager2) readBw(uri string) (*bbi.BigWigReader, error) {
	reader, err := netio.NewReadSeeker(uri)
	checkErr(err)
	bwf := bbi.NewBbiReader(reader)
	fn, mode := m.checkUri(uri)
	log.Println(fn, mode)
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
}

func (m *BigWigManager2) Add(key string, reader io.ReadSeeker, uri string) error {
	m.uriMap[key] = uri
	bwf := bbi.NewBbiReader(reader)
	fn, mode := m.checkUri(uri)
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
	log.Println("add uri", uri, key)
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
func (m *BigWigManager2) ServeTo(router *mux.Router) {
	prefix := "/" + m.dbname
	router.HandleFunc(prefix+"/ls", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		jsonHic, _ := json.Marshal(m.uriMap)
		w.Write(jsonHic)
	})
	AddBwsHandle(router, m.bwMap, prefix)
}

func NewBigWigManager2(uri string, dbname string, indexRoot string) *BigWigManager2 {
	//prefix := "/" + dbname
	uriMap := LoadURI(uri)
	bwmap := make(map[string]*bbi.BigWigReader)
	for k, v := range uriMap {
		bwmap[k] = readBw(v)
	}
	m := BigWigManager2{
		uriMap,
		bwmap,
		dbname,
		indexRoot,
	}
	//m.ServeTo(router)
	return &m
}

func InitBigWigManager2(dbname string, indexRoot string) *BigWigManager2 {
	uriMap := make(map[string]string)
	bwMap := make(map[string]*bbi.BigWigReader)
	m := BigWigManager2{
		uriMap,
		bwMap,
		dbname,
		indexRoot,
	}
	return &m
}
