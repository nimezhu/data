package data

import (
	"errors"
	"io"
	"io/ioutil"
	"log"
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

/*
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
		return NewBigBedManager2(uri, dbname)
	case "track":
		return NewTrackManager(uri, dbname)
		//case "image":
		//	return NewTabixImageManager(uri, dbname)
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
