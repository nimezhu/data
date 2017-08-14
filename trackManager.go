package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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

type TrackManager struct {
	id        string
	uriMap    map[string]string
	formatMap map[string]string
	managers  map[string]DataManager
}

func NewTrackManager(uri string, dbname string) *TrackManager {
	m := InitTrackManager(dbname)
	uriMap := LoadURI(uri)
	for k, v := range uriMap {
		m.AddURI(v, k)
	}
	return m
}
func InitTrackManager(dbname string) *TrackManager {
	uriMap := make(map[string]string)
	formatMap := make(map[string]string)
	dataMap := make(map[string]DataManager)
	m := TrackManager{
		dbname,
		uriMap,
		formatMap,
		dataMap,
	}
	return &m
}
func newManager(prefix string, format string) DataManager {
	if format == "bigwig" {
		return InitBigWigManager(prefix + ".bigwig")
	}
	if format == "bigbed" {
		return InitBigBedManager(prefix + ".bigbed")
	}
	if format == "hic" {
		return InitHicManager(prefix + ".hic")
	}
	return nil
}
func (m *TrackManager) AddURI(uri string, key string) error {
	format, _ := indexed.Magic(uri)
	if _, ok := m.managers[format]; !ok {
		m.managers[format] = newManager(m.id, format)
	}
	m.managers[format].AddURI(uri, key)
	m.formatMap[key] = format
	m.uriMap[key] = uri
	return nil
}

func (m *TrackManager) Del(k string) error {
	if uri, ok := m.uriMap[k]; ok {
		format, _ := indexed.Magic(uri)
		delete(m.uriMap, k)
		return m.managers[format].Del(k)
	}
	return errors.New("key not found")
}

func (m *TrackManager) ServeTo(router *mux.Router) {
	prefix := "/" + m.id
	router.HandleFunc(prefix+"/ls", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		jsonHic, _ := json.Marshal(m.uriMap)
		w.Write(jsonHic)
	})
	router.HandleFunc(prefix+"/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		a := make([]map[string]string, 0)
		for k, _ := range m.uriMap {
			a = append(a, map[string]string{"id": k, "format": m.formatMap[k]})
		}
		jsonHic, _ := json.Marshal(a)
		w.Write(jsonHic)
	})
	router.HandleFunc(prefix+"/{id}/{cmd:.*}", func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		cmd := params["cmd"]
		id := params["id"]
		format, _ := m.formatMap[id]
		fmt.Println("format", format, "id", id)
		//w.Write([]byte(code))
		//TODO redirect with format
		url := prefix + "." + format + "/" + id + "/" + cmd
		w.Header().Set("Access-Control-Allow-Origin", "*")
		http.Redirect(w, r, url, http.StatusPermanentRedirect)
	})
	for _, v := range m.managers {
		v.ServeTo(router)
	}
}

func (m *TrackManager) List() []string {
	s := make([]string, 0)
	for k, _ := range m.uriMap {
		s = append(s, k)
	}
	return s
}
func (m *TrackManager) Get(key string) (string, bool) {
	v, ok := m.uriMap[key]
	return v, ok
}
func (m *TrackManager) Move(key1 string, key2 string) bool {

	if v, ok := m.uriMap[key1]; ok {
		m.uriMap[key2] = v
		delete(m.uriMap, key1)
		format, _ := indexed.Magic(v)
		m.managers[format].Move(key1, key2)
		return true
	}
	return false
}
