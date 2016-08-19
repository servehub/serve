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
	f := data.GetString("from")
	t := data.GetString("to")

	err := runSshCmd(
		data.GetString("server"),
		data.GetString("ssh-user"),
		fmt.Sprintf("sudo -EHu postgres createdb -O "+
			"`psql postgres -c \"SELECT d.datname, pg_catalog.pg_get_userbyid(d.datdba) FROM pg_catalog.pg_database d "+
			"WHERE d.datname='%s' ORDER BY 1;\" | grep %s | awk '{print $3}'` %s && pg_dump %s | psql %s", f, f, t, f, t),
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
		fmt.Sprintf("sudo -EHu postgres dropdb %s", data.GetString("to")),
	)
}
