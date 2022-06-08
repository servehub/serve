package test

import (
	"encoding/xml"
	"errors"
	"fmt"
	"os"

	"github.com/servehub/utils"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/servehub/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("test.coverage", TestCoverageUpload{})
}

type TestCoverageUpload struct{}

func (p TestCoverageUpload) Run(data manifest.Manifest) error {

	// get repo, commit information
	meta := Meta{
		Repo:     data.GetString("repo"),
		Branch:   data.GetString("branch"),
		Ref:      data.GetString("ref"),
		Version:  data.GetString("version"),
		TestType: data.GetString("test-type"),
	}

	// configurable main branch name
	mainBranch := data.GetStringOr("main-branch", "master")

	// allow a small tolerance for decrease in coverage
	coverage_tolerance := data.GetFloat("coverage-tolerance")
	// the jacoco metric to
	coverage_metric := data.GetStringOr("coverage-metric", "INSTRUCTION")
	println("coverage_tolerance =", coverage_tolerance)

	if generateCmd := data.GetString("generate"); generateCmd != "" {
		utils.RunCmd(generateCmd)
	}

	coverageFile, err := os.ReadFile(data.GetString("coverage-file"))
	if err != nil {
		return errors.New("failed to read coverage file")
	}
	coverageXml, err := os.ReadFile(data.GetString("coverage-xml"))
	if err != nil {
		return errors.New("failed to read coverage XML file")
	}

	// (partially) unmarshall report from xml
	var coverageReportXml CoverageReportXML
	xml.Unmarshal(coverageXml, &coverageReportXml)

	// search for the XML counter matching the desired coverage metric
	var coverageCounter CounterXML
	for _, counter := range coverageReportXml.Counter {
		if counter.Type == coverage_metric {
			coverageCounter = counter
			break
		}
	}

	// calculate coverage for current branch
	coveragePercent := getCoveragePercent(coverageCounter)

	// --- DB Access ---
	// https://github.com/jackc/pgx#example-usage
	// DSN should be in format:
	// "postgres://username:password@localhost:5432/database_name"
	//  -- or --
	// "host=localhost user=postgres password=postgres port=5432"
	dsn := os.Getenv(data.GetString("database-connection-env"))

	// https://gorm.io/docs/connecting_to_the_database.html#PostgreSQL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return errors.New("failed to connect to database")
	}

	// Migrate the DB schema
	db.AutoMigrate(&CoverageReport{})

	// only the main branch should publish coverage
	if meta.Branch == mainBranch {
		db.Create(&CoverageReport{
			Meta:            meta,
			CoveragePercent: coveragePercent,
			CoverageFile:    coverageFile,
		})
	} else {
		// get latest coverage report from database
		var latestCoverage CoverageReport
		db.Where(&CoverageReport{Meta: Meta{
			Repo:     meta.Repo,
			Branch:   mainBranch,
			TestType: meta.TestType,
		}}).Last(&latestCoverage)

		fmt.Printf("latest coverage: %s => %f", latestCoverage.CreatedAt, latestCoverage.CoveragePercent)

		// TODO: handle no results from db

		// TODO: check the current coverage against latest main branch coverage
	}

	return nil
}

func getCoveragePercent(counter CounterXML) float64 {
	return 100 * float64(counter.Covered) / float64(counter.Covered+counter.Missed)
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
	CoveragePercent float64 `json:"coverage_percent"`
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
