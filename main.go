package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"text/template"

	"github.com/BurntSushi/toml"
	"github.com/fgerling/github-cards/internal/config"
	obs "github.com/fgerling/gobs"
)

func main() {
	var config_file = flag.String("conf", "", "Set the config file.")
	var group = flag.String("group", "", "Set the group to search for.")
	var user = flag.String("user", "", "Set the obs user.")
	var password = flag.String("password", "", "Set the password.")

	flag.Parse()
	if *config_file == "" {
		*config_file = "./config.toml"
		log.Printf("Config file: %q\n", *config_file)
	}
	var conf config.Config
	dat, err := ioutil.ReadFile(*config_file)
	if err != nil {
		log.Printf("%+v", err)
	} else {
		_, err = toml.Decode(string(dat), &conf)
		if err != nil {
			panic(err)
		}
	}
	if *user == "" {
		user = &conf.Username
	}
	log.Printf("User: %q\n", *user)
	if *password == "" {
		password = &conf.Password
	}
	if *group == "" {
		group = &conf.Group
	}
	log.Printf("Group: %q\n", *group)

	var rrs []obs.ReleaseRequest
	client := obs.NewClient(*user, *password)
	rrs, err = client.GetReleaseRequests(*group, "new,review")
	if err != nil {
		log.Fatal(err)
	}
	type TmplStruct struct {
		Request obs.ReleaseRequest
		Summary string
	}
	tmpl, err := template.New("list-requests").Parse("{{if eq .Request.Priority \"important\"}}!{{else}} {{end}} RR#{{.Request.Id}} {{.Summary}}\nhttps://maintenance.suse.de/request/{{.Request.Id}}\n\n")
	if err != nil {
		log.Fatal(err)
	}

	for _, request := range rrs {
		patchinfo, err := client.GetPatchinfo(request)
		if err != nil {
			log.Fatal(err)
		}
		err = tmpl.Execute(os.Stdout, TmplStruct{Request: request, Summary: patchinfo.Summary})
		if err != nil {
			log.Fatal(err)
		}
	}
}
