package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"

	"github.com/gorilla/mux"
	. "github.com/nimezhu/indexed/bbi"
)

func AddBwsHandle(router *mux.Router, bwMap map[string]*BigWigReader) {
	router.HandleFunc("/{id}/get/{chr}:{start}-{end}/{width}", func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		chr := params["chr"]
		id := params["id"]
		start, _ := strconv.Atoi(params["start"])
		end, _ := strconv.Atoi(params["end"])
		width, _ := strconv.Atoi(params["width"])
		bw, ok := bwMap[id]
		if !ok {
			io.WriteString(w, id+" not found")
		} else {
			if iter, err := bw.Query(chr, start, end, width); err == nil {
				for i := range iter {
					io.WriteString(w, fmt.Sprintln(i.From, "\t", i.To, "\t", i.Sum))
				}
			}
		}
	})
	router.HandleFunc("/{id}/getjson/{chr}:{start}-{end}/{width}", func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		id := params["id"]
		chr := params["chr"]
		start, _ := strconv.Atoi(params["start"])
		end, _ := strconv.Atoi(params["end"])
		width, _ := strconv.Atoi(params["width"])
		bw, ok := bwMap[id]
		if !ok {
			io.WriteString(w, id+" not found")
		} else {
			arr := []*BbiQueryType{}
			if iter, err := bw.Query(chr, start, end, width); err == nil {
				for i := range iter {
					arr = append(arr, i)
					//io.WriteString(w, fmt.Sprintln(i.From, "\t", i.To, "\t", i.Sum))
				}
			}
			j, err := json.Marshal(arr)
			checkErr(err)
			io.WriteString(w, string(j))
		}
	})
	router.HandleFunc("/{id}/get/{chr}:{start}-{end}", func(w http.ResponseWriter, r *http.Request) { //BinSize Corrected.

		params := mux.Vars(r)
		id := params["id"]
		chr := params["chr"]
		start, _ := strconv.Atoi(params["start"])
		end, _ := strconv.Atoi(params["end"])
		bw, ok := bwMap[id]
		if !ok {
			io.WriteString(w, id+" not found")
		} else {
			arr := []*BbiBlockDecoderType{}
			if iter, err := bw.QueryRaw(chr, start, end); err == nil {
				for i := range iter {
					/*
						if i.To == 0 {
							break
						} //TODO DEBUG THIS FOR STREAM
					*/
					arr = append(arr, i)
					//io.WriteString(w, fmt.Sprintln(i.From, "\t", i.To, "\t", i.Sum))
				}
			}
			j, err := json.Marshal(arr)
			checkErr(err)
			io.WriteString(w, string(j))
		}
	})
	router.HandleFunc("/{id}/getbin/{chr}:{start}-{end}/{binsize}", func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		id := params["id"]
		chr := params["chr"]
		start, _ := strconv.Atoi(params["start"])
		end, _ := strconv.Atoi(params["end"])
		binsize, _ := strconv.Atoi(params["binsize"])
		bw, ok := bwMap[id]
		if !ok {
			io.WriteString(w, id+" not found")
		} else {
			arr := []*BbiQueryType{}
			if iter, err := bw.QueryBin(chr, start, end, binsize); err == nil {
				for i := range iter {
					/*
						if i.To == 0 {
							break
						} //TODO DEBUG THIS FOR STREAM
					*/
					arr = append(arr, i)
					//io.WriteString(w, fmt.Sprintln(i.From, "\t", i.To, "\t", i.Sum))
				}
			}
			j, err := json.Marshal(arr)
			checkErr(err)
			io.WriteString(w, string(j))
		}
	})
	router.HandleFunc("/{id}/list", func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		id := params["id"]
		bw, ok := bwMap[id]
		if !ok {
			io.WriteString(w, id+" not found")
		} else {
			j, err := json.Marshal(bw.Genome)
			checkErr(err)
			io.WriteString(w, string(j))
		}
	})
	router.HandleFunc("/{id}/binsize/{length}/{width}", func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		id := params["id"]
		length, _ := strconv.Atoi(params["length"])
		width, _ := strconv.Atoi(params["width"])
		bw, ok := bwMap[id]
		if !ok {
			io.WriteString(w, id+" not found")
		} else {
			binsize := bw.GetBinsize(length, width)
			j, err := json.Marshal(binsize)
			checkErr(err)
			io.WriteString(w, string(j))
		}
	})
	router.HandleFunc("/{id}/binsizes", func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		id := params["id"]
		bw, ok := bwMap[id]
		if !ok {
			io.WriteString(w, id+" not found")
		} else {
			j, err := json.Marshal(bw.Binsizes())
			checkErr(err)
			io.WriteString(w, string(j))
		}
	})
	router.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {

		var keys []string
		for k := range bwMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		j, err := json.Marshal(keys)
		checkErr(err)
		io.WriteString(w, string(j))
	})
}
