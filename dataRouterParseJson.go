package data

import (
	"encoding/json"
	"io/ioutil"

	"github.com/nimezhu/netio"
)

//Replace AddDataManagers
func parseJson(uri string) ([]dataIndex, error) {
	di := []dataIndex{}
	r, err := netio.NewReadSeeker(uri)
	checkErr(err)
	var dat map[string]interface{}
	byt, err := ioutil.ReadAll(r)
	checkErr(err)
	if err = json.Unmarshal(byt, &dat); err != nil {
		panic(err)
	}
	data := dat["data"].([]interface{})
	meta := dat["meta"].([]interface{})
	for i, v := range meta {
		v1 := v.(map[string]interface{})
		format, _ := v1["format"].(string)
		dbname, _ := v1["dbname"].(string)
		data2 := map[string]string{}
		for k, v := range data[i].(map[string]interface{}) {
			data2[k] = v.(string)
		}
		di = append(di, dataIndex{dbname, data2, format})
	}

	return di, nil
}
