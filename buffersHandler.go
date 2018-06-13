package data

import (
	"encoding/json"
	"log"
	//"fmt"
	"io/ioutil"

	"net/http"

	"github.com/gorilla/mux"
	"github.com/nimezhu/netio"
)

func AddBuffersHandle(router *mux.Router, uriMap map[string]string, prefix string) {
	bufferMap := make(map[string][]byte)
	for k, v := range uriMap {
		r, err := netio.NewReadSeeker(v)
		if err != nil {
			log.Println("error load " + k)
		} else {
			bufferMap[k], _ = ioutil.ReadAll(r)
		}
	}
	router.HandleFunc(prefix+"/list", func(w http.ResponseWriter, r *http.Request) {
		
		keys := []string{}
		for key, _ := range bufferMap {
			keys = append(keys, key)
		}
		jsonBuffers, _ := json.Marshal(keys)
		w.Write(jsonBuffers)
	})
	router.HandleFunc(prefix+"/get/{id}", func(w http.ResponseWriter, r *http.Request) {
		
		params := mux.Vars(r)
		id := params["id"]
		content, ok := bufferMap[id]
		if ok {
			w.Write(content)
		} else {
			w.Write([]byte("can not find " + id))
		}
	})

}
