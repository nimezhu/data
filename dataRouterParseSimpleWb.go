package data

import (
	"log"
	"os"
	path "path/filepath"
	"regexp"
	"strconv"
	"strings"
)

/* JSON -> Simple Workbook parsed already */
func ParseSimpleWb(xlFile *SimpleWorkbook) ([]dataIndex, error) {
	di := []dataIndex{}
	cfgRows, ok := xlFile.Sheets["Config"] //get root
	if !ok {
	}
	var root string
	for _, x := range cfgRows[1:] {
		if x[0] == "root" {
			root = x[1]
		}
	}
	idxRows, ok := xlFile.Sheets["Index"]
	if !ok {
		log.Println(root)
		log.Println("can not find Index sheet")
	}
	httpP, _ := regexp.Compile("^http://")
	httpsP, _ := regexp.Compile("^https://")

	//..TODO preserver
	type item struct {
		Genome string `xlsx:"0"`
		Id     string `xlsx:"1"`
		Type   string `xlsx:"2"`
		Ns     string `xlsx:"3"` //TODO string parse to int
		Vs     string `xlsx:"4"` //TODO VC is []ints
	}
	type entry struct {
		Genome string
		Id     string
		Type   string
		Nc     int
		Vc     []int
	}
	for _, row := range idxRows[1:] {
		a := &item{
			row[0],
			row[1],
			row[2],
			row[3],
			row[4],
		}
		if a == nil {
			continue
		}
		if len(a.Id) == 0 {
			break
		}
		if a.Id[0] == '#' {
			continue
		}
		if vsheet, ok := xlFile.Sheets[a.Id]; ok {
			log.Println("TO Process", a)
			format := a.Type
			k := a.Id
			g := a.Genome
			data := make(map[string]interface{})
			if a.Type == "map" {
				vcs := strings.Split(a.Vs, ",")
				vc, err := strconv.Atoi(vcs[0])
				if err != nil {
					vc = colNameToNumber(vcs[0])
				}
				nc, err := strconv.Atoi(a.Ns)
				if err != nil {
					nc = colNameToNumber(a.Ns)
				}
				for _, r := range vsheet[1:] {
					v0 := r[vc-1]
					k0 := r[nc-1]
					data[k0] = v0

				}
			} else {
				vcs := strings.Split(a.Vs, ",")
				vc, err := strconv.Atoi(vcs[0])
				if err != nil {
					vc = colNameToNumber(vcs[0])
				}
				nc, err := strconv.Atoi(a.Ns)
				if err != nil {
					nc = colNameToNumber(a.Ns)
				}
				log.Println(nc, vc)
				if len(vcs) == 1 {
					for _, r := range vsheet[1:] {
						id := r[nc-1]
						loc := r[vc-1] //TODO Vc To ints
						var uri string
						if httpP.MatchString(loc) || httpsP.MatchString(loc) {
							uri = loc
							data[id] = uri
						} else {
							uri = path.Join(root, loc) //TODO
							if _, err := os.Stat(uri); err == nil {
								data[id] = uri
							} else {
								log.Println("WARNING!!! cannot reading", uri, id)
							}
						}
					}
				} else {
					vci := make([]int, len(vcs))
					for i, k := range vcs {
						v, err := strconv.Atoi(k)
						if err != nil {
							v = colNameToNumber(k)
						}
						vci[i] = v
					}
					header := make([]string, len(vcs))

					for i := 0; i < len(vcs); i++ {
						header[i] = vsheet[0][vci[i]-1]
					}
					for _, r := range vsheet[1:] {
						id := r[nc-1]
						loc := r[vc-1]
						vals := make([]string, len(vcs))
						for i, k := range vci {
							vals[i] = r[k-1]
						}
						var uri string
						if httpP.MatchString(loc) || httpsP.MatchString(loc) {
							data[id] = objectFactory(header, vals)
						} else {
							uri = path.Join(root, loc) //TODO
							vals[0] = uri
							if _, err := os.Stat(uri); err == nil {
								data[id] = objectFactory(header, vals)
							} else {
								log.Println("WARNING!!! cannot reading", uri, id)
							}
						}
					}
				}
			}
			d := dataIndex{
				g,
				k,
				data,
				format,
			}
			di = append(di, d)
		}
	}

	return di, nil
}
