package data

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/nimezhu/data/image"
)

/* not working in windows */
func getRootDir(dirs []string) string {
	if len(dirs) == 1 {
		return path.Dir(dirs[0])
	}
	r := ""
	a := make([][]string, len(dirs))
	for i, v := range dirs {
		a[i] = strings.Split(v, string(filepath.Separator))
	}
	for j, sign := 0, true; sign; j++ {
		//v := a[0][j]
		for i := 1; i < len(a); i++ {
			if len(a[i]) < j {
				sign = false
				break
			}
			if a[i][j] != a[0][j] {
				sign = false
				break
			}
		}
		if sign && len(a[0][j]) > 0 {
			r += "/" + a[0][j]
		}
	}
	return r
}

/*TODO implement DataRouter
 *
 */
type Bed3 struct {
	chr   string
	start int
	end   int
}

func (b *Bed3) Chr() string {
	return b.chr
}
func (b *Bed3) Start() int {
	return b.start
}
func (b *Bed3) End() int {
	return b.end
}

type BinindexImageRouter struct {
	dataMap map[string]*bedImage //file name
	index   *BinIndexMap
	dbname  string //sheet tab name
	root    string //need to calc
	inited  bool
}
type bedImage struct {
	Name     string
	Uri      string
	Position []Bed3
}

func InitBinindexImageRouter(dbname string) *BinindexImageRouter {
	return &BinindexImageRouter{
		make(map[string]*bedImage),
		NewBinIndexMap(),
		dbname,
		"",
		false,
	}
}

func (db *BinindexImageRouter) ServeTo(router *mux.Router) {
	//Warning : Add or Load Before ServeTo.
	db.inited = true
	uris := make([]string, len(db.dataMap))
	i := 0
	for _, v := range db.dataMap {
		uris[i] = v.Uri
		i++
	}
	db.root = getRootDir(uris)
	idToUri := make(map[string]string)
	for k, v := range db.dataMap {
		u := strings.Replace(v.Uri, db.root, "", 1)
		idToUri[k] = u
	}
	image.AddTo(router, db.dbname+"/images", db.root) //start server for host image files
	router.HandleFunc("/"+db.dbname+"/get/{chr}:{start}-{end}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		params := mux.Vars(r)
		//id := params["id"]
		chrom := params["chr"]
		s, _ := strconv.Atoi(params["start"])
		e, _ := strconv.Atoi(params["end"])
		vals, _ := db.index.QueryRegion(chrom, s, e)
		for v := range vals {
			io.WriteString(w, fmt.Sprint(v))
			io.WriteString(w, "\n")
		}
	})
	router.HandleFunc("/"+db.dbname+"/img/{id}", func(w http.ResponseWriter, r *http.Request) {
		a := r.URL.String()
		k := strings.Split(a, "?")
		s := ""
		if len(k) > 1 {
			s = k[1]
		}
		params := mux.Vars(r)
		id := params["id"]
		if uri, ok := idToUri[id]; ok {
			url := "/" + db.dbname + "/images/" + uri
			if s != "" {
				url += "?" + s
			}
			http.Redirect(w, r, url, http.StatusTemporaryRedirect)
		}
	})

}
func (db *BinindexImageRouter) Add(image *bedImage) error {
	if db.inited {
		//TODO Fix Judge Root
		return errors.New("server has been inited, couldn't add more images from other directory")
	}
	for _, bed := range image.Position {
		bed4 := Bed4{
			bed.Chr(), bed.Start(), bed.End(), image.Name,
		}
		db.index.Insert(bed4)
	}
	db.dataMap[image.Name] = image
	return nil
}
func (db *BinindexImageRouter) Load(images []bedImage) error {
	for _, v := range images {
		err := db.Add(&v)
		if err != nil {
			return err
		}
	}
	return nil
}
