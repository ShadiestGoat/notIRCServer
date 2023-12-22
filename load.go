package main

import "flag"

var (
	userEnv,
	dbFile,
	envFile,
	outFile,
	asUser *string
)

func init() {
	userEnv = flag.String("u", "users.yaml", "The location of the users config")
	dbFile = flag.String("s", "db.shitdb", "The location of messages db")
	envFile = flag.String("e", ".env", "The location of the env file")
	outFile = flag.String("o", "-", "The location to output to")
	asUser = flag.String("a", "", "The location to output to")
}
