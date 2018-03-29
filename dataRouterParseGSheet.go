package data

import (
	"context"
	"log"
	"os"
	"path"
	"regexp"

	"golang.org/x/oauth2/google"
	sheets "google.golang.org/api/sheets/v4"
)

//TODO: dataIndex data mv to map[string]interface{}
//      based on header
func objectFactory(keys []string, values []string) map[string]interface{} {
	if len(keys) != len(values) {
		return nil
	}
	retv := make(map[string]interface{})
	for i := 0; i < len(keys); i++ {
		retv[keys[i]] = values[i]
	}
	return retv
}

func parseGSheet(spreadsheetId string, dir string) ([]dataIndex, error) {
	di := []dataIndex{}
	httpP, _ := regexp.Compile("^http://")
	httpsP, _ := regexp.Compile("^https://")
	myFormatP, _ := regexp.Compile("^_format_:\\S+:\\S+$")

	ctx := context.Background()

	b, err := Asset("client_secret.json")
	//TODO any json which allow access gsheet.
	//b, err := ioutil.ReadFile(clientSecretJson)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	//TODO Get Client Agent.
	gA := gAgent{dir}
	client := gA.getClient(ctx, config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets Client %v", err)
	}
	_, c := readSheet("Config", srv, spreadsheetId, 1, []int{2}) //header TODO
	index := readIndex(srv, spreadsheetId)
	root, _ := c["root"]

	for _, e := range index {
		g := e.Genome //TODO
		k := e.Id
		v := e.Type
		if k[0] == '#' {
			continue
		}
		//TODO add Interface For Complicated Data Input multi Columns.
		var s map[string]interface{}
		var h []string
		h, s = readSheet(k, srv, spreadsheetId, e.Nc, e.Vc) //header TODO
		log.Println("TODO Handle Header", h)
		format := v
		data := make(map[string]interface{})
		if len(e.Vc) == 1 {
			if format == "map" {
				for k0, v0 := range s {
					data[k0] = v0.(string)
				}
			} else { //for file
				for id, loc := range s {
					var uri string
					// Add MORE TOLERATE GUESS URI INDEX FROM HEADER add support _format_:[format]:[uri]
					if httpP.MatchString(loc.(string)) || httpsP.MatchString(loc.(string)) || myFormatP.MatchString(loc.(string)) {
						uri = loc.(string)
						data[id] = uri
					} else {
						uri = path.Join(root.(string), loc.(string)) //TODO
						if _, err := os.Stat(uri); err == nil {
							data[id] = uri
						} else {
							log.Println("WARNING!!! cannot reading", uri, id)
						}
					}

				}
			}
		} else { //Vc columns > 1
			for id, vals := range s {
				var uri string
				//TODO
				// Add MORE TOLERATE GUESS URI INDEX FROM HEADER add support _format_:[format]:[uri]
				loc := vals.([]string)[0]
				if httpP.MatchString(loc) || httpsP.MatchString(loc) || myFormatP.MatchString(loc) {
					data[id] = vals
				} else {
					uri = path.Join(root.(string), loc) //TODO
					if _, err := os.Stat(uri); err == nil {
						//TODO Add MORE TOLERATE GUESS URI INDEX FROM HEADER
						vals.([]string)[0] = uri
						//data[id] = vals //TODO Add Vals Mapping to value Map handle h here
						//map[string]
						//convert vals to map[string]interface{}
						data[id] = objectFactory(h, vals.([]string))
					} else {
						log.Println("WARNING!!! cannot reading", uri, id)
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
	return di, nil
}
