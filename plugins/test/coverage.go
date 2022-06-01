package test

import (
	"errors"
	"github.com/servehub/utils"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/servehub/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("test.coverage", TestCoverageUpload{})
}

type TestCoverageUpload struct{}

func (p TestCoverageUpload) Run(data manifest.Manifest) error {

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
		CoverageFile []byte
	}

	// get repo, commit information
	meta := Meta{
		Repo:     data.GetString("repo"),
		Branch:   data.GetString("branch"),
		Ref:      data.GetString("ref"),
		Version:  data.GetString("version"),
		TestType: data.GetString("test-type"),
	}

	if generateCmd := data.GetString("generate"); generateCmd != "" {
		utils.RunCmd(generateCmd)
	}

	coverageFile, err := os.ReadFile(data.GetString("coverage-file"))
	if err != nil {
		return errors.New("failed to read coverage file")
	}

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

	db.Create(&CoverageReport{Meta: meta, CoverageFile: coverageFile})

	return nil
}
