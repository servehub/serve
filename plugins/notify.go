package plugins

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
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
				return fmt.Errorf("slack webhook url is empty: %s", ch.GetString("webhook"))
			}
			webhook = value
		}

		attachments := []map[string]string{
			{
				"color": ch.GetString("color"),
				"text":  ch.GetString("message"),
			},
		}

		changelogFor := ch.GetStringOr("changelog-for", "")

		if changelogFor != "" {
			consul, _ := utils.ConsulClient(data.GetString("consul.address"))

			if pair, _, err := consul.KV().Get(data.GetString("consul.path"), nil); err == nil && pair != nil {
				latestRelease := latestRelease{}
				if err := json.Unmarshal(pair.Value, &latestRelease); err != nil {
					return err
				}

				gitLogCmd := fmt.Sprintf(`git log --pretty=format:" — %%an: %%s" %s..%s`, latestRelease.CommitHash, changelogFor)
				log.Println(color.YellowString("> %s", gitLogCmd))

				out, err := exec.Command("/bin/bash", "-ec", gitLogCmd).CombinedOutput()
				if err != nil {
					if strings.Contains(string(out), "Invalid revision range") {
						log.Println(color.YellowString("> git fetch --depth=100"))

						if err2 := exec.Command("/bin/bash", "-ec", "git fetch --depth=100").Run(); err2 == nil {
							out, err = exec.Command("/bin/bash", "-ec", gitLogCmd).CombinedOutput()
						}
					}

					if err != nil {
						log.Println(color.RedString("%s", out))
					}
				}

				attachments = append(attachments, map[string]string{
					"color": "#CCC",
					"text":  string(out),
				})
			}

			bts, _ := json.Marshal(latestRelease{changelogFor})
			if err := utils.PutConsulKv(consul, data.GetString("consul.path"), string(bts)); err != nil {
				return err
			}
		}

		bts, _ := json.MarshalIndent(map[string]interface{}{
			"attachments": attachments,
		}, "", "  ")

		req, _ := http.NewRequest("POST", webhook, bytes.NewReader(bts))
		_, err := http.DefaultClient.Do(req)
		return err

	default:
		return fmt.Errorf("Unknown event `%s`", data.GetString("event"))
	}
}

type latestRelease struct {
	CommitHash string `json:"commitHash,omitempty"`
}
