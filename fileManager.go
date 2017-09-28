package data

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nimezhu/netio"
)

type FileManager struct {
	uri map[string]string
	//data   map[string]io.ReadSeeker
	dbname    string
	bufferMap map[string][]byte
}

/*
func (m *FileManager) Add(key string, reader io.ReadSeeker, uri string) error {
	_, ok := m.uri[key]
	if ok {
		return errors.New("duplicated key string")
	}
	m.uri[key] = uri
}
*/
func (m *FileManager) AddURI(uri string, key string) error {
	_, ok := m.uri[key]
	if ok {
		return errors.New("duplicated key string")
	}
	/*
		  f, err := netio.NewReadSeeker(uri)
			if err != nil {
				return err
			}
	*/
	m.uri[key] = uri
	//m.data[key] = f
	return nil
}

func (m *FileManager) Del(key string) error {
	_, ok := m.uri[key]
	if !ok {
		return errors.New("key not found")
	}
	//delete(m.data, key)
	delete(m.uri, key)

	_, ok2 := m.bufferMap[key]
	if ok2 {
		delete(m.bufferMap, key)
	}
	//delete(m.data, key)
	return nil
}

func (m *FileManager) ServeTo(router *mux.Router) {
	//TODO File Handler
	prefix := "/" + m.dbname
	m.initBuffersHandle(router) //TODO change buffermap into m.
	router.HandleFunc(prefix+"/ls", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		jsonHic, _ := json.Marshal(m.uri)
		w.Write(jsonHic)
	})
}

func (m *FileManager) List() []string {
	keys := []string{}
	for k, _ := range m.uri {
		keys = append(keys, k)
	}
	return keys
}
func (m *FileManager) Get(key string) (string, bool) {
	v, ok := m.uri[key]
	return v, ok
}
func (m *FileManager) Move(key1 string, key2 string) bool {
	v, ok := m.uri[key1]
	if ok {
		m.uri[key2] = v
		delete(m.uri, key1)
	}
	b, ok2 := m.bufferMap[key1]
	if ok2 {
		m.bufferMap[key2] = b
		delete(m.bufferMap, key1)
	}
	return ok
}

func (m *FileManager) initBuffersHandle(router *mux.Router) {
	//bufferMap := make(map[string][]byte)
	prefix := "/" + m.dbname
	router.HandleFunc(prefix+"/list", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		keys := []string{}
		for key, _ := range m.uri {
			keys = append(keys, key)
		}
		jsonBuffers, _ := json.Marshal(keys)
		w.Write(jsonBuffers)
	})
	router.HandleFunc(prefix+"/get/{id}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		params := mux.Vars(r)
		id := params["id"]
		content, ok := m.bufferMap[id]
		if ok {
			w.Write(content)
		} else {
			//w.Write([]byte("can not find " + id))
			f, ok := m.uri[id]
			if ok {
				r, err := netio.NewReadSeeker(f)
				if err != nil {
					log.Println("error load " + id)
				} else {
					m.bufferMap[id], _ = ioutil.ReadAll(r)
					w.Write(m.bufferMap[id])
				}
			} else {
				w.Write([]byte("file not found"))
			}
		}
	})

}

func NewFileManager(uri string, dbname string) *FileManager {
	uriMap := LoadURI(uri)
	bufferMap := make(map[string][]byte)
	m := FileManager{
		uriMap,
		dbname,
		bufferMap,
	}
	return &m
}

func InitFileManager(dbname string) *FileManager {
	uriMap := make(map[string]string)
	bufferMap := make(map[string][]byte)
	m := FileManager{
		uriMap,
		dbname,
		bufferMap,
	}
	return &m
}
