package plugins

import (
	"fmt"

	"github.com/InnovaCo/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("db.create.postgresql", DBCreatePostgresql{})
}

type DBCreatePostgresql struct{}

func (p DBCreatePostgresql) Run(data manifest.Manifest) error {
	if data.GetBool("purge") {
		return p.Drop(data)
	} else {
		return p.Create(data)
	}
}

func (p DBCreatePostgresql) Create(data manifest.Manifest) error {
	var cmd string

	if data.Has("source") {
		s := data.GetString("source")
		t := data.GetString("target")
		cmd = fmt.Sprintf("sudo -EHu postgres createdb -O "+
			"`psql postgres -c \"SELECT d.datname, pg_catalog.pg_get_userbyid(d.datdba) FROM pg_catalog.pg_database d "+
			"WHERE d.datname='%s' ORDER BY 1;\" | grep %s | awk '{print $3}'` \"%s\" && pg_dump \"%s\" | psql \"%s\"", s, s, t, s, t)

	} else {
		cmd = fmt.Sprintf("sudo -EHu postgres createdb -O %s \"%s\"", data.GetStringOr("db-user", "postgres"), data.GetString("target"))
	}

	return runSshCmdSingle(data.GetString("host"), data.GetString("ssh-user"), cmd)
}

func (p DBCreatePostgresql) Drop(data manifest.Manifest) error {
	return runSshCmdSingle(
		data.GetString("host"),
		data.GetString("ssh-user"),
		fmt.Sprintf("sudo -EHu postgres dropdb \"%s\"", data.GetString("target")),
	)
}
