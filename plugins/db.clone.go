package plugins

import (
	"github.com/InnovaCo/serve/manifest"
	"fmt"
)

func init() {
	manifest.PluginRegestry.Add("db.clone", DBClone{})
}

type DBClone struct{}

func (p DBClone) Run(data manifest.Manifest) error {
	if data.GetBool("purge") {
		return p.Drop(data)
	} else {
		return p.Clone(data)
	}

}

func (p DBClone) Clone(data manifest.Manifest) error {
	err := runSshCmd(
		data.GetString("server"),
		data.GetString("ssh-user"),
		fmt.Sprintf("sudo -u postgres createdb -O postgres %s && pg_dump %s | psql %s", data.GetString("to"), data.GetString("from"), data.GetString("to")),
	)
	if err != nil {
		// ToDo analize db exist
		return err
	}
	return nil
}

func (p DBClone) Drop(data manifest.Manifest) error {
	return runSshCmd(
		data.GetString("server"),
		data.GetString("ssh-user"),
		fmt.Sprintf("sudo -u postgres dropdb %s", data.GetString("to")),
	)
}
