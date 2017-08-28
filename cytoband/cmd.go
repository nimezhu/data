package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/nimezhu/data"
	"github.com/urfave/cli"
)

const VERSION = "0.0.0"

func main() {
	app := cli.NewApp()
	app.Version = VERSION
	app.Name = "cytoband"
	app.Usage = "cytoband demo server"
	app.EnableBashCompletion = true //TODO
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "Show more output",
		},
	}

	// Commands
	app.Commands = []cli.Command{
		{
			Name:   "start",
			Usage:  "start a server",
			Action: CmdStart,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "input,i",
					Usage: "input genome strings",
					Value: "hg19,mm10",
				},
				cli.IntFlag{
					Name:  "port,p",
					Usage: "data server port",
					Value: 5000,
				},
			},
		},
	}

	app.Run(os.Args)
}
func CmdStart(c *cli.Context) error {
	genomes := c.String("input")
	port := c.Int("port")
	router := mux.NewRouter()
	m := data.NewCytoBandManager("band")
	gs := strings.Split(genomes, ",")
	for _, v := range gs {
		m.Add(v)
	}
	m.ServeTo(router)
	log.Println("Listening...")
	log.Println("Please open http://127.0.0.1:" + strconv.Itoa(port))
	http.ListenAndServe(":"+strconv.Itoa(port), router)
	return nil
}
