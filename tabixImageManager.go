package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	path "path/filepath"
	"strconv"
	"strings"

	"github.com/brentp/bix"
	"github.com/gorilla/mux"
	"github.com/nimezhu/data/image"
	"github.com/nimezhu/netio"
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
	uriMap         map[string]string
	dataMap        map[string]*bix.Bix
	imageToRegions map[string]string
	dbname         string
	root           string
}
type Bed4 struct {
	chr   string
	start int
	end   int
	name  string
}

func (b Bed4) Chr() string {
	return b.chr
}

func (b Bed4) Start() int {
	return b.start
}
func (b Bed4) End() int {
	return b.end
}
func (b Bed4) Id() string {
	return b.name
}

type ShortBed interface {
	Chr() string
	Start() int
	End() int
}

func regionText(b ShortBed) string {
	s := strconv.Itoa(b.Start())
	e := strconv.Itoa(b.End())
	return b.Chr() + ":" + s + "-" + e
}

//TODO
func regionsText(bs []Bed4) string { //TODO interface array function
	r := make([]string, len(bs))
	for i, v := range bs {
		r[i] = regionText(v)
	}
	return strings.Join(r, ",")
}
func bedsText(bs []Bed3) string { //TODO interface array function
	r := make([]string, len(bs))
	for i, v := range bs {
		r[i] = regionText(v)
	}
	return strings.Join(r, ",")
}

func (T *TabixImageManager) addBeds(uri string) error {
	f, err := netio.ReadAll(uri)
	if err != nil {
		return err
	}
	m := make(map[string][]Bed4)
	lines := strings.Split(string(f), "\n")
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		w := strings.Split(string(line), "\t")
		start, _ := strconv.Atoi(w[1])
		end, _ := strconv.Atoi(w[2])
		if _, ok := m[w[3]]; ok {
			m[w[3]] = append(m[w[3]], Bed4{w[0], start, end, w[3]})
		} else {
			m[w[3]] = []Bed4{Bed4{w[0], start, end, w[3]}}
		}
	}

	for k, v := range m {
		T.imageToRegions[k] = regionsText(v)
	}

	return nil
}
func (m *TabixImageManager) Add(key string, reader io.ReadSeeker, uri string) error {
	return errors.New("TODO")
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
	T.addBeds(uri)
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

		if _, ok := T.dataMap[id]; ok {
			//jsonChr, _ := json.Marshal(a.Names())
			//w.Write(jsonChr)
			w.Write([]byte("TODO CHRS"))
		} else {
			fmt.Println("can not find id", id)
		}
	})
	router.HandleFunc(prefix+"/getpos/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		params := mux.Vars(r)
		id := params["id"]
		if v, ok := T.imageToRegions[id]; ok {
			io.WriteString(w, v)
		} else {
			io.WriteString(w, "[]")
		}

	})
	router.HandleFunc(prefix+"/images/ls", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		io.WriteString(w, "id\tpos\n")
		for k, v := range T.imageToRegions {
			s := fmt.Sprintf("%s\t%s\n", k, v)
			io.WriteString(w, s)
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
		make(map[string]string),
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
		make(map[string]string),
		dbname,
		root,
	}
	//m.ServeTo(router)
	return &m
}
