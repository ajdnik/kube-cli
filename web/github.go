package web

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

const latestReleaseURL = "https://api.github.com/repos/ajdnik/kube-cli/releases/latest"

// LatestRelease represents the GitHub latest release asset
// download info.
type LatestRelease struct {
	Version string
	ShaURL  string
	TarURL  string
}

// GetLatestRelease returns the LatestRelease information.
func GetLatestRelease(shaFile, tarFile string) (LatestRelease, error) {
	var release LatestRelease
	var trans = &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}
	var client = &http.Client{
		Timeout:   time.Second * 10,
		Transport: trans,
	}
	res, err := client.Get(latestReleaseURL)
	if err != nil {
		return release, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return release, err
	}
	var raw map[string]interface{}
	json.Unmarshal(body, &raw)
	if _, ok := raw["tag_name"]; !ok {
		return release, errors.New("problem parsing json, missing tag_name property")
	}
	latestVersion := raw["tag_name"].(string)
	if _, ok := raw["assets"]; !ok {
		return release, errors.New("problem parsing json, missing assets property")
	}
	assets := raw["assets"].([]interface{})
	if len(assets) == 0 {
		return release, errors.New("problem parsing json, assets array is empty")
	}
	release = LatestRelease{
		Version: latestVersion,
	}
	for _, asset := range assets {
		a := asset.(map[string]interface{})
		if _, ok := a["name"]; !ok {
			continue
		}
		if _, ok := a["browser_download_url"]; !ok {
			continue
		}
		name := a["name"].(string)
		url := a["browser_download_url"].(string)
		if name == shaFile {
			release.ShaURL = url
		}
		if name == tarFile {
			release.TarURL = url
		}
	}
	if len(release.ShaURL) == 0 || len(release.TarURL) == 0 {
		return release, errors.New("problem parsing json, assets not found")
	}
	return release, nil
}
