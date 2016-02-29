package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path"
)

var (
	ZDUSER     string
	ZDPASS     string
	ZDFILE     string
	ZDKBID     string
	ZDROOT     string
	ZDENDPOINT = "/api/v2/help_center/articles"

	FILEPATH = flag.String("f", "", "Provide path of file for upload")
	KBID     = flag.String("k", "", "KBID of article you wish to upload file to")

	/*
		https://developer.zendesk.com/rest_api/docs/help_center/article_attachments#list-article-attachments
		GET
		https://domain/api/v2/help_center/articles/217546277/attachments.json
		{"article_attachments":[]}
	*/
	LISTATTACHMENTS string

	/*
		https://developer.zendesk.com/rest_api/docs/help_center/article_attachments#create-article-attachment
		POST
		curl https://domain/api/v2/help_center/articles/217546277/attachments.json -F "inline=false" -F "file=@gpmt.gz" -v -u user@pivotal.io:password -X POST
	*/
	CREATEATTACHMENTS string

	/*
		No global for this one as we need to build it on the fly and its in a for loop so just handle it in the loop
		https://developer.zendesk.com/rest_api/docs/help_center/article_attachments#delete-article-attachment
		DELETE
		https://domain/api/v2/help_center/articles/attachments/12345.json

	*/
)

type Attachments struct {
	Attachement []struct {
		ID          int    `json:"id"`
		Url         string `json:"url"`
		ArticleID   int    `json:"article_id"`
		FileName    string `json:"file_name"`
		ContentURL  string `json:"content_url"`
		ContentType string `json:"content_type"`
		Size        int    `json:"content_type"`
		Inline      bool   `json:"inline"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	} `json:"article_attachments"`
}

/*
   sends a HTTP GET request to the given URl and returns []byte response
*/
func httpClient(URL, REQTYPE, user, pass string, file *os.File) ([]byte, error) {
	client := &http.Client{}
	var noData []byte

	var req *http.Request
	var reqErr error
	if file == nil {
		req, reqErr = http.NewRequest(REQTYPE, URL, nil)
		if reqErr != nil {
			return nil, reqErr
		}
		req.Header.Add("Content-Type", "application/json")
	} else {

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", path.Base(file.Name()))
		if err != nil {
			return noData, err
		}
		_, err = io.Copy(part, file)
		if err != nil {
			return noData, err
		}
		err = writer.Close()
		if err != nil {
			return noData, err
		}

		req, reqErr = http.NewRequest(REQTYPE, URL, body)
		if reqErr != nil {
			return nil, reqErr
		}
		req.Header.Add("Content-Type", writer.FormDataContentType())
	}

	req.SetBasicAuth(user, pass)

	resp, clientErr := client.Do(req)
	if clientErr != nil {
		return nil, clientErr
	}
	defer resp.Body.Close()
	data, respErr := ioutil.ReadAll(resp.Body)
	if respErr != nil {
		return nil, respErr
	}
	return data, nil

}

func main() {
	flag.Parse()

	ZDUSER = os.Getenv("ZDUSER")
	ZDPASS = os.Getenv("ZDPASS")
	ZDKBID = *KBID
	ZDROOT = os.Getenv("ZDROOT")
	ZDFILE = path.Base(*FILEPATH)
	files := new(Attachments)

	if ZDKBID == "" || ZDFILE == "" {
		flag.PrintDefaults()
		return
	}

	LISTATTACHMENTS = ZDROOT + ZDENDPOINT + "/" + ZDKBID + "/attachments.json"
	CREATEATTACHMENTS = ZDROOT + ZDENDPOINT + "/" + ZDKBID + "/attachments.json"

	/*
		1. get attachments list
		2. if ZDFILE exists then delete
		3. upload new ZDFILE attachement
	*/

	/* ######################## List Existing ################################# */
	raw, err := httpClient(LISTATTACHMENTS, "GET", ZDUSER, ZDPASS, nil)
	if err != nil {
		fmt.Printf("Failed to get list of attachements: %s\n", err)
		return
	}
	fmt.Printf("%s\n", raw)

	err = json.Unmarshal(raw, &files)
	if err != nil {
		fmt.Printf("Failed to unmarshal zendesk list attachments reponse: %s\n", err)
		return
	}

	/* ######################## DELETE Existing ################################# */
	for i := range files.Attachement {
		if files.Attachement[i].FileName == ZDFILE {
			fileid := fmt.Sprintf("%d", files.Attachement[i].ID)
			url := ZDROOT + ZDENDPOINT + "/attachments/" + fileid + ".json"
			_, err = httpClient(url, "DELETE", ZDUSER, ZDPASS, nil)
			if err != nil {
				fmt.Printf("Failed to delete kb %s: %s\n", url, err)
				return
			}
		}
	}

	/* ######################## Upload file ################################# */
	filebuff, fderr := os.Open(*FILEPATH)
	if fderr != nil {
		fmt.Print("Failed to open file %s: %s\n", ZDFILE, fderr)
		return
	}
	defer filebuff.Close()

	raw, err = httpClient(CREATEATTACHMENTS, "POST", ZDUSER, ZDPASS, filebuff)
	fmt.Printf("%s: %s\n", CREATEATTACHMENTS, raw)
}
