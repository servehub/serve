package test

import (
	"errors"
	"log"
	"os"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/serve/tools/zephyr"
)

func init() {
	manifest.PluginRegestry.Add("test.upload-zephyr", ExecutionUpload{})
}

type ExecutionUpload struct{}

func (p ExecutionUpload) Run(data manifest.Manifest) error {
	accessToken := os.Getenv("ZEPHYR_SCALE_TOKEN")
	if accessToken == "" {
		return errors.New("`ZEPHYR_SCALE_TOKEN` is required")
	}

	reportFilePath := data.GetString("report-file")
	if _, err := os.Stat(reportFilePath); errors.Is(err, os.ErrNotExist) {
		log.Printf("report file doesn't exist - skipping: %s", reportFilePath)
		return nil
	}

	return zephyr.UploadJunitReport(
		accessToken,
		data.GetString("project-key"),
		reportFilePath,
		nil,
	)
}
