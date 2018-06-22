package data

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type MapManager struct {
	name string
	data map[string]string
}

func (m *MapManager) AddURI(uri string, key string) error {
	m.data[key] = uri
	return nil
}
func (m *MapManager) Del(key string) error {
	delete(m.data, key)
	return nil
}
func (m *MapManager) List() []string {
	s := make([]string, 0)
	for k, _ := range m.data {
		s = append(s, k)
	}
	return s
}
func (m *MapManager) Get(key string) (string, bool) {
	v, ok := m.data[key]
	return v, ok
}
func (m *MapManager) Move(key1 string, key2 string) bool {
	if v, ok := m.data[key1]; ok {
		m.data[key2] = v
		delete(m.data, key1)
		return true
	}
	return false
}
func (m *MapManager) ServeTo(router *mux.Router) {
	prefix := "/" + m.name
	sub := router.PathPrefix(prefix).Subrouter()
	sub.HandleFunc("/ls", func(w http.ResponseWriter, r *http.Request) {

		jsonHic, _ := json.Marshal(m.data)
		w.Write(jsonHic)
	})
	sub.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {

		jsonHic, _ := json.Marshal(m.List())
		w.Write(jsonHic)
	})
	sub.HandleFunc("/get/{id}", func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		id := params["id"]
		content, ok := m.data[id]
		if ok {
			w.Write([]byte(content))
		} else {
			w.Write([]byte("can not find " + id))
		}
	})

}
func NewMapManager(uri string, name string) *MapManager {
	data := LoadURI(uri)
	//d := make(map[string]string)
	m := MapManager{
		name,
		data,
	}
	return &m
}

func InitMapManager(name string) *MapManager {
	data := make(map[string]string)
	m := MapManager{
		name,
		data,
	}
	return &m
}
