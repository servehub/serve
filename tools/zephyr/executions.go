package zephyr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/cenk/backoff"
	"golang.org/x/oauth2"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"strings"
)

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

func CreateFormFile(w *multipart.Writer, fieldName, fileName string, contentType string) (io.Writer, error) {
	return w.CreatePart(textproto.MIMEHeader{
		"Content-Disposition": {fmt.Sprintf(`form-data; name="%s"; filename="%s"`, escapeQuotes(fieldName), escapeQuotes(fileName))},
		"Content-Type":        {contentType},
	})
}

type TypedReader struct {
	contentType string
	filename    string
	reader      io.Reader
}

type ErrorResponse struct {
	ErrorCode int    `json:"errorCode"`
	Message   string `json:"message"`
}

type TestCycle struct {
	Name               string `json:"name"`
	Description        string `json:"description"`
	JiraProjectVersion int    `json:"jiraProjectVersion"`
	FolderId           int    `json:"folderId"`
	CustomFields       string `json:"customFields"`
}

const URL = "https://api.zephyrscale.smartbear.com/v2/automations/executions/junit?"

func UploadJunitReport(token string, projectKey string, reportFilePath string, testCycleInfo *TestCycle) error {
	reportFile, err := os.Open(reportFilePath)
	if err != nil {
		return err
	}
	defer reportFile.Close()

	values := map[string]TypedReader{
		"file": {"application/xml", reportFile.Name(), reportFile},
	}
	if testCycleInfo != nil {
		testCycleBytes, _ := json.Marshal(testCycleInfo)
		values["testCycle"] = TypedReader{"application/json", "cycle.json", strings.NewReader(string(testCycleBytes))}
	}

	buf := new(bytes.Buffer)
	writer := multipart.NewWriter(buf)
	defer writer.Close()

	for key, data := range values {
		part, err := CreateFormFile(writer, key, data.filename, data.contentType)
		if err != nil {
			return err
		}
		if _, err = io.Copy(part, data.reader); err != nil {
			return err
		}
	}
	if err = writer.Close(); err != nil {
		return err
	}

	params := url.Values{
		"projectKey": {projectKey},
		//"autoCreateTestCases": {"true"},
	}
	remoteURL := URL + params.Encode()

	client := oauth2.NewClient(
		context.Background(),
		oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))
	return backoff.Retry(func() error {
		output := bytes.NewReader(buf.Bytes())
		response, err := client.Post(remoteURL, writer.FormDataContentType(), output)
		if err != nil {
			return err
		}
		if shouldBeRetried(response.StatusCode) {
			return fmt.Errorf("invalid status: %s", response.Status)
		}
		if response.StatusCode != http.StatusOK {
			defer response.Body.Close()
			body, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Printf("failed to upload report to zephyr: Invalid responce [%d] %s", response.StatusCode, response.Status)
				return nil
			}
			msg := ErrorResponse{}
			if err := json.Unmarshal(body, &msg); err != nil {
				log.Printf("failed to upload report to zephyr: Invalid responce [%d] %s:\n%s", response.StatusCode, response.Status, string(body))
				return nil
			}
			log.Printf("failed to upload report to zephyr: [%d] %s", msg.ErrorCode, msg.Message)
		}
		return nil
	}, backoff.NewExponentialBackOff())
}

func shouldBeRetried(statusCode int) bool {
	switch statusCode {
	case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	}
	return false
}
