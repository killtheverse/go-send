package main

import (
	"log"
	"os"

	"github.com/urfave/cli"
	"github.com/killtheverse/go-send/src/util"
)

func main() {
	app := cli.NewApp()
	app.Name = "go-send"
	app.Usage = "p2p file sharing application"

	app.Flags = []cli.Flag {
		&cli.StringFlag {
			Name: "file",
			Value: "",
			Usage: "File name to be shared",
		},
		&cli.StringFlag{
			Name: "server",
			Value: "127.0.0.1:8000",
			Usage: "Address of server",
		},
	}

	app.Commands = []*cli.Command {
		{
			Name: "send",
			Usage: "sends a file",
			Action: func(c *cli.Context) error {
				util.GoSend(c.String("file"))
				return nil
			},
		},	
		{
			Name: "recieve",
			Usage: "recieves the file",
			Action: func(c *cli.Context) error {
				util.GoRecv(c.String("file"))
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

