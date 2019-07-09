package data

import (
	"context"
	"log"
	"os"
	path "path/filepath"
	"regexp"

	"github.com/nimezhu/asheets"
	"golang.org/x/oauth2/google"
	sheets "google.golang.org/api/sheets/v4"
)

func parseNamedGSheet(spreadsheetId string, dir string) ([]dataIndex, error) {
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
	gA := asheets.NewGAgent(dir)
	client := gA.GetClient(ctx, config)

	srv, err := sheets.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets Client %v", err)
	}
	_, c := readSheet("Config", srv, spreadsheetId, 1, []int{2}) //header TODO
	index := readNamedIndex(srv, spreadsheetId)
	root, _ := c["root"]

	for _, e := range index {
		g := e.Genome //TODO
		k := e.Id
		v := e.Type
		if k[0] == '#' {
			continue
		}
		var s map[string]interface{}
		var h []string
		h, s = readNamedSheet(k, srv, spreadsheetId, e.Nid, e.Vids) //h == e.Vids
		format := v
		data := make(map[string]interface{})
		if len(e.Vids) == 1 {
			if format == "map" {
				for k0, v0 := range s {
					data[k0] = v0.(string)
				}
			} else { //for file
				for id, loc := range s {
					var uri string
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
				loc := vals.([]string)[0]
				if httpP.MatchString(loc) || httpsP.MatchString(loc) || myFormatP.MatchString(loc) {
					data[id] = objectFactory(h, vals.([]string))
				} else {
					uri = path.Join(root.(string), loc) //TODO
					if _, err := os.Stat(uri); err == nil {
						vals.([]string)[0] = uri
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
