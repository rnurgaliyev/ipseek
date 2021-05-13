package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/integrii/flaggy"
	"gopkg.in/yaml.v2"
)

type configuration struct {
	UpdateGroups []struct {
		Name    string             `yaml:"name"`
		Objects []updateObjectData `yaml:"objects"`
	} `yaml:"update_groups"`
}

type updateObjectData struct {
	Name          string            `yaml:"name"`
	Type          string            `yaml:"type"`
	Configuration map[string]string `yaml:"configuration"`
}

var config configuration

const maxAttempts = 5

func updateObject(o updateObjectData, group string, address string) {
	log.Printf("[INFO] Updating object %s@%s\n", o.Name, group)
	switch o.Type {
	case "openstack_ipsec_site_connection":
		for i := 0; i < maxAttempts; i++ {
			if updateOpenstackIpsecSiteConnection(o.Configuration, address) {
				log.Printf("[INFO] Successfully updated object %s@%s\n", o.Name, group)
				break
			} else {
				log.Printf("[WARNING] Failed to update object %s@%s (attempt %d)\n", o.Name, group, i+1)
				if i+1 < maxAttempts {
					time.Sleep(5 * time.Second)
				} else {
					log.Printf("[ERROR] Failed to update object %s@%s after 5 attempts. Giving up.\n", o.Name, group)
				}
			}
		}
	default:
		log.Printf("[ERROR] Don't know how to handle type %s of object %s@%s",
			o.Type, o.Name, group)
	}
}

func httpUpdate(w http.ResponseWriter, r *http.Request) {
	group, ok := r.URL.Query()["group"]
	if !ok {
		log.Printf("[WARNING] Received wrong update request %s from %s\n",
			r.RequestURI, r.RemoteAddr)
		response := []byte("Update group is not specified\n")
		w.WriteHeader(400)
		w.Write(response)
		return
	}

	address, ok := r.URL.Query()["address"]
	if !ok {
		log.Printf("[WARNING] Received wrong update request %s from %s\n",
			r.RequestURI, r.RemoteAddr)
		response := []byte("IP address is not specified\n")
		w.WriteHeader(400)
		w.Write(response)
		return
	}

	for _, v := range config.UpdateGroups {
		if v.Name == group[0] {
			log.Printf("[INFO] Scheduling update of group \"%s\" with %s, received from %s\n",
				group[0], address[0], r.RemoteAddr)

			for _, o := range v.Objects {
				go updateObject(o, group[0], address[0])
			}

			response := []byte("Update request accepted\n")

			w.WriteHeader(204)
			w.Write(response)

			return
		}
	}

	log.Printf("[WARNING] Received wrong update request %s from %s\n",
		r.RequestURI, r.RemoteAddr)
	response := []byte("Update group is not found\n")
	w.WriteHeader(404)
	w.Write(response)
}

func main() {
	var configFileFlag = "ipseek.yml"
	var bindFlag = "0.0.0.0"
	var portFlag = 8088

	flaggy.String(&configFileFlag, "c", "config", "File to read configuration from")
	flaggy.String(&bindFlag, "b", "bind", "Address to listen on")
	flaggy.Int(&portFlag, "p", "port", "Port number to listen on")
	flaggy.DefaultParser.DisableShowVersionWithVersion()
	flaggy.Parse()

	log.Println("[INFO] Parsing configuration file")

	configFile, err := ioutil.ReadFile(configFileFlag)
	if err != nil {
		log.Fatalf("[FATAL] Unable to read configuration file %s: %v\n", configFileFlag, err)
	}

	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalf("[FATAL] Unable to parse configuration file: %v\n", err)
	}

	log.Printf("[INFO] Listening on http://%s:%d/update\n", bindFlag, portFlag)
	http.HandleFunc("/update", httpUpdate)
	http.ListenAndServe(fmt.Sprintf("%s:%d", bindFlag, portFlag), nil)
}
