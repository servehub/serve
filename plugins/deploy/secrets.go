package deploy

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/servehub/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("deploy.secrets", DeploySecrets{})
}

type DeploySecrets struct{}

func (p DeploySecrets) Run(data manifest.Manifest) error {
	data.DelTree("env")
	data.DelTree("consul")

	manifestBytes, _ := ioutil.ReadFile("/Users/kulikov/Work/copper/infra/provisioning/manifest.yml")
  manifest := string(manifestBytes)

	for key, sec := range data.GetMap(".") {
		for _, env := range []string{"qa", "stage", "live"} {
			if sec.Has("value." + env) {
				oldValue := sec.GetString("value."+env)

				println(key + ":")

				cmd := exec.Command("/bin/bash", "-ec", `echo -n "` + oldValue + `" | base64 -D | openssl smime -decrypt -aes256 -inform pem -inkey /Users/kulikov/Work/copper/infra/provisioning/.secrets/marathon-secrets-` + env + `.key`)
				cmd.Stderr = os.Stderr

				out, err := cmd.Output()
				if err != nil {
					log.Fatal(err)
				}

				plainvalue := strings.TrimSpace(string(out))

				cmd2 := exec.Command("/bin/bash", "-ec", `echo -n "` + plainvalue + `" | openssl rsautl -encrypt -inkey /Users/kulikov/Work/copper/infra/provisioning/keys/secrets-`+env+`-public.key -pubin | base64`)
				cmd2.Stderr = os.Stderr

				out2, err2 := cmd2.Output()
				if err2 != nil {
					log.Fatal(err2)
				}

				newValue := strings.TrimSpace(string(out2))

				manifest = strings.Replace(manifest, oldValue, newValue, -1)

				println("")

			}
		}
	}

	return ioutil.WriteFile("/Users/kulikov/Work/copper/infra/provisioning/new-manifest.yml", []byte(manifest), 0644)
}
