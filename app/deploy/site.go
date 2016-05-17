package deploy

import (
	"log"

	"github.com/InnovaCo/serve/manifest"
)

type SiteDeploy struct {}
type SiteRelease struct {}

func (_ SiteDeploy) Run(m *manifest.Manifest, sub *manifest.Manifest) error {
	log.Println("Deploy done!", sub)
	return nil
}

func (_ SiteRelease) Run(m *manifest.Manifest, sub *manifest.Manifest) error {
	log.Println("Release done!", sub)
	return nil
}
