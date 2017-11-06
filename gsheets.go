package data

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/oauth2"
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

/* readSheet is 1-index */
func readSheet(id string, srv *sheets.Service, spreadsheetId string, nameIdx int, valueIdxs []int) map[string]interface{} {
	if len(valueIdxs) == 1 {
		a := make(map[string]interface{})
		m := _readSheetToStringMap(id, srv, spreadsheetId, nameIdx, valueIdxs[0])
		for k, v := range m {
			a[k] = v
		}
		return a
	} else {
		a := make(map[string]interface{})
		m := _readSheet(id, srv, spreadsheetId, nameIdx, valueIdxs)
		for k, v := range m {
			a[k] = v
		}
		return a
	}
	return nil
}
func _readSheet(id string, srv *sheets.Service, spreadsheetId string, nameIdx int, valueIdxs []int) map[string][]string {
	readRange := id + "!A2:ZZ"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet. %v", err)
	}
	fmt.Println("Reading ", id)
	r := make(map[string][]string)
	l := len(valueIdxs)
	if len(resp.Values) > 0 {
		for _, row := range resp.Values {
			// Print columns A and E, which correspond to indices 0 and 4.
			s := make([]string, l)
			for i, k := range valueIdxs {
				s[i] = row[k-1].(string)
			}
			r[row[nameIdx-1].(string)] = s
		}
	} else {
		fmt.Print("No data found.")
	}
	fmt.Println(r)
	return r
}

/* readSheetToStringMap is 1-index */
func _readSheetToStringMap(id string, srv *sheets.Service, spreadsheetId string, nameIdx int, valueIdx int) map[string]string {
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
			r[row[nameIdx-1].(string)] = row[valueIdx-1].(string)
		}
	} else {
		fmt.Print("No data found.")
	}
	return r
}

/*TODO THIS add more interface */
type IndexEntry struct {
	Id   string
	Type string
	Nc   int
	Vc   []int
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
			//vc, _ := strconv.Atoi(row[3].(string))
			//TODO
			vs := strings.Split(row[3].(string), ",")
			vc := make([]int, len(vs))
			for i, v := range vs {
				vc[i], _ = strconv.Atoi(v)
			}
			a[i] = IndexEntry{row[0].(string), row[1].(string), nc, vc}
		}
	} else {
		fmt.Print("No data found.")
	}
	return a
}
