package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

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
		r.ServeTo(router)
	}
	return err
}

/*TODO SHORT */
func loadData(dbname string, data interface{}, format string) (DataRouter, error) {
	switch format {
	case "file":
		switch v := data.(type) {
		default:
			fmt.Printf("unexpected type %T", v)
		case string:
			return NewFileManager(data.(string), dbname), nil
		}
	case "bigwig":
		switch v := data.(type) {
		default:
			fmt.Printf("unexpected type %T", v)
			return nil, errors.New(fmt.Sprintf("bigwig format not support type %T", v))
		case string:
			return NewBigWigManager(data.(string), dbname), nil
		case map[string]interface{}:
			a := InitBigWigManager(dbname)
			for key, val := range data.(map[string]interface{}) {
				a.AddURI(val.(string), key)
			}
			return a, nil
		}
	case "hic":
		switch v := data.(type) {
		default:
			fmt.Printf("unexpected type %T", v)
			return nil, errors.New(fmt.Sprintf("hic format not support type %T", v))
		case string:
			return NewHicManager(data.(string), dbname), nil
		case map[string]interface{}:
			a := InitHicManager(dbname)
			for key, val := range data.(map[string]interface{}) {
				a.AddURI(val.(string), key)
			}
			return a, nil
		}
	case "map":
		switch v := data.(type) {
		default:
			fmt.Printf("unexpected type %T", v)
		case string:
			return NewMapManager(data.(string), dbname), nil
		case map[string]interface{}:
			a := InitMapManager(dbname)
			for key, val := range data.(map[string]interface{}) {
				a.AddURI(val.(string), key)
			}
			return a, nil
		}
	case "tabix":
		switch v := data.(type) {
		default:
			fmt.Printf("unexpected type %T", v)
		case string:
			return NewTabixManager(data.(string), dbname), nil
		}
	case "bigbed":
		switch v := data.(type) {
		default:
			fmt.Printf("unexpected type %T", v)
			return nil, errors.New(fmt.Sprintf("bigbed format not support type %T", v))
		case string:
			return NewBigBedManager(data.(string), dbname), nil
		case map[string]interface{}:
			a := InitBigBedManager(dbname)
			for key, val := range data.(map[string]interface{}) {
				a.AddURI(val.(string), key)
			}
			return a, nil
		}
	case "track":
		switch v := data.(type) {
		default:
			fmt.Printf("unexpected type %T", v)
			return nil, errors.New(fmt.Sprintf("track format not support type %T", v))
		case string:
			return NewTrackManager(data.(string), dbname), nil
		case map[string]interface{}:
			a := InitTrackManager(dbname)
			for key, val := range data.(map[string]interface{}) {
				a.AddURI(val.(string), key)
			}
			return a, nil
		}
	case "image":
		switch v := data.(type) {
		default:
			fmt.Printf("unexpected type %T", v)
		case string:
			return NewTabixImageManager(data.(string), dbname), nil //TODO MODIFICATION
		}
	case "img":
		switch v := data.(type) {
		default:
			fmt.Printf("unexpected type %T \n", v)
		case map[string]interface{}:
			r := InitBinindexImageRouter(dbname)
			r.Load(_parseToBedImage(data.(map[string]interface{})))
			return r, nil
			//return NewTabixImageManager(data.(string), dbname), nil //
		}
	}
	return nil, errors.New("format not support")
}

func _parseToBedImage(d map[string]interface{}) []bedImage {
	r := make([]bedImage, len(d))
	i := 0
	for k, v := range d {
		r[i] = bedImage{
			k,
			v.([]string)[0],
			parseRegions(v.([]string)[1]),
		}
		i++
	}
	return r
}
func parseRegions(txt string) []Bed3 {
	l := strings.Split(txt, ";")
	b := make([]Bed3, len(l))
	for i, v := range l {
		b[i] = parseRegion(v)
	}
	return b
}
func parseRegion(txt string) Bed3 {
	x := strings.Split(txt, ":")
	chr := x[0]
	se := strings.Split(x[1], "-")
	start, _ := strconv.Atoi(se[0])
	end, _ := strconv.Atoi(se[1])
	return Bed3{chr, start, end}
}
