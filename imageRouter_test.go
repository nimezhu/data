package data

import (
	"log"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
)

func TestGetRoot(t *testing.T) {
	roots := []string{"/home/zhuxp/1", "/home/zhuxp/2", "/home/zhuxp/image/3"}
	r := getRootDir(roots)
	t.Log(r)
}

func TestImageRouter(t *testing.T) {
	image := bedImage{"earth", "/home/zhuxp/mylib/data/planets/earth.png", []Bed3{Bed3{"chr1", 1, 10000}}}
	image2 := bedImage{"mars", "/home/zhuxp/mylib/data/planets/mars.png", []Bed3{Bed3{"chr1", 1, 10000}}}
	image3 := bedImage{"sun", "/home/zhuxp/mylib/data/planets/sun/sun.png", []Bed3{Bed3{"chr1", 1, 10000}}}
	s := InitBinindexImageRouter("planets")
	s.Add(image)
	s.Add(image2)
	s.Add(image3)
	router := mux.NewRouter()
	s.ServeTo(router)
	log.Fatal(http.ListenAndServe(":8082", router))
}
