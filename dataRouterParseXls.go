package data

import (
	"io/ioutil"
	"log"
	"os"
	path "path/filepath"
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
	//cfgRows := config.Rows
	var root string
	config.RemoveRowAtIndex(0)
	config.ForEachRow(func(row *xlsx.Row) error {
		name := row.GetCell(0)
		if name.String() == "root" {
			root = row.GetCell(1).String()
		}
		return nil
	})
	/*
		for _, row := range cfgRows[1:] {
		}
	*/
	index, ok := xlFile.Sheet["Index"]
	if !ok {
		log.Println("can not find Index sheet")
	}
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
	//for _, row := range idxRows[1:] {
	index.RemoveRowAtIndex(0)
	index.ForEachRow(func(row *xlsx.Row) error {
		a := &item{}
		row.ReadStruct(a)
		if a == nil {
			return nil
		}
		if len(a.Id) == 0 {
			return nil
		}
		if a.Id[0] == '#' {
			return nil
		}
		if vsheet, ok := xlFile.Sheet[a.Id]; ok {
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
				vsheet.RemoveRowAtIndex(0)

				//for _, r := range vsheet.Rows[1:] {
				vsheet.ForEachRow(func(r *xlsx.Row) error {
					v0 := r.GetCell(vc - 1).String()
					k0 := r.GetCell(nc - 1).String()
					data[k0] = v0
					return nil

				})

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
				if len(vcs) == 1 {
					//for _, r := range vsheet.Rows[1:] {
					vsheet.RemoveRowAtIndex(0)
					vsheet.ForEachRow(func(r *xlsx.Row) error {
						id := r.GetCell(nc - 1).String()
						loc := r.GetCell(vc - 1).String() //TODO Vc To ints
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
						return nil
					})
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
						_r, _ := vsheet.Row(0)
						header[i] = _r.GetCell(vci[i] - 1).String()
					}
					vsheet.RemoveRowAtIndex(0)
					//for _, r := range vsheet.Rows[1:] {
					vsheet.ForEachRow(func(r *xlsx.Row) error {
						id := r.GetCell(nc - 1).String()
						loc := r.GetCell(vc - 1).String()
						vals := make([]string, len(vcs))
						for i, k := range vci {
							vals[i] = r.GetCell(k - 1).String()
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
						return nil
					})
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
		return nil
	})
	return di, nil
}
