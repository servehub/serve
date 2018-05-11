package deploy

import (
	"bytes"
	"fmt"
	"log"
	"strings"
	"text/template"

	"github.com/servehub/serve/manifest"
	"github.com/servehub/utils"
)

func init() {
	manifest.PluginRegestry.Add("deploy.copy-db", DeployCopyDb{})
}

type DeployCopyDb struct{}

func (p DeployCopyDb) Run(data manifest.Manifest) error {
	if !data.GetBool("enabled") {
		log.Printf("Skip copy-db: disabled for this environment")
		return nil
	}

	if data.GetBool("purge") {
		return p.Purge(data)
	} else {
		return p.Create(data)
	}
}

func (p DeployCopyDb) Create(data manifest.Manifest) error {
	cmd, err := applyTemplate(data.GetString("create-command"), data.Unwrap())
	if err != nil {
		return err
	}

	return utils.RunSingleSshCmd(
		data.GetString("ssh.host"),
		data.GetString("ssh.user"),
		cmd,
	)
}

func (p DeployCopyDb) Purge(data manifest.Manifest) error {
	cmd, err := applyTemplate(data.GetString("purge-command"), data.Unwrap())
	if err != nil {
		return err
	}

	return utils.RunSingleSshCmd(
		data.GetString("ssh.host"),
		data.GetString("ssh.user"),
		cmd,
	)
}

func applyTemplate(cmd string, data interface{}) (string, error) {
	t, err := template.New(cmd).Delims("{", "}").Parse(cmd)
	if err != nil {
		return "", fmt.Errorf("Error on template command `%v`: %v", cmd, data)
	}

	var tplout bytes.Buffer
	if err := t.Execute(&tplout, data); err != nil {
		return "", fmt.Errorf("Error on execute template command `%v`: %v", cmd, data)
	}

	return strings.TrimSpace(tplout.String()), nil
}
