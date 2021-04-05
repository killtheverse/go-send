package main

import (
	"log"
	"os"

	"github.com/urfave/cli"

	"github.com/killtheverse/go-send/src/goSend"
	"github.com/killtheverse/go-send/src/goRecv"
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
		&cli.StringFlag{
			Name: "port",
			Value: ":9000",
			Usage: "Port no.",
		},
		&cli.StringFlag{
			Name: "tcpport",
			Value: ":9001",
			Usage: "Tcp port no.",
		},
	}

	app.Commands = []*cli.Command {
		{
			Name: "send",
			Usage: "sends a file",
			Action: func(c *cli.Context) error {
				goSend.GoSend(c.String("file"), c.String("server"), c.String("port"))
				return nil
			},
		},	
		{
			Name: "recieve",
			Usage: "recieves the file",
			Action: func(c *cli.Context) error {
				goRecv.GoRecv(c.String("file"), c.String("server"), c.String("port"), c.String("tcpport"))
				return nil
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}