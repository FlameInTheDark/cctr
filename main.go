package main

import (
	"context"
	"log"
	"os"

	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:        "cctr",
		Usage:       "Half-Life closed captions translator",
		Description: "Translations provided by Deepl. Requires API key to work.",
		Commands: []*cli.Command{
			translate(),
			count(),
			languages(),
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}

}
