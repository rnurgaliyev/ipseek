package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"path"
)

// OpenStack API JSON structures

type authRequest struct {
	Auth struct {
		Identity struct {
			Methods  []string `json:"methods"`
			Password struct {
				User struct {
					Name   string `json:"name"`
					Domain struct {
						Name string `json:"name"`
					} `json:"domain"`
					Password string `json:"password"`
				} `json:"user"`
			} `json:"password"`
		} `json:"identity"`
	} `json:"auth"`
}

type siteConnectionSettings struct {
	IpsecSiteConnection struct {
		PeerAddress string `json:"peer_address"`
		PeerID      string `json:"peer_id"`
	} `json:"ipsec_site_connection"`
}

func updateOpenstackIpsecSiteConnection(c map[string]string, address string) bool {
	var client *http.Client = &http.Client{}

	r := authRequest{}

	r.Auth.Identity.Methods = []string{"password"}
	r.Auth.Identity.Password.User.Name = c["user"]
	r.Auth.Identity.Password.User.Domain.Name = c["domain"]
	r.Auth.Identity.Password.User.Password = c["password"]

	// Prepare authentication request

	authRequestBody, err := json.Marshal(r)
	if err != nil {
		log.Printf("[ERROR] openstack_ipsec_site_connection: Unexpected error %v\n", err)
		return false
	}

	authRequest, err := http.NewRequest(http.MethodPost, c["authURL"], bytes.NewBuffer(authRequestBody))
	if err != nil {
		log.Printf("[ERROR] openstack_ipsec_site_connection: Unexpected error %v\n", err)
		return false
	}

	authRequest.Header.Set("Content-Type", "application/json; charset=utf-8")

	// Make authentication request

	authRespone, err := client.Do(authRequest)
	if err != nil {
		log.Printf("[ERROR] openstack_ipsec_site_connection: Unable make a request to Identity API (%v)\n", err)
		return false
	}

	if authRespone.StatusCode != 201 {
		log.Printf("[ERROR] openstack_ipsec_site_connection: Authentication failed\n")
		return false
	}

	token := authRespone.Header["X-Subject-Token"][0]

	// Process ipsec site connection API URL

	u, err := url.Parse(c["url"])
	if err != nil {
		log.Printf("[ERROR] openstack_ipsec_site_connection: Wrong URL (%v)\n", err)
		return false
	}

	u.Path = path.Join(u.Path, c["id"])

	// Get current site connection settings

	getRequest, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		log.Printf("[ERROR] openstack_ipsec_site_connection: Unable retreive current configuration (%v)\n", err)
		return false
	}
	getRequest.Header.Set("X-Auth-Token", token)

	getResponse, err := client.Do(getRequest)

	b, err := ioutil.ReadAll(getResponse.Body)
	if err != nil {
		log.Printf("[ERROR] openstack_ipsec_site_connection: Unexpected error %v\n", err)
		return false
	}

	currentConfig := siteConnectionSettings{}
	json.Unmarshal(b, &currentConfig)

	// Stop if no update is needed

	if currentConfig.IpsecSiteConnection.PeerAddress == address && currentConfig.IpsecSiteConnection.PeerID == address {
		log.Printf("[DEBUG] openstack_ipsec_site_connection: No update is needed\n")
		return true
	}

	// Prepare update request

	s := siteConnectionSettings{}
	s.IpsecSiteConnection.PeerAddress = address
	s.IpsecSiteConnection.PeerID = address
	siteConnectionSettingsBody, err := json.Marshal(s)
	if err != nil {
		log.Printf("[ERROR] openstack_ipsec_site_connection: Unexpected error %v\n", err)
		return false
	}

	updateRequst, err := http.NewRequest(http.MethodPut, u.String(), bytes.NewBuffer(siteConnectionSettingsBody))

	updateRequst.Header.Set("Content-Type", "application/json; charset=utf-8")
	updateRequst.Header.Set("X-Auth-Token", token)

	// Perform update request

	updateRespone, err := client.Do(updateRequst)
	if err != nil {
		log.Printf("[ERROR] openstack_ipsec_site_connection: Failed to commit update request (%v)\n", err)
		return false
	}

	if updateRespone.StatusCode != 200 {
		log.Printf("[ERROR] openstack_ipsec_site_connection: Site connection update failed (Status Code: %d)\n", updateRespone.StatusCode)
		return false
	}

	return true
}
