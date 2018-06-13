package data

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/nimezhu/indexed"
)

/*
type DataManager interface {
	AddURI(uri string, key string) error
	Del(string) error
	ServeTo(*mux.Router)
	List() []string
	Get(string) (string, bool)
	Move(key1 string, key2 string) bool
}
*/
/* TrackManager supports bigbed bigwig and hic , tabixImage */
type TrackManager2 struct {
	id        string
	uriMap    map[string]string
	formatMap map[string]string
	managers  map[string]Manager2
	root      string
}

func NewTrackManager2(uri string, dbname string, root string) *TrackManager2 {
	m := InitTrackManager2(dbname, root)
	uriMap := LoadURI(uri)
	for k, v := range uriMap {
		m.AddURI(v, k)
	}
	return m
}
func InitTrackManager2(dbname string, root string) *TrackManager2 {
	uriMap := make(map[string]string)
	formatMap := make(map[string]string)
	dataMap := make(map[string]Manager2)
	m := TrackManager2{
		dbname,
		uriMap,
		formatMap,
		dataMap,
		root,
	}
	return &m
}
func _newManager2(prefix string, format string, root string) Manager2 {
	if format == "bigwig" {
		return InitBigWigManager2(prefix+".bigwig", root)
	}
	if format == "bigbed" {
		return InitBigBedManager2(prefix+".bigbed", root)
	}
	if format == "bigbedLarge" {
		return InitBigBedManager2(prefix+".bigbedLarge", root)
	}
	if format == "hic" {
		return InitHicManager(prefix + ".hic")
	}
	/* obsoleted
	if format == "image" {
		return InitTabixImageManager(prefix + ".image")
	}
	*/
	return nil
}
func (m *TrackManager2) SetAttr(key string, values map[string]interface{}) error {
	if v, ok := m.managers[m.formatMap[key]]; ok {
		return v.SetAttr(key, values)
	}
	return errors.New("not found manager")
}
func (m *TrackManager2) GetAttr(key string) (map[string]interface{}, bool) {
	if _, ok := m.formatMap[key]; !ok {
		return nil, false
	}
	return m.managers[m.formatMap[key]].GetAttr(key)
}
func (m *TrackManager2) Add(key string, reader io.ReadSeeker, uri string) error {
	format, _ := indexed.MagicReadSeeker(reader)
	if format == "hic" || format == "bigwig" || format == "bigbed" {
		reader.Seek(0, 0)
		if _, ok := m.managers[format]; !ok {
			m.managers[format] = _newManager2(m.id, format, m.root)
		}
		m.managers[format].Add(key, reader, uri)
		m.formatMap[key] = format
		m.uriMap[key] = uri
	}
	return nil
}
func (m *TrackManager2) AddURI(uri string, key string) error {
	format, _ := indexed.Magic(uri)
	//HANDLE binindex in mem
	if format == "binindex" {
		m.uriMap[key] = strings.Replace(uri, "_format_:binindex:", "", 1)
		return nil
	}
	if _, ok := m.managers[format]; !ok {
		if _m := _newManager2(m.id, format, m.root); _m != nil {
			m.managers[format] = _m
			m.managers[format].AddURI(uri, key)
		} else {
			return errors.New("format not support yet")
		}
	} else {
		m.managers[format].AddURI(uri, key)
	}
	m.formatMap[key] = format
	m.uriMap[key] = uri
	return nil
}

func (m *TrackManager2) Del(k string) error {
	if uri, ok := m.uriMap[k]; ok {
		format, _ := indexed.Magic(uri)
		//TODO binindex
		delete(m.uriMap, k)
		return m.managers[format].Del(k)
	}
	return errors.New("key not found")
}

func (m *TrackManager2) ServeTo(router *mux.Router) {
	prefix := "/" + m.id
	router.HandleFunc(prefix+"/ls", func(w http.ResponseWriter, r *http.Request) {
		attr, ok := r.URL.Query()["attr"]
		
		if !ok || len(attr) < 1 || !(attr[0] == "1" || attr[0] == "true") {
			jsonHic, _ := json.Marshal(m.uriMap)
			w.Write(jsonHic)
		} else {
			//TODO PRECOMPUTING ??
			retv := make(map[string]map[string]interface{})
			for _, m0 := range m.managers {
				for _, k := range m0.List() {
					retv[k], _ = m0.GetAttr(k)
				}
			}
			jsonAttr, _ := json.Marshal(retv)
			w.Write(jsonAttr)

		}
	})
	router.HandleFunc(prefix+"/list", func(w http.ResponseWriter, r *http.Request) {
		
		a := make([]map[string]string, 0)
		//TODO fix this for trackAgent
		attr, ok := r.URL.Query()["attr"]
		sign := true
		if !ok || len(attr) < 1 || !(attr[0] == "1" || attr[0] == "true") {
			sign = false
		}

		for k, _ := range m.uriMap {
			a = append(a, map[string]string{"id": k, "format": m.formatMap[k]})
			if sign {
				if attrs, ok := m.managers[m.formatMap[k]].GetAttr(k); ok {
					for k0, v0 := range attrs {
						switch v0.(type) {
						case string:
							a[len(a)-1][k0] = v0.(string)
						default:

						}
					}
				}
			}
		}
		jsonHic, _ := json.Marshal(a)
		w.Write(jsonHic)
	})
	router.HandleFunc(prefix+"/{id}/{cmd:.*}", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		cmd := params["cmd"]
		id := params["id"]
		format, _ := m.formatMap[id]
		//w.Write([]byte(code))
		//TODO redirect with format
		if format == "binindex" { //binindex in memory
			uri, _ := m.uriMap[id] //name
			a1 := strings.Replace(r.URL.String(), prefix+"/", "", 1)
			a2 := strings.Replace(a1, id, uri, 1)
			//url := "/" + uri + "/" + cmd
			//fmt.Println(a2)
			//fmt.Println(url)
			url := "/" + a2
			
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		} else {
			url := prefix + "." + format + "/" + id + "/" + cmd
			
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		}
	})
	for _, v := range m.managers {
		v.ServeTo(router)
	}
}

func (m *TrackManager2) List() []string {
	s := make([]string, 0)
	for k, _ := range m.uriMap {
		s = append(s, k)
	}
	return s
}
func (m *TrackManager2) Get(key string) (string, bool) {
	v, ok := m.uriMap[key]
	return v, ok
}
func (m *TrackManager2) Move(key1 string, key2 string) bool { //TODO
	if v, ok := m.uriMap[key1]; ok {
		m.uriMap[key2] = v
		delete(m.uriMap, key1)
		format, _ := indexed.Magic(v) // TODO Fix this with ReadSeeker.
		m.managers[format].Move(key1, key2)
		return true
	}
	return false
}
