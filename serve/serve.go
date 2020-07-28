package main

import "net/http"
import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"os"
	"path/filepath"
)

func main() {
	app := cli.NewApp()
	app.Name = "serve"
	app.Usage = "Starts HTTP File Server for given host, port and directory"
	app.Version = "1.0"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:    "address",
			Aliases: []string{"a"},
			Usage:   "Host for the HTTP server",
			Value:   "0.0.0.0",
		}, &cli.IntFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Usage:   "Port for the HTTP server",
			Value:   8080,
		}, &cli.StringFlag{
			Name:    "directory",
			Aliases: []string{"d"},
			Usage:   "Root directory for the HTTP server",
			Value:   ".",
		},
	}
	app.Action = func(c *cli.Context) (ret error) {
		var directory string
		if directory, ret = filepath.Abs(c.String("d")); ret != nil {
			panic(ret)
		}
		ip := c.String("address")
		port := c.Int("port")
		log.Print(fmt.Sprintf("Start server %v:%v for %v", ip, port, directory))
		http.Handle("/", http.FileServer(http.Dir(directory)))
		panic(http.ListenAndServe(fmt.Sprintf("%v:%v", ip, port), nil))
		return
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
