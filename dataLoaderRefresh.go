package data

import (
	"errors"
	"log"
)

func (e *Loader) Refresh(dbname string, data interface{}, format string) error {
	if format != "track" {
		log.Println("not support refresh this format ", format)
		return errors.New("not support this format refresh yet")
	}
	if r, ok := e.Data[dbname]; ok {
		log.Println("refresh ", dbname)
		newdata := data.(map[string]interface{})
		newr := r.(*TrackManager2)
		/* Delete OLD */
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
			v0 := v.(string)
			/* if not loaded before */
			if v1, ok := newr.Get(k); !ok {
				newr.AddURI(v0, k)
			} else {
				/* if uri changed  */
				if v1 != v0 {
					newr.Del(k)
					newr.AddURI(v0, k)
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
