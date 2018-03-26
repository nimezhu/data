package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/nimezhu/indexed"
)

//TODO Load Indexes From gsheet and xls
//TODO dataIndexes with attrs ...
//TODO DATAURI SYSTEM
//TODO SERVER / GENOME(VERSION) / NAME
//TODO LS Parameters Get Specific Attrs Such as longLabels
//TODO Parse Descriptions
type dataIndex struct {
	genome string      `json:"genome"`
	dbname string      `json:"dbname"`
	data   interface{} // map[string]string or map[string][]string? could be uri or more sofisticated data structure such as binindex image
	format string      `json:"format"`
}

/*middleware*/
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

/* LoadIndexURI is a replace for Load
 *  it is more flexible for complex struture data
 */
func (m *Loader) loadIndexURI(uri string, router *mux.Router) error {
	d, err := m.smartParseURI(uri)
	if err != nil {
		return err
	}
	err = m.loadIndexesTo(d, router)
	return err
}

/* LoadIndexURI is a replace for Load
 *  it is more flexible for complex struture data
 */
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

/* uri could be google sheets id
 * or uri
 * or .xls file
 */
func SaveIndex(uri string, root string) error {
	h1, _ := regexp.Compile("^http://")
	h2, _ := regexp.Compile("^https://")
	m := NewLoader(root)
	d, err := m.smartParseURI(uri)
	if err != nil {
		return err
	}
	//bw := InitBigWigManager2("noname", root)
	for _, v := range d {
		if v.format == "bigwig" || v.format == "bigbed" || v.format == "track" {
			switch t := v.data.(type) {
			default:
				fmt.Printf("unexpected type %T", t)
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
						log.Println("process " + val.(string))
						saveIdx(val.(string), root)
					}
				}
			}
		}
	}
	return nil
}
func (m *Loader) smartParseURI(uri string) ([]dataIndex, error) {
	http, _ := regexp.Compile("^http://")
	https, _ := regexp.Compile("^https://")

	if len(uri) == len("1DEvA94QkN1KZQT51IYOOcIvGL2Ux7Qwqe5IpE9Pe1N8") { //GOOGLE SHEET ID ?
		if _, err := os.Stat(uri); os.IsNotExist(err) {
			if !http.MatchString(uri) && !https.MatchString(uri) {
				//managers = AddGSheets(uri, "client_secret.json", router) //TODO handle client_secret json variable.
				root := path.Dir(m.IndexRoot)
				return parseGSheet(uri, root)

			}
		}
	} else {
		ext := path.Ext(uri)
		if ext == ".json" {
			return parseJson(uri) //not for obsoleted version
		} else if ext == ".xlsx" {
			return parseXls(uri)
		} else {
			return parseTxt(uri)
		}
	}
	return nil, errors.New("not recognize uri")
}

/* For API , please using LoadIndexURI
 * TODO: uri change to attrs
 *       or add attrs in string
 */
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

//TODO ADD Refresh Indexes
func (m *Loader) loadIndexesTo(indexes []dataIndex, router *mux.Router) error {
	//add genome versions .
	fail := 0
	//todo entry data and genomes in loader.
	//event driven to reload entry and jdata?
	//entry := []string{}
	//jdata := []map[string]string{}
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
	//gdb := map[string][]map[string]string{}
	for g, v := range gs {
		log.Println("Loading Genome Data ", g)
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
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(e)
	})
	router.HandleFunc("/ls", func(w http.ResponseWriter, r *http.Request) {
		e, _ := json.Marshal(m.jdata)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(e)
	})
	router.HandleFunc("/genomes", func(w http.ResponseWriter, r *http.Request) {
		g := []string{}
		for k, _ := range gs {
			g = append(g, k)
		}
		e, _ := json.Marshal(g)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(e)
	})
	for k, v := range gs {
		go func(k string, v []dataIndex) {
			router.HandleFunc("/"+k+"/list", func(w http.ResponseWriter, r *http.Request) {
				u := []string{}
				fmt.Println("list", k)
				for _, v0 := range v {
					u = append(u, v0.dbname)
				}
				e, _ := json.Marshal(u)
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Write(e)
			})
		}(k, v)
		go func(k string, e []map[string]string) {
			router.HandleFunc("/"+k+"/ls", func(w http.ResponseWriter, r *http.Request) {
				e, _ := json.Marshal(e)
				w.Header().Set("Access-Control-Allow-Origin", "*")
				w.Write(e)
			})
		}(k, m.gdb[k])
	}
	//add json and list to router
	if fail > 0 {
		return errors.New(fmt.Sprintf("Fail loading %d / %d database", fail, len(indexes)))
	}

	return nil
}

/*
func LoadIndexTo(index dataIndex, router *mux.Router) error {
	return loadIndex(index, router)
}
*/

/* serve: Add DataRouter to Router
 */
func (m *Loader) loadIndex(index dataIndex, router *mux.Router) error {
	//TODO: add genome version
	var err error
	if index.genome != "all" {
		r, err := m.loadData(index.genome+"/"+index.dbname, index.data, index.format)
		//TODO not really need to load uri
		if err == nil {
			log.Println("Loading to server", index.genome, index.dbname) //TODO double name
			m.Data[index.genome+"/"+index.dbname] = r                    //TODO double name
			r.ServeTo(router)
		} else {
			log.Println(err)
		}
	} else {
		r, err := m.loadData(index.dbname, index.data, index.format)
		if err == nil {
			log.Println("Loading to server", index.genome, index.dbname) //TODO double name
			m.Data[index.dbname] = r                                     //TODO double name
			r.ServeTo(router)
		} else {
			log.Println(err)
		}
	}
	return err
}

/* TODO load Data With
Change Data into map[string]interface{}
For all Factories
*/
func (m *Loader) loadData(dbname string, data interface{}, format string) (DataRouter, error) {
	f := m.Factory(dbname, data, format)
	if f == nil {
		return nil, errors.New("format not support")
	}
	return f(dbname, data)
}
