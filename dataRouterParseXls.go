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

	}
	idxRows := index.Rows
	type item struct {
		Genome   string `xlsx:"0"`
		Id       string `xlsx:"1"`
		Type     string `xlsx:"2"`
		Nc       int    `xlsx:"3"`
		Vc       string `xlsx:"4"` //TODO VC is []ints
		Preserve bool   `xlsx:"5"`
	}
	for _, row := range idxRows[1:] {
		a := &item{}
		row.ReadStruct(a)
		fmt.Println(a)
		if len(a.Id) == 0 {
			break
		}
		if a.Id[0] == '#' {
			continue
		}
		if vsheet, ok := xlFile.Sheet[a.Id]; ok {
			format := a.Type
			k := a.Id
			g := a.Genome
			data := make(map[string]interface{})
			if a.Type == "map" {
				vcs := strings.Split(a.Vc, ",")
				vc, _ := strconv.Atoi(vcs[0])
				for _, r := range vsheet.Rows[1:] {
					//v0 := r.Cells[a.Vc-1].String() //TODO vc is ints
					v0 := r.Cells[vc-1].String()
					k0 := r.Cells[a.Nc-1].String()
					//m0.AddURI(v0, k0)
					data[k0] = v0

				}
			} else {
				vcs := strings.Split(a.Vc, ",")
				vc, _ := strconv.Atoi(vcs[0])
				if len(vcs) == 1 {
					for _, r := range vsheet.Rows[1:] {
						id := r.Cells[a.Nc-1].String()
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
					for _, r := range vsheet.Rows[1:] {
						id := r.Cells[a.Nc-1].String()
						loc := r.Cells[vc-1].String() //TODO Vc To ints
						vals := make([]string, len(vcs))
						for i, k := range vcs {
							v, _ := strconv.Atoi(k)
							vals[i] = r.Cells[v-1].String()
						}
						var uri string
						if httpP.MatchString(loc) || httpsP.MatchString(loc) {
							data[id] = vals
						} else {
							uri = path.Join(root, loc) //TODO
							vals[0] = uri
							if _, err := os.Stat(uri); err == nil {
								data[id] = vals
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
