package data

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nimezhu/data/image"
)

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

func (db *BinindexImageRouter) ServeTo(router *mux.Router) {
	//Warning : Add or Load Before ServeTo.
	db.inited = true
	uris := make([]string, len(db.dataMap))
	i := 0
	for _, v := range db.dataMap {
		uris[i] = v.Uri
		i++
	}
	image.AddTo(router, db.dbname+".image", db.root) //start server for host image files
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
