package data

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type Loader struct {
	IndexRoot string
	Plugins   map[string]func(string, interface{}) (DataRouter, error)
	Data      map[string]DataRouter
	entry     []string
	jdata     []map[string]string
	gdb       map[string][]map[string]string
}

var (
	loaders = map[string]func(string, interface{}) (DataRouter, error){
		"file":  _fileLoader,
		"hic":   _hicLoader,
		"map":   _mapLoader,
		"tabix": _tabixLoader,
	}
)

func (e *Loader) AddLoader(format string, f func(string, interface{}) (DataRouter, error)) error {
	//TODO
	_, ok := loaders[format]
	if ok || format == "bigwig" || format == "bigbed" || format == "track" {
		return errors.New("keyword reserved")
	}
	e.Plugins[format] = f
	return nil
}
func NewLoader(root string) *Loader {
	return &Loader{root, make(map[string]func(string, interface{}) (DataRouter, error)), make(map[string]DataRouter), make([]string, 0, 0), make([]map[string]string, 0, 0), make(map[string][]map[string]string)}
}

func (e *Loader) Factory(dbname string, data interface{}, format string) func(string, interface{}) (DataRouter, error) {
	if f0, ok := e.Plugins[format]; ok {
		return f0
	}
	if f, ok := loaders[format]; ok {
		return f
	}
	switch format {
	case "bigwig": //bigwig with buffer
		return func(dbname string, data interface{}) (DataRouter, error) {
			switch v := data.(type) {
			default:
				log.Printf("unexpected type %T\n", v)
				return nil, errors.New(fmt.Sprintf("bigwig format not support type %T", v))
			case string:
				return NewBigWigManager(data.(string), dbname, e.IndexRoot), nil
			case map[string]interface{}:
				a := InitBigWigManager(dbname, e.IndexRoot)
				for key, val := range data.(map[string]interface{}) {
					switch val.(type) {
					case string:
						a.AddURI(val.(string), key)
						a.SetAttr(key, map[string]interface{}{"uri": val})
					case map[string]interface{}:
						if uri, ok := val.(map[string]interface{})["uri"]; ok {
							a.AddURI(uri.(string), key)
							a.SetAttr(key, val.(map[string]interface{}))
						}
					}
				}
				return a, nil
			}
		}
	case "bigbed":
		return func(dbname string, data interface{}) (DataRouter, error) {
			switch v := data.(type) {
			default:
				fmt.Printf("unexpected type %T", v)
				return nil, errors.New(fmt.Sprintf("bigwig format not support type %T", v))
			case string:
				return NewBigBedManager(data.(string), dbname, e.IndexRoot), nil
			case map[string]interface{}:
				a := InitBigBedManager(dbname, e.IndexRoot)
				for key, val := range data.(map[string]interface{}) {
					switch val.(type) {
					case string:
						a.AddURI(val.(string), key)
						a.SetAttr(key, map[string]interface{}{"uri": val})
					case map[string]interface{}:
						if uri, ok := val.(map[string]interface{})["uri"]; ok {
							a.AddURI(uri.(string), key)
							//TODO Add Set Attr API
							a.SetAttr(key, val.(map[string]interface{}))
						}
					}
				}
				return a, nil
			}
		}
	case "track":
		return func(dbname string, data interface{}) (DataRouter, error) {
			switch v := data.(type) {
			default:
				log.Printf("unexpected type %T\n", v)
				return nil, errors.New(fmt.Sprintf("bigwig format not support type %T", v))
			case string:
				return NewTrackManager(data.(string), dbname, e.IndexRoot), nil
			case map[string]interface{}:
				a := InitTrackManager(dbname, e.IndexRoot)
				for key, val := range data.(map[string]interface{}) {
					switch val.(type) {
					case string:
						a.AddURI(val.(string), key)
						a.SetAttr(key, map[string]interface{}{"uri": val})
					case map[string]interface{}:
						if uri, ok := val.(map[string]interface{})["uri"]; ok {
							a.AddURI(uri.(string), key)
							a.SetAttr(key, val.(map[string]interface{}))
						}
					}
				}
				return a, nil
			}
		}

	}

	return nil
}

func _fileLoader(dbname string, data interface{}) (DataRouter, error) {
	switch v := data.(type) {
	default:
		return nil, errors.New(fmt.Sprintf("unexpected type %T", v))
	case string:
		log.Println("in file load string", data.(string))
		return NewFileManager(data.(string), dbname), nil
	case map[string]interface{}:
		m := InitFileManager(dbname)
		for k, v := range data.(map[string]interface{}) {
			m.AddURI(v.(string), k)
		}
		return m, nil
	}
}

func _hicLoader(dbname string, data interface{}) (DataRouter, error) {
	switch v := data.(type) {
	default:
		log.Printf("unexpected type %T\n", v)
		return nil, errors.New(fmt.Sprintf("hic format not support type %T", v))
	case string:
		return NewHicManager(data.(string), dbname), nil
	case map[string]interface{}:
		a := InitHicManager(dbname)
		for key, val := range data.(map[string]interface{}) {
			switch val.(type) {
			case string:
				a.AddURI(val.(string), key)
			case map[string]interface{}:
				if uri, ok := val.(map[string]interface{})["uri"]; ok {
					a.AddURI(uri.(string), key)
					a.SetAttr(key, val.(map[string]interface{}))
				}
			}
		}
		return a, nil
	}
}

func _mapLoader(dbname string, data interface{}) (DataRouter, error) {
	switch v := data.(type) {
	default:
		log.Printf("unexpected type %T\n", v)
		return nil, errors.New(fmt.Sprintf("map format not support type %T", v))
	case string:
		return NewMapManager(data.(string), dbname), nil
	case map[string]interface{}:
		a := InitMapManager(dbname)
		for key, val := range data.(map[string]interface{}) {
			a.AddURI(val.(string), key)
		}
		return a, nil
	}
}
func _tabixLoader(dbname string, data interface{}) (DataRouter, error) {
	switch v := data.(type) {
	default:
		log.Printf("unexpected type %T\n", v)
		return nil, errors.New(fmt.Sprintf("tabix format not support type %T", v))
	case string:
		return NewTabixManager(data.(string), dbname), nil //TODO
	case map[string]interface{}:
		a := InitTabixManager(dbname)
		for key, val := range data.(map[string]interface{}) {
			a.AddURI(val.(string), key)
		}
		return a, nil
	}
}

func parseRegions(txt string) []Bed4 {
	l := strings.Split(txt, ";")
	b := make([]Bed4, len(l))
	for i, v := range l {
		b[i] = parseRegion(v)
	}
	return b
}
func parseRegion(txt string) Bed4 {
	x := strings.Split(txt, ":")
	chr := x[0]
	se := strings.Split(x[1], "-")
	start, _ := strconv.Atoi(se[0])
	end, _ := strconv.Atoi(se[1])
	return Bed4{chr, start, end, "."}
}
