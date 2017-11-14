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

func parseGSheet(spreadsheetId string) ([]dataIndex, error) {
	di := []dataIndex{}
	httpP, _ := regexp.Compile("^http://")
	httpsP, _ := regexp.Compile("^https://")
	binindexP, _ := regexp.Compile("^binindex:")

	ctx := context.Background()

	b, err := Asset("client_secret.json")
	//TODO any json which allow access gsheet.
	//b, err := ioutil.ReadFile(clientSecretJson)
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/sheets.googleapis.com-go-quickstart.json
	config, err := google.ConfigFromJSON(b, "https://www.googleapis.com/auth/spreadsheets.readonly")
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets Client %v", err)
	}
	c := readSheet("Config", srv, spreadsheetId, 1, []int{2})
	index := readIndex(srv, spreadsheetId)
	root, _ := c["root"]

	for _, e := range index {
		k := e.Id
		v := e.Type
		if k[0] == '#' {
			continue
		}
		//TODO add Interface For Complicated Data Input multi Columns.
		var s map[string]interface{}
		s = readSheet(k, srv, spreadsheetId, e.Nc, e.Vc) //TODO GET FROM INDEX
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
					if httpP.MatchString(loc.(string)) || httpsP.MatchString(loc.(string)) || binindexP.MatchString(loc.(string)) {
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
				loc := vals.([]string)[0]
				if httpP.MatchString(loc) || httpsP.MatchString(loc) {
					data[id] = vals
				} else {
					uri = path.Join(root.(string), loc) //TODO
					if _, err := os.Stat(uri); err == nil {
						vals.([]string)[0] = uri
						data[id] = vals
					} else {
						log.Println("WARNING!!! cannot reading", uri, id)
					}
				}

			}
		}
		d := dataIndex{
			k,
			data,
			format,
		}
		di = append(di, d)
	}
	return di, nil
}
