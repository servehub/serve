package dashboard

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/servehub/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("dashboard.kibana3", DashboardKibana3{})
}

type DashboardKibana3 struct{}

func (p DashboardKibana3) Run(data manifest.Manifest) error {
	if data.GetBool("purge") {
		req, _ := http.NewRequest("DELETE", data.GetString("elastic.url"), nil)

		resp, err := http.DefaultClient.Do(req)
		log.Println(resp)

		return err
	} else {
		exists, err := http.DefaultClient.Get(data.GetString("elastic.url"))
		if err != nil {
			return err
		}

		if exists.StatusCode != 404 {
			log.Println("Kibana dashboard already exists!")
			return nil
		}

		value := map[string]string{
			"user":      data.GetString("user"),
			"group":     data.GetString("group"),
			"title":     data.GetString("title"),
			"dashboard": data.GetTree("dashboard").String(),
		}

		bts, _ := json.MarshalIndent(value, "", "  ")
		req, _ := http.NewRequest("PUT", data.GetString("elastic.url"), bytes.NewReader(bts))

		resp, err := http.DefaultClient.Do(req)
		log.Println(resp)

		return err
	}
}
