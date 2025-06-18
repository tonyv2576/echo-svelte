package main

import (
	"flag"
	"fmt"
	"log"
)

var project ProjectConfig

func init() {
	flag.BoolVar(&project.DevMode, "dev", false, "runs the frontend and backend with hot-reloading")
	flag.StringVar(&project.Name, "name", "", "the title of your project")
	flag.BoolVar(&project.Typescript, "ts", false, "enables typescript")
	flag.StringVar(&project.Directory, "dir", "", "the path to your project")
	flag.IntVar(&project.Port, "port", 8080, "the port to the http server")
	flag.Parse()
}

func main() {
	if err := project.checkForRights(); err != nil {
		log.Fatal(err)
	}
	if project.DevMode {
		project.RunDevMode()
	} else {
		if err := project.Validate(); err != nil {
			log.Fatal(err)
		}
		if err := project.InstallModule(); err != nil {
			log.Fatal(err)
		}
		if err := project.ApplyDefaults(); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Completed")
	}
}
