package data

import (
	"encoding/json"
	//"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	. "github.com/nimezhu/indexed/hic"
)

func addHicsHandle(router *mux.Router, hicMap map[string]*HiC) {
	router.HandleFunc("/list", func(w http.ResponseWriter, r *http.Request) {

		/*
		   io.WriteString(w, "Idx\tName\tLength\n")
		   for i, v := range hic.Chr {
		     s := fmt.Sprintf("%d\t%s\t%d\n", i, v.Name, v.Length)
		     io.WriteString(w, s)
		   }
		*/
		keys := []string{}
		for key, _ := range hicMap {
			keys = append(keys, key)
		}
		jsonHic, _ := json.Marshal(keys)
		w.Write(jsonHic)
	})

	router.HandleFunc("/{id}/list", func(w http.ResponseWriter, r *http.Request) {

		/*
		   io.WriteString(w, "Idx\tName\tLength\n")
		   for i, v := range hic.Chr {
		     s := fmt.Sprintf("%d\t%s\t%d\n", i, v.Name, v.Length)
		     io.WriteString(w, s)
		   }
		*/
		params := mux.Vars(r)
		id := params["id"]
		jsonChr, _ := json.Marshal(hicMap[id].Chr)
		w.Write(jsonChr)
	})
	router.HandleFunc("/{id}/norms", func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		id := params["id"]
		normidx := hicMap[id].Footer.NormTypeIdx()
		jsonNormIdx, _ := json.Marshal(normidx)
		w.Write(jsonNormIdx)
	})
	router.HandleFunc("/{id}/units", func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		id := params["id"]
		units := hicMap[id].Footer.Units
		a := []int{}
		for k, _ := range units {
			a = append(a, k)
		}
		jsonUnits, err := json.Marshal(a)
		if err != nil {
			log.Println(err)
		}
		w.Write(jsonUnits)
	})
	router.HandleFunc("/{id}/bpres", func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		id := params["id"]
		bpres := hicMap[id].BpRes
		jsonBpres, err := json.Marshal(bpres)
		if err != nil {
			log.Println(err)
		}
		w.Write(jsonBpres)
	})

	router.HandleFunc("/{id}/info", func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		id := params["id"]
		io.WriteString(w, hicMap[id].String())
	})

	router.HandleFunc("/{id}/get/{chr}:{start}-{end}/{width}/{format}", func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		id := params["id"]
		chr := params["chr"]
		start, _ := strconv.Atoi(params["start"])
		end, _ := strconv.Atoi(params["end"])
		width, _ := strconv.Atoi(params["width"])
		if h, ok := hicMap[id]; ok {
			m, err := h.QueryOne(chr, start, end, width)
			if err != nil {
				io.WriteString(w, err.Error())
			} else {
				format := params["format"]
				if format == "bin" {
					w.Header().Set("Content-Type", "application/octet-stream")
					a := matrixToBytes(m)
					w.Write(a)
				} else {
					io.WriteString(w, sprintMat64(m))
				}
			}

		} else {
			io.WriteString(w, "hic reader not found")
		}

	})
	router.HandleFunc("/{id}/get2d/{chr}:{start}-{end}/{chr2}:{start2}-{end2}/{resIdx}/{format}", func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		id := params["id"]
		chr := params["chr"]
		start, _ := strconv.Atoi(params["start"])
		end, _ := strconv.Atoi(params["end"])
		resIdx, _ := strconv.Atoi(params["resIdx"])
		chr2 := params["chr2"]
		start2, _ := strconv.Atoi(params["start2"])
		end2, _ := strconv.Atoi(params["end2"])
		format := params["format"]
		m, err := hicMap[id].QueryTwo(chr, start, end, chr2, start2, end2, resIdx)
		if err == nil {
			if format == "bin" {
				w.Header().Set("Content-Type", "application/octet-stream")
				a := matrixToBytes(m)
				w.Write(a)
			} else {
				io.WriteString(w, sprintMat64(m))
			}
		} else {
			io.WriteString(w, err.Error())
		}

	})
	router.HandleFunc("/{id}/get2dnorm/{chr}:{start}-{end}/{chr2}:{start2}-{end2}/{resIdx}/{norm}/{unit}/{format}", func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		id := params["id"]
		chr := params["chr"]
		start, _ := strconv.Atoi(params["start"])
		end, _ := strconv.Atoi(params["end"])
		resIdx, _ := strconv.Atoi(params["resIdx"])
		chr2 := params["chr2"]
		start2, _ := strconv.Atoi(params["start2"])
		end2, _ := strconv.Atoi(params["end2"])
		format := params["format"]
		norm, _ := strconv.Atoi(params["norm"])
		unit, _ := strconv.Atoi(params["unit"])
		m, err := hicMap[id].QueryTwoNormMat(chr, start, end, chr2, start2, end2, resIdx, norm, unit)
		if err == nil {
			if format == "bin" {
				w.Header().Set("Content-Type", "application/octet-stream")
				a := matrixToBytes(m)
				w.Write(a)
			} else {
				io.WriteString(w, sprintMat64(m))
			}
		} else {
			io.WriteString(w, err.Error())
		}
	})
	router.HandleFunc("/{id}/get2doe/{chr}:{start}-{end}/{chr2}:{start2}-{end2}/{resIdx}/{norm}/{unit}/{format}", func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		id := params["id"]
		chr := params["chr"]
		start, _ := strconv.Atoi(params["start"])
		end, _ := strconv.Atoi(params["end"])
		resIdx, _ := strconv.Atoi(params["resIdx"])
		chr2 := params["chr2"]
		start2, _ := strconv.Atoi(params["start2"])
		end2, _ := strconv.Atoi(params["end2"])
		format := params["format"]
		norm, _ := strconv.Atoi(params["norm"])
		unit, _ := strconv.Atoi(params["unit"])
		m, err := hicMap[id].QueryOE2(chr, start, end, chr2, start2, end2, norm, unit, resIdx)
		if err == nil {
			if format == "bin" {
				w.Header().Set("Content-Type", "application/octet-stream")
				a := matrixToBytes(m)
				w.Write(a)
			} else {
				io.WriteString(w, sprintMat64_2(m)) //TODO resolution
			}
		} else {
			io.WriteString(w, err.Error())
		}
	})
	router.HandleFunc("/{id}/corrected/{chr}:{start}-{end}/{resIdx}", func(w http.ResponseWriter, r *http.Request) {

		params := mux.Vars(r)
		//chr := params["chr"]
		id := params["id"]
		start, _ := strconv.Atoi(params["start"])
		end, _ := strconv.Atoi(params["end"])
		resIdx, _ := strconv.Atoi(params["resIdx"])
		s, e := hicMap[id].Corrected(start, end, resIdx)
		j, _ := json.Marshal([]int{s, e})
		w.Write(j)
	})

}
