# [gogas](https://github.com/tnum5/gogas)

**gogas** is tool (written in GO) for exporting and importing Google App Script files to/from a local directory.

The tool is useful to those of use that prefer to use our own Javascript Editors and tools locally, rather than edit in Googles cloud based editor. (Also, you can put the code in a source control system)

## Installing

Since this is a GO program, this instructions assume that you have GO installed. If not, you can get GO installation instructions [here](https://golang.org/doc/install).

After cloning, point your $GOPATH to the ```gogas```  directory, then run ```go build gogas``` or ```go install gogas```.

## Usage

Usage:

```code
gogas [options] <projectname>

  -cmd string
    	command: one of 'get' or 'put', 'get' for exporting (downloading) or 'put' for importing (uploading)

  -dir string
    	directory to download to; absolute path or relative to the current working directory

  -s	do not expand into local files. Simply download to/from <projectname>.json


  ```

<projectname> is the name on the library or script project on Google drive.

## First time use

The first time ```gogas``` is used, a oauth2 authentication needs to take place. During that process, a set of tokens are created and stored locally. Once this is done, the tool can be used repeatedly without this step.

###Step 1: Enable the Drive API

1. Use this wizard to create or select a project in the   Google Developers Console and automatically enable the  API. Click the Go to credentials button to continue.
2. At the top of the page, select the OAuth consent screen tab. Select an Email address, enter a Product name if not already set, and click the Save button.
3. Back on the Credentials tab, click the Add credentials button and select OAuth 2.0 client ID.
4. Select the application type Other and click the Create button.
5. Click OK to dismiss the resulting dialog.
6. Click the (Download JSON) button to the right of the client ID. Move this file to the directory ```$HOME/.credentials/client_secret/``` and rename it **gogas_client_secret.json**.
