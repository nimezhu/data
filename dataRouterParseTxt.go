package data

import (
	"encoding/csv"

	"github.com/nimezhu/netio"
)

//Replace AddDataManagers
func parseTxt(uri string) ([]DataIndex, error) {
	di := []DataIndex{}
	reader, err := netio.NewReadSeeker(uri)
	checkErr(err)
	r := csv.NewReader(reader)
	r.Comma = '\t'
	r.Comment = '#'
	lines, err := r.ReadAll()
	checkErr(err)
	for i, line := range lines {
		if i == 0 {
			continue
		}
		dbname, uri, format := line[0], line[1], line[2]
		di = append(di, DataIndex{
			dbname,
			uri,
			format,
		})
	}
	return di, nil
}
