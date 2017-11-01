package data

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	sheets "google.golang.org/api/sheets/v4"
)

func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("sheets.googleapis.com-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func readSheet(id string, srv *sheets.Service, spreadsheetId string, nameIdx int, valueIdx int) map[string]string {
	readRange := id + "!A2:ZZ"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet. %v", err)
	}
	fmt.Println("Reading ", id)
	r := make(map[string]string)
	if len(resp.Values) > 0 {
		for _, row := range resp.Values {
			// Print columns A and E, which correspond to indices 0 and 4.
			r[row[nameIdx].(string)] = row[valueIdx].(string)
		}
	} else {
		fmt.Print("No data found.")
	}
	return r
}

type IndexEntry struct {
	Id   string
	Type string
	Nc   int
	Vc   int
}

func readIndex(srv *sheets.Service, spreadsheetId string) []IndexEntry {
	id := "Index"
	readRange := id + "!A2:D"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet. %v", err)
	}
	a := make([]IndexEntry, len(resp.Values))
	if len(resp.Values) > 0 {
		for i, row := range resp.Values {
			// Print columns A and E, which correspond to indices 0 and 4.
			//r[row[nameIdx].(string)] = row[valueIdx].(string)
			nc, _ := strconv.Atoi(row[2].(string))
			vc, _ := strconv.Atoi(row[3].(string))
			a[i] = IndexEntry{row[0].(string), row[1].(string), nc, vc}
		}
	} else {
		fmt.Print("No data found.")
	}
	return a
}

func AddGSheets(spreadsheetId string, clientSecretJson string, router *mux.Router) map[string]DataManager {
	m := map[string]DataManager{}
	entry := []string{}
	jdata := []map[string]string{}
	httpP, _ := regexp.Compile("^http://")
	httpsP, _ := regexp.Compile("^https://")

	ctx := context.Background()

	//b, err := ioutil.ReadFile("client_secret.json")
	b, err := ioutil.ReadFile(clientSecretJson)
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
	//server, _ := c["server"]
	/*
		router.HandleFunc("/server/ls", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("{ \"main\":\"" + server + "\"}"))
		}) //TODO FIX
	*/
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
		s := readSheet(k, srv, spreadsheetId, e.Nc-1, e.Vc-1) //TODO GET FROM INDEX
		format := v
		a := initDataManager(k, format) //TODO
		jdata = append(jdata, map[string]string{
			"dbname": k,
			"uri":    spreadsheetId + ":" + k,
			"format": format,
		})
		entry = append(entry, k)
		fmt.Println(s)
		if v == "map" {
			for k0, v0 := range s {
				a.AddURI(v0, k0)
			}
		} else {
			for id, loc := range s {
				var uri string
				if httpP.MatchString(loc) || httpsP.MatchString(loc) {
					uri = loc
					a.AddURI(uri, id)
				} else {
					uri = path.Join(root, loc) //TODO
					if _, err := os.Stat(uri); err == nil {
						a.AddURI(uri, id)
					} else {
						log.Println("WARNING!!! cannot reading", uri, id)
					}
				}

			}
		}
		a.ServeTo(router)
		m[k] = a
		fmt.Println("")
	}

	router.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {
		e, _ := json.Marshal(entry)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(e)
	})
	router.HandleFunc("/ls", func(w http.ResponseWriter, r *http.Request) {
		e, _ := json.Marshal(jdata)
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Write(e)
	})
	return m
}
