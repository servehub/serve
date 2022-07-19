package test

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/serve/tools/zephyr"
)

func init() {
	manifest.PluginRegestry.Add("test.upload-zephyr", ExecutionUpload{})
}

type ExecutionUpload struct{}

func (p ExecutionUpload) Run(data manifest.Manifest) error {
	mainBranchName := data.GetStringOr("main-branch", "master")
	branch := data.GetString("branch")
	if branch != mainBranchName {
		log.Printf("uploading only for a `%s` branch - skipping: %s", mainBranchName, branch)
		return nil
	}

	accessToken := os.Getenv("ZEPHYR_SCALE_TOKEN")
	if accessToken == "" {
		return errors.New("`ZEPHYR_SCALE_TOKEN` is required")
	}

	reportFilePath := data.GetString("report-file")
	files, err := filepath.Glob(reportFilePath)
	if err != nil {
		return err
	}
	if files == nil {
		log.Printf("report file(s) doesn't exist - skipping: %s", reportFilePath)
		return nil
	}

	dirName, err := os.MkdirTemp("", "xml-reports")
	if err != nil {
		return errors.New("failed to create directory for temporary zip with XML report files")
	}

	defer func(path string) {
		if err := os.RemoveAll(path); err != nil {
			log.Fatal(err)
		}
	}(dirName)

	zipReportFile := filepath.Join(dirName, "junit-tests.zip")
	if err := mergeIntoZip(zipReportFile, files); err != nil {
		return err
	}

	cycle := zephyr.TestCycle{
		Name: fmt.Sprintf(`%s %s [%s]`,
			data.GetString("app-name"),
			data.GetString("version"),
			data.GetString("test-type"),
		),
	}

	return zephyr.UploadJunitReport(
		accessToken,
		data.GetString("project-key"),
		zipReportFile,
		data.GetBool("auto-create-test-cases"),
		&cycle,
	)
}

func mergeIntoZip(zipReportFile string, files []string) error {
	flags := os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	file, err := os.OpenFile(zipReportFile, flags, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	func() {
		zipw := zip.NewWriter(file)
		defer zipw.Close()

		for _, filename := range files {
			if err := appendFiles(filename, zipw); err != nil {
				log.Fatalf("Failed to add file %s to zip: %s", filename, err)
			}
		}
	}()

	return nil
}

func appendFiles(filename string, zipw *zip.Writer) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open %s: %s", filename, err)
	}
	defer file.Close()

	wr, err := zipw.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create entry for %s in zip file: %s", filename, err)
	}

	if _, err := io.Copy(wr, file); err != nil {
		return fmt.Errorf("failed to write %s to zip: %s", filename, err)
	}

	return nil
}
