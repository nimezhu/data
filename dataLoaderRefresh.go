package data

import (
	"errors"
	"log"
)

func (e *Loader) Refresh(dbname string, data interface{}, format string) error {
	if format != "track" {
		log.Println("not support refresh this format yet, format:", format)
		return errors.New("not support this format refresh yet")
	}
	if r, ok := e.Data[dbname]; ok {
		log.Println("refresh ", dbname)
		newdata := data.(map[string]interface{})
		newr := r.(*TrackManager)
		/* Delete OLD inteface could be string or map[string]interface */
		for _, k := range newr.List() {
			if _, ok := newdata[k]; !ok {
				err := newr.Del(k)
				if err != nil {
					log.Println("error in delete", k)
				}
			}
		}
		/* Update New */
		for k, v := range newdata {
			switch v.(type) {
			case string:
				v0 := v.(string)
				if v1, ok := newr.Get(k); !ok {
					newr.AddURI(v0, k)
				} else if v1 != v0 {
					newr.Del(k)
					newr.AddURI(v0, k)
				}
			case map[string]interface{}:
				//TODO
				v0 := v.(map[string]interface{})
				if v1, ok := newr.GetAttr(k); !ok {
					if uri, ok := v0["uri"]; ok {
						newr.AddURI(uri.(string), k)
						newr.SetAttr(k, v0)
					}
				} else {
					if uri, ok := v0["uri"]; ok {
						if olduri, ok := v1["uri"].(string); ok {
							if olduri == uri.(string) {
								newr.SetAttr(k, v0)
							} else {
								newr.Del(k)
								newr.AddURI(uri.(string), k)
								newr.SetAttr(k, v0)
							}
						}
					}

				}

			}

		}
		log.Println("refresh done")
	} else {
		log.Println("not found db", dbname)
		return errors.New("db not found")
	}

	return nil
}
