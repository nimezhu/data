package data

import (
	"compress/gzip"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/nimezhu/netio"
)

type Band struct {
	Start int
	End   int
	Id    string
	Value string
}
type CytoBand struct {
	data map[string][]Band
}
type CytoBandManager struct {
	data map[string]*CytoBand
}

/*
type DataManager interface {
	AddURI(uri string, key string) error
	Del(string) error
	ServeTo(*mux.Router)
	List() []string
	Get(string) (string, bool)
	Move(key1 string, key2 string) bool
}
 Add(key string) -> AddURI
*/

func parseCytoBand(txt string) (CytoBand, error) {
	lines := strings.Split(txt, "\n")
	cytoband := CytoBand{}
	m := make(map[string][]Band)
	for _, line := range lines {
		t := strings.Split(line, "\t")
		if len(t) >= 5 {
			_, ok := m[t[0]]
			if !ok {
				m[t[0]] = make([]Band, 0, 100)
			}
			start, _ := strconv.Atoi(t[1])
			end, _ := strconv.Atoi(t[2])
			m[t[0]] = append(m[t[0]], Band{start, end, t[3], t[4]})
		}
	}
	return cytoband, nil
}
func (m *CytoBandManager) AddURI(uri string, key string) error {
	f, err := netio.NewReadSeeker(uri)
	if err != nil {
		return err
	}
	z, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	txt, err := ioutil.ReadAll(z)
	if err != nil {
		return err
	}
	//log.Println("TODO....", string(txt))
	cytoband, err := parseCytoBand(string(txt))
	if err != nil {
		return err
	}
	m.data[key] = &cytoband
	return nil
}
func generateUrl(genome string) string {
	uri := "http://hgdownload.cse.ucsc.edu/goldenPath/" + genome + "/database/cytoBand.txt.gz"
	return uri
}
func (m *CytoBandManager) Add(genome string) error {
	return m.AddURI(generateUrl(genome), genome)
}

func (m *CytoBandManager) Del(genome string) error {
	delete(m.data, genome)
	return nil
}
func (m *CytoBandManager) List() []string {
	keys := []string{}
	for k, _ := range m.data {
		keys = append(keys, k)
	}
	return keys
}

func (m *CytoBandManager) Get(genome string) (string, bool) {
	if _, ok := m.data[genome]; ok {
		return generateUrl(genome), true
	}
	return "", false
}
func (m *CytoBandManager) Move(key1 string, key2 string) bool {
	v, ok := m.data[key1]
	if ok {
		m.data[key2] = v
		delete(m.data, key1)
	}
	return ok
}
