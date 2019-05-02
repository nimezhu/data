package data

import (
	"io"

	"github.com/gorilla/mux"
)

type DataRouter interface {
	ServeTo(*mux.Router)
}
type DataLoader interface {
	Load(interface{}) error
}
type DataServer interface {
	DataRouter
	DataLoader
}
type DataManager interface {
	DataRouter
	AddURI(uri string, key string) error
	Del(string) error
	List() []string
	Get(string) (string, bool)
	Move(key1 string, key2 string) bool
}
type Manager interface {
	DataManager
	Add(key string, reader io.ReadSeeker, uri string) error
}
type Manager2 interface { //Version2  to add attrs
	Manager
	SetAttr(key string, values map[string]interface{}) error
	GetAttr(key string) (map[string]interface{}, bool)
}
type Entry struct {
	Name string
	URI  string
}

func IterEntry(d DataManager) chan Entry {
	ch := make(chan Entry)
	go func() {
		for _, k := range d.List() {
			v, ok := d.Get(k)
			if ok {
				ch <- Entry{k, v}
			}
		}
		close(ch)
	}()
	return ch
}
