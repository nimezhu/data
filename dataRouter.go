package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	path "path/filepath"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/nimezhu/indexed"
)

type dataIndex struct {
	genome string      `json:"genome"`
	dbname string      `json:"dbname"`
	data   interface{} // map[string]string or map[string][]string? could be uri or more sofisticated data structure such as binindex image
	format string      `json:"format"`
}

func (m *Loader) AddDataMiddleware(uri string, h http.Handler) http.Handler {
	router := mux.NewRouter()
	m.Load(uri, router)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		match := &mux.RouteMatch{}
		if router.Match(r, match) {
			router.ServeHTTP(w, r)
		} else {
			h.ServeHTTP(w, r)
		}
	})
}
func (m *Loader) Load(uri string, router *mux.Router) error {
	return m.loadIndexURI(uri, router)
}
func (m *Loader) Reload(uri string) error {
	return m.refreshIndexURI(uri)
}

func (m *Loader) loadIndexURI(uri string, router *mux.Router) error {
	d, err := m.smartParseURI(uri)
	if err != nil {
		return err
	}
	err = m.loadIndexesTo(d, router)
	return err
}

func (m *Loader) refreshIndexURI(uri string) error {
	d, err := m.smartParseURI(uri)
	if err != nil {
		return err
	}
	for _, v := range d {
		if v.format == "track" {
			m.Refresh(v.genome+"/"+v.dbname, v.data, v.format)
		}
	}
	return err
}

func SaveIndex(uri string, root string) error {
	h1, _ := regexp.Compile("^http://")
	h2, _ := regexp.Compile("^https://")
	m := NewLoader(root)
	d, err := m.smartParseURI(uri)
	if err != nil {
		return err
	}
	for _, v := range d {
		if v.format == "bigwig" || v.format == "bigbed" || v.format == "track" {
			switch t := v.data.(type) {
			default:
				log.Printf("unexpected type %T\n", t)
				return errors.New(fmt.Sprintf("bigwig format not support type %T", t))
			case string:
				return errors.New("Not support text input yet")
			case map[string]interface{}:
				for _, val := range v.data.(map[string]interface{}) {
					if !h1.MatchString(val.(string)) && !h2.MatchString(val.(string)) {
						continue
					}
					format, _ := indexed.Magic(val.(string))
					if format == "bigwig" || format == "bigbed" {
						log.Println("    Save index of ", val.(string))
						saveIdx(val.(string), root)
					}
				}
			}
		}
	}
	return nil
}

func smartCheckSheetHeader(uri string) {

}

func (m *Loader) smartParseURI(uri string) ([]dataIndex, error) {
	http, _ := regexp.Compile("^http://")
	https, _ := regexp.Compile("^https://")
	if len(uri) == len("1DEvA94QkN1KZQT51IYOOcIvGL2Ux7Qwqe5IpE9Pe1N8") { //GOOGLE SHEET ID ?
		if _, err := os.Stat(uri); os.IsNotExist(err) {
			if !http.MatchString(uri) && !https.MatchString(uri) {
				root := path.Dir(m.IndexRoot)
				return parseGSheet(uri, root)
			}
		}
	} else {
		ext := path.Ext(uri)
		if ext == ".xlsx" || ext == ".xls" {
			return parseXls(uri)
		} else {
			return nil, errors.New("not recognize format")
		}
	}
	return nil, errors.New("not recognize uri")
}

func trans(v dataIndex) map[string]string {
	switch v.data.(type) {
	case string:
		return map[string]string{
			"genome": v.genome,
			"dbname": v.dbname,
			"uri":    v.data.(string),
			"format": v.format,
		}
	default:
		return map[string]string{
			"genome": v.genome,
			"dbname": v.dbname,
			"uri":    "null",
			"format": v.format,
		}
	}
	return nil
}

func (m *Loader) loadIndexesTo(indexes []dataIndex, router *mux.Router) error {
	fail := 0
	gs := map[string][]dataIndex{}
	for _, v := range indexes {
		if v.genome != "all" {
			if _, ok := gs[v.genome]; !ok {
				gs[v.genome] = []dataIndex{}
			}
			gs[v.genome] = append(gs[v.genome], v)
		} else {
			err := m.loadIndex(v, router)
			if err != nil {
				fail += 1
			} else {
				m.jdata = append(m.jdata, trans(v))
				m.entry = append(m.entry, v.dbname)
			}
		}
	}
	for g, v := range gs {
		log.Println("Loading genome", g)
		m.gdb[g] = []map[string]string{}
		for _, v0 := range v {
			err := m.loadIndex(v0, router)
			if err != nil {
				fail += 1
			} else {
				m.gdb[g] = append(m.gdb[g], trans(v0))
			}
		}
	}
	router.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		e, _ := json.Marshal(m.entry)

		w.Write(e)
	})
	router.HandleFunc("/ls", func(w http.ResponseWriter, r *http.Request) {
		e, _ := json.Marshal(m.jdata)

		w.Write(e)
	})
	router.HandleFunc("/genomes", func(w http.ResponseWriter, r *http.Request) {
		g := []string{}
		for k, _ := range gs {
			g = append(g, k)
		}
		e, _ := json.Marshal(g)

		w.Write(e)
	})

	router.HandleFunc("/{genome}/ls", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		k := vars["genome"]
		if e, ok := m.gdb[k]; ok {
			e, _ := json.Marshal(e)
			w.Write(e)
		} else {
			w.Write([]byte("[]"))
		}
	})
	router.HandleFunc("/{genome}/list", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		k := vars["genome"]
		if v, ok := gs[k]; ok {
			u := []string{}
			for _, v0 := range v {
				u = append(u, v0.dbname)
			}
			e, _ := json.Marshal(u)
			w.Write(e)
		} else {
			w.Write([]byte("[]"))
		}
	})
	if fail > 0 {
		s := fmt.Sprintf("Fail loading %d / %d sheets", fail, len(indexes))
		log.Println(s)
		return errors.New(s)
	} else {
		log.Println(fmt.Sprintf("Successful loaded %d sheets", len(indexes)))
	}

	return nil
}

func (m *Loader) LoadWorkbook(wb *SimpleWorkbook, router *mux.Router) error {
	indexes, err := ParseSimpleWb(wb)
	if err != nil {
		return err
	}
	return m.loadIndexesTo(indexes, router)
}

func (m *Loader) loadIndex(index dataIndex, router *mux.Router) error {
	var err error
	if index.genome != "all" {
		log.Println("  Loading sheet", index.dbname) //TODO double name
		r, err := m.loadData(index.genome+"/"+index.dbname, index.data, index.format)
		//TODO not really need to load uri
		if err == nil {
			m.Data[index.genome+"/"+index.dbname] = r //TODO double name
			r.ServeTo(router)
		} else {
			log.Println(err)
		}
	} else {
		log.Println("  Loading sheet", index.dbname) //TODO double name
		r, err := m.loadData(index.dbname, index.data, index.format)
		if err == nil {
			m.Data[index.dbname] = r //TODO double name
			r.ServeTo(router)
		} else {
			log.Println(err)
		}
	}
	return err
}

func (m *Loader) loadData(dbname string, data interface{}, format string) (DataRouter, error) {
	f := m.Factory(dbname, data, format)
	if f == nil {
		return nil, errors.New("format not support")
	}
	return f(dbname, data)
}
