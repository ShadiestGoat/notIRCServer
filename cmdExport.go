package main

import (
	"io"
	"os"

	"github.com/shadiestgoat/log"
	"github.com/shadiestgoat/notIRCServer/export"
	"github.com/shadiestgoat/notIRCServer/users"
)

func cmdExport() {
	format := "log"
	if len(cmds) >= 2 {
		format = cmds[1]
	}

	var u *users.User

	if *asUser == "" {
		u = &users.User{
			AbleToReadAllWhispers: true,
		}
	} else if *asUser == "*" {
		u = &users.User{
			ReadWhispers: map[string]bool{},
		}
	} else {
		u = users.GetUser(*asUser)
		if u == nil {
			log.Fatal("User '%v' not found", *asUser)
		}
	}

	var out io.WriteCloser

	if *outFile == "-" {
		out = os.Stdout
	} else {
		f, err := os.OpenFile(*outFile, os.O_WRONLY, 0755)
		log.FatalIfErr(err, "opening output file")
		out = f
	}

	err := export.Export(format, u, out)

	if *outFile != "-" {
		out.Close()
	}

	log.FatalIfErr(err, "exporting")
}
