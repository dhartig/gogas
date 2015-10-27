package gogas

import (
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/api/drive/v2"
	"io/ioutil"
	"net/http"
)

// ExportProject retrieves the GAS project files and writes them to local directory
func ExportProject(srv *drive.Service, client *http.Client, projName string, fileID string, short bool) error {
	data, err := DownloadFile(srv, *client, fileID)
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return err
	}
	if short {
		filename := projName + ".json"
		ioutil.WriteFile(filename, []byte(data), 0644)
	} else {
		if err = parseAndSave(data); err != nil {
			fmt.Printf("An error occured saving json to files = %v\n", err)
			return err
		}
	}
	return nil
}

// DownloadFile downloads the content of a given file object
func DownloadFile(d *drive.Service, c http.Client, fileID string) ([]byte, error) {
	f, err := d.Files.Get(fileID).Do()
	if err != nil {
		fmt.Printf("An error occurred reading the File object: %v\n", err)
		return nil, err
	}
	// t parameter should use an oauth.Transport
	downloadURL := f.ExportLinks["application/vnd.google-apps.script+json"]
	if downloadURL == "" {
		// If there is no downloadUrl, there is no body
		fmt.Printf("An error occurred: File is not downloadable\n")
		err := errors.New("File is not downloadable")
		return nil, err
	}
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return nil, err
	}
	resp, err := c.Do(req)
	// Make sure we close the Body later
	defer resp.Body.Close()
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("An error occurred: %v\n", err)
		return nil, err
	}
	return body, nil
}

// PrintFile fetches and displays the given file.
func PrintFile(f *drive.File) error {

	fmt.Printf("Title: %v\n", f.Title)
	fmt.Printf("Description: %v\n", f.Description)
	fmt.Printf("MIME type: %v\n", f.MimeType)
	fmt.Printf("DownloadUrl: %s\n", f.DownloadUrl)
	fmt.Printf("Export Links: \n")
	for key, value := range f.ExportLinks {
		fmt.Printf("Key, Value: %s, %s\n ", key, value)
	}
	fmt.Printf("\n\n")
	return nil
}

// ParseAndSave parses the exported json and breaks it up into files
func parseAndSave(data []byte) error {
	var jmap map[string]interface{}
	if err := json.Unmarshal(data, &jmap); err != nil {
		fmt.Printf("Error decoding exported json: %v\n", err)
		return err
	}
	files := jmap["files"].([]interface{})

	for _, each := range files {
		e := each.(map[string]interface{})
		var filename string
		if e["type"] == "server_js" {
			filename = e["name"].(string) + ".gs"
		} else if e["type"] == "html" {
			filename = e["name"].(string) + ".html"
		} else {
			filename = e["name"].(string)
		}
		if err := ioutil.WriteFile(filename, []byte(e["source"].(string)), 0644); err != nil {
			fmt.Printf("Error writing json to file: %v\n", err)
			return err
		}
	}
	return nil
}
