package data

import (
	"log"
	"os"
	path "path/filepath"
	"regexp"
	"strings"

	"github.com/nimezhu/indexed/bbi"
	"github.com/nimezhu/netio"
)

var (
	h1, _ = regexp.Compile("^http://")
	h2, _ = regexp.Compile("^https://")
)

/* TODO: Manage Indexes Functions
   0 : local file
   1 : remote file without local index
   2 : remote file with local index
*/
func checkUri(uri string, root string) (string, int) {
	if h1.MatchString(uri) || h2.MatchString(uri) {
		var dir string
		if h1.MatchString(uri) {
			dir = strings.Replace(uri, "http://", "", 1)
		} else {
			dir = strings.Replace(uri, "https://", "", 1)
		}
		dir += ".index"
		p := strings.Split(dir, "/")
		ps := append([]string{root}, p...)
		fn := path.Join(ps...)
		if _, err := os.Stat(fn); os.IsNotExist(err) {
			return fn, 1
		} else {
			return fn, 2
		}
	} else {
		return "", 0
	}
}

/* Force Update */
func updateIdx(uri string, root string) (int, error) {
	fn, mode := checkUri(uri, root)
	log.Println("fetching")
	reader, err := netio.NewReadSeeker(uri)
	if err != nil {
		return -1, err
	}
	bwf := bbi.NewBbiReader(reader)
	defer bwf.Close()
	err = bwf.InitIndex()
	if err != nil {
		log.Println(err)
		return -1, err
	}
	/* TODO Handle Remove os.Exists */
	os.MkdirAll(path.Dir(fn), 0700)
	f, err := os.Create(fn)
	if err != nil {
		log.Println("error in creating", err)
	}

	err = bwf.WriteIndex(f)
	if err != nil {
		return -1, err
	}
	log.Println("saved")
	f.Close()
	return mode, nil

}
func saveIdx(uri string, root string) (int, error) {
	_, mode := checkUri(uri, root)
	if mode == 2 {
		log.Println("index in local")
	}
	if mode == 1 {
		return updateIdx(uri, root)
	}
	return mode, nil
}

func readBw(uri string) *bbi.BigWigReader {
	reader, err := netio.NewReadSeeker(uri)
	checkErr(err)
	bwf := bbi.NewBbiReader(reader)
	bwf.InitIndex()
	//log.Println("in reading idx of", uri)
	bw := bbi.NewBigWigReader(bwf)
	return bw
}
