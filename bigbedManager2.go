package data

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

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

type BigBedManager2 struct {
	uriMap    map[string]string
	dataMap   map[string]*bbi.BigBedReader
	dbname    string
	indexRoot string
}

func (m *BigBedManager2) checkUri(uri string) (string, int) {
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
func (m *BigBedManager2) readBw(uri string) (*bbi.BigBedReader, error) {
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
	bw := bbi.NewBigBedReader(bwf)
	return bw, nil
}

func (m *BigBedManager2) Add(key string, reader io.ReadSeeker, uri string) error {
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
	bw := bbi.NewBigBedReader(bwf)
	m.dataMap[key] = bw
	return nil
}
func (bb *BigBedManager2) AddURI(uri string, key string) error {
	bb.uriMap[key] = uri
	var err error
	bb.dataMap[key], err = bb.readBw(uri)
	/*
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
	*/
	return err
}

func (m *BigBedManager2) Del(key string) error {
	delete(m.uriMap, key)
	delete(m.dataMap, key)
	return nil
}
func (m *BigBedManager2) Get(key string) (string, bool) {
	v, ok := m.uriMap[key]
	return v, ok
}
func (m *BigBedManager2) Move(key1 string, key2 string) bool {
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

func (m *BigBedManager2) List() []string {
	keys := []string{}
	for k, _ := range m.uriMap {
		keys = append(keys, k)
	}
	return keys
}

func (m *BigBedManager2) ServeTo(router *mux.Router) {
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

	router.HandleFunc(prefix+"/{id}/getbin/{chr}:{start}-{end}/{binsize}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
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

func NewBigBedManager2(uri string, dbname string, root string) *BigBedManager2 {
	uriMap := LoadURI(uri)
	dataMap := make(map[string]*bbi.BigBedReader)
	//dataList := []string{}
	m := BigBedManager2{
		uriMap,
		dataMap,
		dbname,
		root,
	}
	for k, v := range uriMap {
		m.AddURI(v, k)
		//dataList = append(dataList, k)
	}

	//m.ServeTo(router)
	return &m
}

func InitBigBedManager2(dbname string, root string) *BigBedManager2 {
	uriMap := make(map[string]string)
	dataMap := make(map[string]*bbi.BigBedReader)
	m := BigBedManager2{
		uriMap,
		dataMap,
		dbname,
		root,
	}
	return &m
}