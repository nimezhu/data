package data

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"

	"golang.org/x/oauth2/google"
	sheets "google.golang.org/api/sheets/v4"
)

func parseGSheet(spreadsheetId string) ([]DataIndex, error) {
	di := []DataIndex{}
	httpP, _ := regexp.Compile("^http://")
	httpsP, _ := regexp.Compile("^https://")

	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json") //TODO
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
	c := readSheet("Config", srv, spreadsheetId, 0, 1)
	index := readIndex(srv, spreadsheetId)
	root, _ := c["root"]

	for _, e := range index {
		k := e.Id
		v := e.Type
		if k[0] == '#' {
			continue
		}
		fmt.Println("sheet:", k) //TODO
		fmt.Println("format:", v)
		//TODO
		fmt.Println(e.Nc, e.Vc)
		//TODO add Interface For Complicated Data Input multi Columns.
		s := readSheet(k, srv, spreadsheetId, e.Nc-1, e.Vc-1) //TODO GET FROM INDEX
		format := v
		data := make(map[string]string)
		if v == "map" {
			for k0, v0 := range s {
				data[k0] = v0
			}
		} else {
			for id, loc := range s {
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
		}
		d := DataIndex{
			k,
			data,
			format,
		}
		di = append(di, d)
	}
	return di, nil
}
