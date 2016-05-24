package marathon

import (
	"fmt"
	"log"
	"os"

	"github.com/fatih/color"
	"github.com/gambol99/go-marathon"

	"github.com/InnovaCo/serve/manifest"
)

func MarathonClient(m *manifest.Manifest) marathon.Marathon {
	conf := marathon.NewDefaultConfig()
	conf.URL = fmt.Sprintf("http://%s:8080", m.GetString("marathon.marathon-host"))
	conf.LogOutput = os.Stdout
	marathonApi, err := marathon.NewClient(conf)
	if err != nil {
		log.Fatalf(color.RedString("Error on create marathon api client: %v", err))
	}
	return marathonApi
}
