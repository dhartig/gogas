package gogas

import (
	"flag"
	"fmt"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v2"
	"io/ioutil"
	"log"
)

// GetFileID get the fileID of a given file name.
// Note 1: Will only return the first fileId if the file has the right Mime type
func GetFileID(srv *drive.Service, fileName string) (string, error) {
	q := srv.Files.List()
	q = q.Q("title = '" + fileName + "' and mimeType = 'application/vnd.google-apps.script'")
	r, err := q.Do()
	if err != nil {
		return "", err
	}
	file := r.Items[0]
	return file.Id, nil
}
func main() {

	var cmdArg string
	var shortFlag bool
	var dirArg string
	flag.Usage = func() {
		fmt.Printf("Usage: gogas [options] <projectname>\n")
		flag.PrintDefaults()
	}

	flag.BoolVar(&shortFlag, "s", false, "don't expand into local files")
	flag.StringVar(&cmdArg, "cmd", "", "command: one of 'get' or 'put'")
	flag.StringVar(&dirArg, "dir", "", "directory to download to")
	flag.Parse()
	projArg := flag.Args()
	if cmdArg != "get" && cmdArg != "put" {
		flag.Usage()
		return
	}
	if len(projArg) == 0 {
		flag.Usage()
		return
	}

	fmt.Printf("Project Name: %s\n", projArg[0])

	ctx := context.Background()
	spath, err := secretsFile()
	b, err := ioutil.ReadFile(spath)
	if err != nil {
		log.Fatalf("gogas: unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, drive.DriveScope, drive.DriveScriptsScope)
	if err != nil {
		log.Fatalf("gogas: unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config)

	srv, err := drive.New(client)
	if err != nil {
		log.Fatalf("gogas: unable to retrieve drive Client %v", err)
	}

	fileID, err := GetFileID(srv, projArg[0])
	if err != nil {
		fmt.Printf("gogas: Error fetching file id = %v\n", err)
		return
	}

	fmt.Printf("gogas: FileID = %s\n", fileID)
	if cmdArg == "get" {
		err = ExportProject(srv, client, projArg[0], fileID, shortFlag)
		if err != nil {
			log.Fatalf("gogas: an error occurred saving files: %v\n", err)
		}
	} else {
		err = ImportProject(srv, client, projArg[0], fileID, shortFlag)
		if err != nil {
			log.Fatalf("gogas: an error occurred uploading the project: %v\n", err)
		}
	}
}
