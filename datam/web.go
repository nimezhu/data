package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

//go:generate go-bindata-assetfs -pkg main index.html
func AddStaticHandle(router *mux.Router) {
	//TODO FIX
	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Accept-Encoding")
		bytes, _ := Asset("index.html")
		w.Write(bytes)
	})
}
