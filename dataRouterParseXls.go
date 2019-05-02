package data

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/nimezhu/netio"
	"github.com/tealeg/xlsx"
)

func logErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

func parseXls(uri string) ([]dataIndex, error) {
	di := []dataIndex{}
	httpP, _ := regexp.Compile("^http://")
	httpsP, _ := regexp.Compile("^https://")
	r, err := netio.NewReadSeeker(uri)
	logErr(err)
	bs, err := ioutil.ReadAll(r)
	logErr(err)
	xlFile, err := xlsx.OpenBinary(bs)
	logErr(err)
	config, ok := xlFile.Sheet["Config"] //get root
	if !ok {

	}
	cfgRows := config.Rows
	var root string
	for _, row := range cfgRows[1:] {
		x := row.Cells
		if x[0].String() == "root" {
			root = x[1].String()
		}
	}
	index, ok := xlFile.Sheet["Index"]
	if !ok {
		log.Println("can not find Index sheet")
	}
	idxRows := index.Rows
	type item struct {
		Genome   string `xlsx:"0"`
		Id       string `xlsx:"1"`
		Type     string `xlsx:"2"`
		Ns       string `xlsx:"3"` //TODO string parse to int
		Vs       string `xlsx:"4"` //TODO VC is []ints
		Preserve bool   `xlsx:"5"`
	}
	type entry struct {
		Genome   string
		Id       string
		Type     string
		Nc       int
		Vc       []int
		Preserve bool
	}
	for _, row := range idxRows[1:] {
		a := &item{}
		row.ReadStruct(a)
		fmt.Println(a)
		if a == nil {
			continue
		}
		if len(a.Id) == 0 {
			break
		}
		if a.Id[0] == '#' {
			continue
		}
		if vsheet, ok := xlFile.Sheet[a.Id]; ok {
			log.Println(a.Id)
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
				//TODO vsheet.Rows[0] as Header
				for _, r := range vsheet.Rows[1:] {
					//v0 := r.Cells[a.Vc-1].String() //TODO vc is ints
					v0 := r.Cells[vc-1].String()
					k0 := r.Cells[nc-1].String()
					//m0.AddURI(v0, k0)
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
					for _, r := range vsheet.Rows[1:] {
						id := r.Cells[nc-1].String()
						loc := r.Cells[vc-1].String() //TODO Vc To ints
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
						header[i] = vsheet.Rows[0].Cells[vci[i]-1].String()
					}
					for _, r := range vsheet.Rows[1:] {
						id := r.Cells[nc-1].String()
						loc := r.Cells[vc-1].String()
						vals := make([]string, len(vcs))
						for i, k := range vci {
							vals[i] = r.Cells[k-1].String()
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
