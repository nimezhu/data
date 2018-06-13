package data

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"

	"github.com/gorilla/mux"
	minio "github.com/minio/minio-go"
	"github.com/nimezhu/netio"
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

func newDataManager(dbname string, uri string, format string) DataManager {
	switch format {
	case "file":
		return NewFileManager(uri, dbname)
	case "bigwig":
		return NewBigWigManager(uri, dbname)
	case "hic":
		return NewHicManager(uri, dbname)
	case "map":
		return NewMapManager(uri, dbname)
	case "tabix":
		return NewTabixManager(uri, dbname)
	case "bigbed":
		return NewBigBedManager(uri, dbname)
	case "track":
		return NewTrackManager(uri, dbname)
	case "image":
		return NewTabixImageManager(uri, dbname)
	}
	return nil
}

func initDataManager(dbname string, format string) DataManager {
	switch format {
	case "file":
		return InitFileManager(dbname)
	case "bigwig":
		return InitBigWigManager(dbname)
	case "hic":
		return InitHicManager(dbname)
	case "map":
		return InitMapManager(dbname)
	case "track":
		return InitTrackManager(dbname)
	}
	return nil
}

/* Obsoleted */
func ReadJsonToManagers(uri string, router *mux.Router) map[string]DataManager {
	m := map[string]DataManager{}
	r, err := netio.NewReadSeeker(uri)
	checkErr(err)
	var dat map[string]interface{}
	byt, err := ioutil.ReadAll(r)
	checkErr(err)
	if err = json.Unmarshal(byt, &dat); err != nil {
		panic(err)
	}
	data := dat["data"].([]interface{})
	meta := dat["meta"].([]interface{})
	jdata := []map[string]string{}
	entry := []string{}
	for i, v := range meta {
		v1 := v.(map[string]interface{})
		format, _ := v1["format"].(string)
		dbname, _ := v1["dbname"].(string)
		dm := initDataManager(dbname, format)
		m[dbname] = dm
		jdata = append(jdata, map[string]string{
			"dbname": dbname,
			"uri":    "null",
			"format": format,
		})
		entry = append(entry, dbname)
		for k, v := range data[i].(map[string]interface{}) {
			fmt.Println("add", v.(string), "as", k, "in db", dbname)
			dm.AddURI(v.(string), k)
		}
		dm.ServeTo(router)
	}
	router.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		e, _ := json.Marshal(entry)
		
		w.Write(e)
	})
	router.HandleFunc("/ls", func(w http.ResponseWriter, r *http.Request) {
		e, _ := json.Marshal(jdata)
		
		w.Write(e)
	})
	//TODO Load Data Manager (loadDataManager)
	return m
}

/*Obsoleted */
func AddDataManagers(uri string, router *mux.Router) map[string]DataManager {
	m := map[string]DataManager{}
	reader, err := netio.NewReadSeeker(uri)
	checkErr(err)
	entry := []string{}
	r := csv.NewReader(reader)
	r.Comma = '\t'
	r.Comment = '#'
	lines, err := r.ReadAll()
	checkErr(err)
	jdata := []map[string]string{}
	for i, line := range lines {
		if i == 0 {
			continue
		}
		dbname, uri, format := line[0], line[1], line[2]
		entry = append(entry, dbname)
		a := newDataManager(dbname, uri, format)
		jdata = append(jdata, map[string]string{
			"dbname": dbname,
			"uri":    uri,
			"format": format,
		})
		a.ServeTo(router)
		m[dbname] = a
	}
	router.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		e, _ := json.Marshal(entry)
		
		w.Write(e)
	})
	router.HandleFunc("/ls", func(w http.ResponseWriter, r *http.Request) {
		e, _ := json.Marshal(jdata)
		
		w.Write(e)
	})
	return m
}

/* obsoleted */
/*
func AddAsticodeToWindow(w *astilectron.Window, dbmap map[string]DataManager) {
	w.On(astilectron.EventNameWindowEventMessage, func(e astilectron.Event) (deleteListener bool) {
		var m string
		var dat map[string]interface{}
		e.Message.Unmarshal(&m)
		if err := json.Unmarshal([]byte(m), &dat); err != nil {
			panic(err)
		}
		if dat["code"] == "add" {
			dbname, ok1 := dat["dbname"].(string) // prefix(dbname) start with \/
			id, ok2 := dat["id"].(string)
			uri, ok3 := dat["uri"].(string)
			//log.Println("add from web", prefix, id, uri)
			//log.Println(ok1, ok2, ok3)
			if ok1 && ok2 && ok3 {
				if dbi, ok := dbmap[dbname]; ok {
					//log.Println("adding", ok)
					dbi.AddURI(uri, id)
				}
			}
		}
		if dat["code"] == "del" {
			dbname, ok1 := dat["dbname"].(string) //prefix(dbname) start with \/
			id, ok2 := dat["id"].(string)
			//uri, ok3 := dat["uri"].(string)
			if ok1 && ok2 {
				if dbi, ok := dbmap[dbname]; ok {
					dbi.Del(id)
				}
			}
		}
		if dat["code"] == "move" {
			log.Println(dat)
			dbname, ok1 := dat["dbname"].(string) //prefix(dbname) start with \/
			id1, ok2 := dat["from"].(string)
			id2, ok3 := dat["to"].(string)
			if ok1 && ok2 && ok3 {
				if dbi, ok := dbmap[dbname]; ok {
					dbi.Move(id1, id2)
				}
			}
		}
		return false
	})
}
*/

/* LoadCloud
 */
func LoadCloud(uri string, id string, router *mux.Router) (Manager, error) { //trackmanager in minio
	f, err := netio.Open(uri)
	if err != nil {
		return nil, err
	}
	c, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(c), "\n")
	t := strings.Split(lines[0], "\t")
	if t[0] == "minio" {

		return loadMinio(t[1:], lines[1:], id, router)
	}
	log.Println(t[0], t[0] == "minio")
	return nil, errors.New("not support")
}

/* format
 * minio	server	accessKey	secretKey
 * id	bucketName	fileName
 */
func loadMinio(c, lines []string, id string, router *mux.Router) (Manager, error) {
	https, _ := regexp.Compile("^https://")
	http, _ := regexp.Compile("^http://")
	markRe, _ := regexp.Compile("^ *#")
	elineRe, _ := regexp.Compile("^ *$")
	uri := c[0]
	accessKey := c[1]
	secretKey := c[2]
	isHttps := false
	if https.MatchString(uri) {
		isHttps = true
		uri = strings.Replace(uri, "https://", "", 1)
	}
	if http.MatchString(uri) {
		isHttps = false
		uri = strings.Replace(uri, "http://", "", 1)
	}
	server, err := minio.New(uri, accessKey, secretKey, isHttps)
	if err != nil {
		return nil, err
	}
	/* Minio Format alias \t bucket \t uri*/
	trackManager := InitTrackManager(id)
	for _, l := range lines {
		if elineRe.MatchString(l) {
			continue
		}
		if markRe.MatchString(l) {
			continue
		}
		t := strings.Split(l, "\t")
		reader, err := server.GetObject(t[1], t[2], minio.GetObjectOptions{})
		if err == nil {
			log.Println("adding")
			trackManager.Add(t[0], reader, "minio:"+uri+":/"+t[1]+"/"+t[2])
		} else {
			log.Println(err)
		}
	}

	return trackManager, nil
}
