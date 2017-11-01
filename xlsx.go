package data

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/nimezhu/netio"
	"github.com/tealeg/xlsx"
)

func logErr(err error) {
	if err != nil {
		log.Println(err)
	}
}
func AddSheets(uri string, router *mux.Router) map[string]DataManager {
	m := map[string]DataManager{}
	entry := []string{}
	jdata := []map[string]string{}
	httpP, _ := regexp.Compile("^http://")
	httpsP, _ := regexp.Compile("^https://")
	/* TODO */
	r, err := netio.NewReadSeeker(uri)
	logErr(err)
	bs, err := ioutil.ReadAll(r)
	logErr(err)
	xlFile, err := xlsx.OpenBinary(bs)
	logErr(err)
	config, ok := xlFile.Sheet["Config"] //get root
	if !ok {

	}
	cfgRows := config.Rows
	var root string
	for _, row := range cfgRows[1:] {
		x := row.Cells
		if x[0].String() == "root" {
			root = x[1].String()
		}
	}
	index, ok := xlFile.Sheet["Index"]
	if !ok {

	}
	idxRows := index.Rows
	type item struct {
		Id       string `xlsx:"0"`
		Type     string `xlsx:"1"`
		Nc       int    `xlsx:"2"`
		Vc       int    `xlsx:"3"`
		Preserve bool   `xlsx:"4"`
	}
	for _, row := range idxRows[1:] {
		a := &item{}
		row.ReadStruct(a)
		fmt.Println(a)
		if len(a.Id) == 0 {
			break
		}
		if a.Id[0] == '#' {
			continue
		}
		if vsheet, ok := xlFile.Sheet[a.Id]; ok {
			format := a.Type
			k := a.Id
			fmt.Println("init", k, format)
			m0 := initDataManager(k, format) //TODO
			jdata = append(jdata, map[string]string{
				"dbname": k,
				"uri":    uri + ":" + k,
				"format": format,
			})
			entry = append(entry, k)
			if a.Type == "map" {
				for _, r := range vsheet.Rows[1:] {
					v0 := r.Cells[a.Vc-1].String()
					k0 := r.Cells[a.Nc-1].String()
					m0.AddURI(v0, k0)
				}
			} else {
				for _, r := range vsheet.Rows[1:] {
					id := r.Cells[a.Nc-1].String()
					loc := r.Cells[a.Vc-1].String()
					var uri string
					if httpP.MatchString(loc) || httpsP.MatchString(loc) {
						uri = loc
						m0.AddURI(uri, id)
					} else {
						uri = path.Join(root, loc) //TODO
						if _, err := os.Stat(uri); err == nil {
							m0.AddURI(uri, id)
						} else {
							log.Println("WARNING!!! cannot reading", uri, id)
						}
					}

				}
			}
			fmt.Println(m0.List())
			m0.ServeTo(router)
			m[k] = m0
		}

	}
	router.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		e, _ := json.Marshal(entry)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(e)
	})
	router.HandleFunc("/ls", func(w http.ResponseWriter, r *http.Request) {
		e, _ := json.Marshal(jdata)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(e)
	})
	return m
}
