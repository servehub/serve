package test

import (
	"errors"
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
		TestType: data.GetString("test-type"),
	}

	coverage_filename := data.GetString("coverage-file")
	coverage_file, err := os.ReadFile(coverage_filename)
	if err != nil {
		return errors.New("failed to read coverage file")
	}

	// https://github.com/jackc/pgx#example-usage
	// DSN should be in format:
	// "postgres://username:password@localhost:5432/database_name"
	//  -- or --
	// "host=localhost user=postgres password=postgres port=5432"
	dsn := os.Getenv("DATABASE_URL")

	// https://gorm.io/docs/connecting_to_the_database.html#PostgreSQL
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return errors.New("failed to connect to database")
	}

	// Migrate the DB schema
	db.AutoMigrate(&CoverageReport{})

	db.Create(&CoverageReport{Meta: meta, CoverageFile: coverage_file})

	return nil
}
