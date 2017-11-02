package data

import (
	"errors"
	"path"
	"strings"

	"github.com/gorilla/mux"
	"github.com/nimezhu/data/image"
)

// No Data Structure implemented , leave interface for future for large data set.
/*
type SmallSetImageManager struct {
	uriMap         map[string]string
	imageToRegions map[string][]ShortBed
	dbname         string
}
*/

/*Implement DataManager
type DataManager interface {
 AddURI(uri string, key string) error
 Del(string) error
 ServeTo(*mux.Router)
 List() []string
 Get(string) (string, bool)
 Move(key1 string, key2 string) bool
}

Customized
AddPos(key string, regions []Region)

*/
func getRootDir(dirs []string) string {
	r := ""
	a := make([][]string, len(dirs))
	for i, v := range dirs {
		a[i] = strings.Split(v, "/") //TODO handle windows ???
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

/* AddImagesTo:
 *   images should install as local files.
 *   TODO: uri is web link?
 *         how to add names ? redirect?
 *         rename tableName to tableName.image
 */
func AddImagesTo(uris []string, tableName string, dataRoot string, router *mux.Router) {
	root := path.Join(dataRoot, getRootDir(uris))
	image.AddTo(router, tableName, root)
}

// BinindexImageManager is a TabixImageManager but in memory
// partially implement data manager interface function.
// ServeTo(*mux.Router) //add handlers to router
type BinindexImageManager struct {
	uriMap         map[string]string //file name
	dataMap        *BinIndexMap
	imageToRegions map[string]string //associate name to regions
	dbname         string            //sheet tab name
	root           string            //need to calc
}

func (T *BinindexImageManager) addBeds(uri string) error {
	/*
		f, err := netio.ReadAll(uri)
		if err != nil {
			return err
		}
		m := make(map[string][]Bed4)
		lines := strings.Split(string(f), "\n")
		for _, line := range lines {
			if len(line) == 0 {
				continue
			}
			w := strings.Split(string(line), "\t")
			start, _ := strconv.Atoi(w[1])
			end, _ := strconv.Atoi(w[2])
			if _, ok := m[w[3]]; ok {
				m[w[3]] = append(m[w[3]], Bed4{w[0], start, end, w[3]})
			} else {
				m[w[3]] = []Bed4{Bed4{w[0], start, end, w[3]}}
			}
		}

		for k, v := range m {
			T.imageToRegions[k] = regionsText(v)
		}
	*/
	return nil
}
func (B *BinindexImageManager) AddURI(k string, v string) error {
	return errors.New("not implemented")
}
func (B *BinindexImageManager) Del(k string) error {
	return errors.New("not implemented")
}
func (B *BinindexImageManager) ServeTo(router *mux.Router) {
}
