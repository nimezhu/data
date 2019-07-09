package data

import (
	"errors"
	"log"
	"strconv"
	"strings"

	sheets "google.golang.org/api/sheets/v4"
)

/* readSheet is 1-index
 *  TODO: Add More Interfaces Such as translate A,B,C,D or Sheet Header
 *  TODO 101 : Return Column Names
 */
func readSheet(id string, srv *sheets.Service, spreadsheetId string, nameIdx int, valueIdxs []int) ([]string, map[string]interface{}) {
	if len(valueIdxs) == 1 {
		a := make(map[string]interface{})
		m := _readSheetToStringMap(id, srv, spreadsheetId, nameIdx, valueIdxs[0])
		for k, v := range m {
			a[k] = v
		}
		return nil, a
	} else {
		a := make(map[string]interface{})
		h, m := _readSheet(id, srv, spreadsheetId, nameIdx, valueIdxs) //TODO handle header
		for k, v := range m {
			a[k] = v
		}
		return h, a
	}
	return nil, nil
}

func readNamedSheet(id string, srv *sheets.Service, spreadsheetId string, nameID string, valueIDs []string) ([]string, map[string]interface{}) {
	header, err := readSheetHeader(id, srv, spreadsheetId)
	if err != nil {
		return nil, nil
	}
	nameIdx := -1
	for i, d := range header {
		if strings.EqualFold(d, nameID) {
			nameIdx = i
			break
		}
	}
	if nameIdx < 0 {
		return nil, nil
	}
	valueIdxs := make([]int, len(valueIDs))
	for i, v := range valueIDs {
		valueIdxs[i] = -1
		for j, d := range header {
			if strings.EqualFold(d, v) {
				valueIdxs[i] = j
				break
			}
		}
		if valueIdxs[i] < 0 {
			return nil, nil
		}
	}
	return readSheet(id, srv, spreadsheetId, nameIdx, valueIdxs)

}
func readSheetHeader(id string, srv *sheets.Service, spreadsheetId string) ([]string, error) {
	readRange := id + "!A1:ZZZ1" //TODO: READ FIRST LINE AS NAMES
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet. %v", err)
		return nil, err
	}
	if len(resp.Values) > 0 {
		for i0, row := range resp.Values {
			if i0 == 0 {
				header := make([]string, len(row))
				for i, d := range row {
					header[i] = d.(string)
				}
				return header, nil
			}
		}
	}
	return nil, errors.New("empty sheet")
}

/*
 * TODO:
 */
func _readSheet(id string, srv *sheets.Service, spreadsheetId string, nameIdx int, valueIdxs []int) ([]string, map[string][]string) {
	readRange := id + "!A1:ZZZ" //TODO: READ FIRST LINE AS NAMES
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet. %v", err)
	}
	header := make([]string, len(valueIdxs))
	r := make(map[string][]string)
	l := len(valueIdxs)
	if len(resp.Values) > 0 {
		for i0, row := range resp.Values {
			if i0 == 0 {
				//handle header
				for i, k := range valueIdxs {
					header[i] = row[k-1].(string)
				}
			} else {
				s := make([]string, l)
				for i, k := range valueIdxs {
					s[i] = row[k-1].(string)
				}
				r[row[nameIdx-1].(string)] = s
			}
		}
	} else {
		log.Println("No data found.")
	}
	return header, r
}

/* readSheetToStringMap is 1-index */
func _readSheetToStringMap(id string, srv *sheets.Service, spreadsheetId string, nameIdx int, valueIdx int) map[string]string {
	readRange := id + "!A2:ZZ"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet. %v", err)
	}
	log.Println("Reading ", id)
	r := make(map[string]string)
	if len(resp.Values) > 0 {
		for _, row := range resp.Values {
			r[row[nameIdx-1].(string)] = row[valueIdx-1].(string)
		}
	} else {
		log.Println("No data found.")
	}
	return r
}

/*TODO THIS add more interface */
type IndexEntry struct {
	Genome string
	Id     string
	Type   string
	Nc     int
	Vc     []int
}

type NamedIndexEntry struct {
	Genome string
	Id     string
	Type   string
	Nid    string
	Vids   []string
}

func readNamedIndex(srv *sheets.Service, spreadsheetId string) []NamedIndexEntry {
	id := "Index"
	readRange := id + "!A2:E"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet. %v", err)
	}
	a := make([]NamedIndexEntry, len(resp.Values))
	if len(resp.Values) > 0 {
		for i, row := range resp.Values {
			ns := row[3].(string)
			vs := strings.Split(row[4].(string), ",")
			a[i] = NamedIndexEntry{row[0].(string), row[1].(string), row[2].(string), ns, vs}
		}
	} else {
		log.Println("No data found.")
		return nil
	}
	return a
}

/* TODO THIS Add string []string inferface */

func readIndex(srv *sheets.Service, spreadsheetId string) []IndexEntry {
	id := "Index"
	readRange := id + "!A2:E"
	resp, err := srv.Spreadsheets.Values.Get(spreadsheetId, readRange).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet. %v", err)
	}
	a := make([]IndexEntry, len(resp.Values))
	if len(resp.Values) > 0 {
		for i, row := range resp.Values {
			ns := row[3].(string)
			nc, err := strconv.Atoi(ns)
			//TODO
			if err != nil {
				nc = colNameToNumber(ns)
			}
			vs := strings.Split(row[4].(string), ",")
			vc := make([]int, len(vs))
			for i, v := range vs {
				vc[i], err = strconv.Atoi(v)
				if err != nil {
					vc[i] = colNameToNumber(v)
				}
			}
			a[i] = IndexEntry{row[0].(string), row[1].(string), row[2].(string), nc, vc}
		}
	} else {
		log.Println("No data found.")
	}
	return a
}
