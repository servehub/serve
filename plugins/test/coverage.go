package test

import (
	"encoding/xml"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/serve/tools/github"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("test.coverage", CoverageUpload{})
}

type CoverageUpload struct{}

func (p CoverageUpload) Run(data manifest.Manifest) error {
	// get repo, commit information
	meta := Meta{
		Repo:     data.GetString("repo"),
		Branch:   data.GetString("branch"),
		Ref:      data.GetString("ref"),
		Version:  data.GetString("version"),
		TestType: data.GetString("test-type"),
	}

	execFilePath := data.GetString("exec-file")
	if _, err := os.Stat(execFilePath); errors.Is(err, os.ErrNotExist) {
		log.Printf("coverage file doesn't exist - skipping: %s", execFilePath)
		return nil
	}

	coveragePercent, err := generateReportsAndGetCoveragePercent([]string{execFilePath}, data)
	if err != nil {
		return err
	}

	// main branch name
	mainBranchName := data.GetStringOr("main-branch", "master")

	connectionUrl := os.Getenv(data.GetString("database-connection-env"))
	db, err := gorm.Open(postgres.Open(connectionUrl), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Migrate the DB schema
	if err := db.AutoMigrate(&CoverageReport{}); err != nil {
		return fmt.Errorf("failed to apply database migration: %w", err)
	}

	if meta.Branch == mainBranchName {
		coverageData, err := os.ReadFile(execFilePath)
		if err != nil {
			return fmt.Errorf("failed to read coverage exec file: %w", err)
		}

		reportRecord := CoverageReport{
			Meta:            meta,
			CoveragePercent: coveragePercent,
			CoverageFile:    coverageData,
		}
		if err := db.Create(&reportRecord).Error; err != nil {
			return fmt.Errorf("failed to upload coverage exec file: %w", err)
		}

		if err := db.Unscoped().Where(&CoverageReport{Meta: Meta{
			Repo:     meta.Repo,
			Branch:   mainBranchName,
			TestType: meta.TestType,
		}}).Where("id != ?", reportRecord.ID).Delete(&CoverageReport{}).Error; err != nil {
			log.Printf("failed to removed outdated records: %s", err)
		}

		return nil
	}

	// check mode
	accessToken := os.Getenv("GITHUB_TOKEN")
	if accessToken == "" {
		return errors.New("`GITHUB_TOKEN` is required")
	}

	targetUrl := data.GetString("check.target-url")
	statusContext := data.GetStringOr("check.context", "coverage")

	// allow a small tolerance for decrease in coverage
	coverageTolerance := data.GetFloat("check.tolerance")

	// get latest coverage report from database
	var latestCoverage CoverageReport

	if err := db.Where(&CoverageReport{Meta: Meta{
		Repo:     meta.Repo,
		Branch:   mainBranchName,
		TestType: meta.TestType,
	}}).Last(&latestCoverage).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return github.SendStatus(accessToken, meta.Repo, meta.Ref, "success",
				"No previous coverage exists - check skipped ",
				statusContext, targetUrl)
		}
		return err
	}

	diff := latestCoverage.CoveragePercent - coveragePercent

	if diff < 0 {
		return github.SendStatus(accessToken, meta.Repo, meta.Ref, "success",
			fmt.Sprintf("Thank you for increasing the test coverage by %.2f%%", -diff),
			statusContext, targetUrl)
	} else if diff < coverageTolerance {
		return github.SendStatus(accessToken, meta.Repo, meta.Ref, "success",
			fmt.Sprintf("Coverage changed by %.2f%%", diff),
			statusContext, targetUrl)
	} else {
		return github.SendStatus(accessToken, meta.Repo, meta.Ref, "failure",
			fmt.Sprintf("Please increase test coverage at least by %.2f%%", diff),
			statusContext, targetUrl)
	}
}

type Meta struct {
	Repo     string `json:"repo"   gorm:"index"`
	Branch   string `json:"branch" gorm:"index"`
	Ref      string `json:"ref"`
	Version  string `json:"version"`
	TestType string `json:"test_type"`
}

type CoverageReport struct {
	gorm.Model
	Meta
	// coverage percentage in the range [0.0, 100.0]
	CoveragePercent float64
	CoverageFile    []byte
}

type CounterXML struct {
	XMLName xml.Name `xml:"counter"`
	Type    string   `xml:"type,attr"`
	Missed  int      `xml:"missed,attr"`
	Covered int      `xml:"covered,attr"`
}

type CoverageReportXML struct {
	XMLName xml.Name     `xml:"report"`
	Counter []CounterXML `xml:"counter"`
}

func generateReportsAndGetCoveragePercent(execCoverageFiles []string, data manifest.Manifest) (float64, error) {
	builder := strings.Builder{}
	builder.WriteString(fmt.Sprintf("java -jar %s report %s",
		data.GetStringOr("generate.jacococli-jar", "jacococli.jar"),
		strings.Join(execCoverageFiles, " "),
	))

	if data.Has("generate.sourcefiles") {
		for _, sourceFiles := range data.GetArray("generate.sourcefiles") {
			builder.WriteString(fmt.Sprintf(" --sourcefiles %s", sourceFiles))
		}
	}

	for _, classFiles := range data.GetArray("generate.classfiles") {
		builder.WriteString(fmt.Sprintf(" --classfiles %s", classFiles))
	}

	if data.Has("generate.html-output-dir") {
		htmlOutputDir := data.GetString("generate.html-output-dir")
		if htmlOutputDir != "" {
			builder.WriteString(fmt.Sprintf(" --html %s", htmlOutputDir))
		}
	}

	dirName, err := os.MkdirTemp("", "xml-report")
	if err != nil {
		return 0, errors.New("failed to create directory for temporary XML report file")
	}

	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			log.Fatal(err)
		}
	}(dirName)

	xmlReportFile := filepath.Join(dirName, "coverage.xml")
	builder.WriteString(fmt.Sprintf(" --xml %s", xmlReportFile))

	if err = utils.RunCmd(builder.String()); err != nil {
		return 0, err
	}

	return getCoveragePercent(xmlReportFile)
}

func getCoveragePercent(xmlReportFile string) (float64, error) {
	coverageXml, err := os.ReadFile(xmlReportFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read xml coverage report file: %w", err)
	}

	// (partially) unmarshall report from xml
	var coverageReportXml CoverageReportXML
	if err = xml.Unmarshal(coverageXml, &coverageReportXml); err != nil {
		return 0, fmt.Errorf("failed to parse xml coverage report file: %w", err)
	}

	// search for the XML counter matching the desired coverage metric
	var coverageCounter CounterXML
	for _, counter := range coverageReportXml.Counter {
		if counter.Type == "INSTRUCTION" {
			coverageCounter = counter
			break
		}
	}

	total := coverageCounter.Covered + coverageCounter.Missed
	if total == 0 {
		return 0, err
	}
	return 100 * float64(coverageCounter.Covered) / float64(total), nil
}
