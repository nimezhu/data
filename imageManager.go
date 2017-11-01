package data

import (
	"strings"

	"github.com/gorilla/mux"
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
 */
func AddImagesTo(names []string, uris []string, tableName string, dataRoot string, router *mux.Router) {

}
