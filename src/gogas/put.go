package gogas

import (
	"bytes"
	"encoding/json"
	"fmt"
	"google.golang.org/api/drive/v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	//"net/http/httputil"
	"os"
	"path"
	"strings"
)

// ImportProject moves files up to a specific project on Drive
func ImportProject(srv *drive.Service, client *http.Client, projName string, fileID string, short bool) error {
	var m io.Reader
	var err error
	// if project already exists in json (short = true) read it,
	// else collect data from files in directory
	if short {
		filename := projName + ".json"
		m, err = os.Open(filename)
		if err != nil {
			fmt.Printf("An error occurred opening the file: %v\n", err)
			return err
		}
	} else {
		data, err := DownloadFile(srv, *client, fileID)
		buf, err := readAndBuild(data)
		if err != nil {
			return err
		}
		m = bytes.NewReader(buf)
	}
	// Now send the file to be uploaded
	_, err = UploadFiles(client, m, fileID)
	if err != nil {
		fmt.Printf("An error occurred updating the script:\n %v\n", err.Error())
		return err
	}
	return nil
}

// UploadFiles performs the http call that uploads GAS files for the project
func UploadFiles(client *http.Client, media io.Reader, fileID string) (*http.Response, error) {
	url := "https://www.googleapis.com/upload/drive/v2/files/" + fileID
	req, _ := http.NewRequest("PUT", url, media)
	req.Header.Set("User-Agent", "google-api-go-client/0.5")
	cacheFile, err := tokenCacheFile()
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		log.Fatalf("Unable to get credential file. %v", err)
	}
	req.Header.Set("Authorization", tok.TokenType+" "+" "+tok.AccessToken)
	req.Header.Set("Content-Type", "application/vnd.google-apps.script+json")

	//dump, _ := httputil.DumpRequest(req, true)
	//fmt.Println(string(dump))
	return client.Do(req)
}

// readAndBuild reads from Drive and the local directory files and builds and
// object that can be uploaded to drive with the local drive files overwritting
// the cloud files
func readAndBuild(data []byte /*, directory string*/) ([]byte, error) {
	directory, err := os.Getwd()
	var jmap map[string]interface{}
	if err = json.Unmarshal(data, &jmap); err != nil {
		fmt.Printf("Error decoding exported json: %v\n", err)
		return nil, err
	}
	pfiles := jmap["files"].([]interface{})
	lfiles, err := ioutil.ReadDir(directory)
	if err != nil {
		fmt.Printf("Error reading cwd: %v\n", err)
		return nil, err
	}

	// Create a array with all the files that end with '.gs' and '.html'
	var gasFiles []string
	validExt := []string{".gs", ".html"}
	for _, file := range lfiles {
		if !file.IsDir() {
			ext := path.Ext(file.Name())
			if Contains(validExt, ext) {
				gasFiles = append(gasFiles, file.Name())
			}
		}
	}

	// match the names of local directory files and cloud project files (on Google Drive)
	// create a modified entry for each match
	var buf []byte
	var fileArray []map[string]string
	for _, e := range pfiles {
		fileMap := make(map[string]string)
		each := e.(map[string]interface{})
		fname := each["name"].(string)
		ftype := each["type"].(string)
		fid := each["id"].(string)
		fullname := ""
		if ftype == "server_js" {
			fullname = fname + ".gs"
		} else if ftype == "html" {
			fullname = fname + ".html"
		} else {
			log.Fatalf("Unexpected GAS file type = %s\n", ftype)
		}
		if Contains(gasFiles, fullname) {
			buf, err = ioutil.ReadFile(fullname)
			if err != nil {
				fmt.Printf("Reading file '%s', error = %v\n", fullname, err)
				return nil, err
			}
			fileMap["source"] = string(buf)
			fileMap["id"] = fid
			fileMap["type"] = ftype
			fileMap["name"] = fname
			fileArray = append(fileArray, fileMap)
			i := Index(gasFiles, fullname)
			gasFiles = append(gasFiles[:i], gasFiles[i+1:]...)
		}
	}

	// Now create entries for files that are in the directory but not in the cloud project
	for _, each := range gasFiles {
		fileMap := make(map[string]string)
		split := strings.Split(each, ".")
		fileMap["name"] = split[0]
		if split[1] == "gs" {
			fileMap["type"] = "server_js"
		} else {
			fileMap["type"] = "html"
		}
		buf, err = ioutil.ReadFile(each)
		if err != nil {
			fmt.Printf("Reading file '%s', error = %v\n", each, err)
			return nil, err
		}
		fileMap["source"] = string(buf)
		fileArray = append(fileArray, fileMap)
	}

	// encapsulate and create a json string
	var project = make(map[string][]map[string]string)
	project["files"] = fileArray
	retval, err := json.Marshal(project)
	if err != nil {
		fmt.Printf("Encoding json, error = %v\n", err)
		return nil, err
	}
	return retval, nil
}

// Contains returns true or false based on whether an element in in an Array
func Contains(list []string, elem string) bool {
	for _, t := range list {
		if t == elem {
			return true
		}
	}
	return false
}

// Index returns the index of an element in an Array or Slice
func Index(list []string, elem string) int {
	for i, t := range list {
		if t == elem {
			return i
		}
	}
	return -1
}
