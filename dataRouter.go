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
)

//TODO Load Indexes From gsheet and xls

type dataIndex struct {
	dbname string
	data   interface{} // map[string]string or map[string][]string? could be uri or more sofisticated data structure such as binindex image
	format string
}

func Load(uri string, router *mux.Router) error {
	return LoadIndexURI(uri, router)
}

/* LoadIndexURI is a replace for Load
 *  it is more flexible for complex struture data
 */
func LoadIndexURI(uri string, router *mux.Router) error {
	d, err := smartParseURI(uri)
	if err != nil {
		return err
	}
	err = loadIndexesTo(d, router)
	return err
}

/* uri could be google sheets id
 * or uri
 * or .xls file
 */
func smartParseURI(uri string) ([]dataIndex, error) {
	http, _ := regexp.Compile("^http://")
	https, _ := regexp.Compile("^https://")

	if len(uri) == len("1DEvA94QkN1KZQT51IYOOcIvGL2Ux7Qwqe5IpE9Pe1N8") { //GOOGLE SHEET ID ?
		if _, err := os.Stat(uri); os.IsNotExist(err) {
			if !http.MatchString(uri) && !https.MatchString(uri) {
				//managers = AddGSheets(uri, "client_secret.json", router) //TODO handle client_secret json variable.
				return parseGSheet(uri)

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

/* For API , please using LoadIndexURI */
func loadIndexesTo(indexes []dataIndex, router *mux.Router) error {
	fail := 0
	entry := []string{}
	jdata := []map[string]string{}
	for _, v := range indexes {
		err := loadIndex(v, router)
		if err != nil {
			fail += 1
		} else {
			switch v.data.(type) {
			case string:
				jdata = append(jdata, map[string]string{
					"dbname": v.dbname,
					"uri":    v.data.(string),
					"format": v.format,
				})
			default:
				jdata = append(jdata, map[string]string{
					"dbname": v.dbname,
					"uri":    "null",
					"format": v.format,
				})

			}
			entry = append(entry, v.dbname)
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
func loadIndex(index dataIndex, router *mux.Router) error {
	r, err := loadData(index.dbname, index.data, index.format) //TODO not really need to load uri
	if err == nil {
		log.Println("Loading to server", index.dbname)
		r.ServeTo(router)
	} else {
		log.Println(err)
	}
	return err
}

func loadData(dbname string, data interface{}, format string) (DataRouter, error) {
	if f, ok := loaders[format]; ok {
		return f(dbname, data)
	}
	return nil, errors.New("format not support")
}
