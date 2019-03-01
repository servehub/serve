package plugins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/servehub/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("notify", Notify{})
}

type Notify struct{}

func (p Notify) Run(data manifest.Manifest) error {
	ch := data.GetTree("channels").GetTree(data.GetString("event"))

	switch ch.GetString("type") {
	case "slack":
		webhook := ch.GetString("webhook")

		if strings.HasPrefix(webhook, "$") {
			value, ok := os.LookupEnv(strings.TrimPrefix(webhook, "$"))
			if !ok || webhook == "" {
				return fmt.Errorf("Slack webhook url is empty: %s!", ch.GetString("webhook"))
			}
			webhook = value
		}

		payload := map[string]interface{}{
			"attachments": []map[string]string{
				{
					"color": ch.GetString("color"),
					"text":  ch.GetString("message"),
				},
			},
		}

		bts, _ := json.MarshalIndent(payload, "", "  ")
		req, _ := http.NewRequest("POST", webhook, bytes.NewReader(bts))

		_, err := http.DefaultClient.Do(req)
		return err

	default:
		return fmt.Errorf("Unknown event `%s`", data.GetString("event"))
	}
}
