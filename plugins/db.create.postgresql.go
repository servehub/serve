package plugins

import (
	"fmt"

	"github.com/InnovaCo/serve/manifest"
	"github.com/InnovaCo/serve/utils"
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
		t := data.GetString("target")
		cmd = fmt.Sprintf("sudo -Hu postgres createdb -O %s \"%s\" && sudo -Hu postgres pg_dump \"%s\" | sudo -Hu postgres psql \"%s\"",
			              data.GetStringOr("db-user", "postgres"), t, data.GetString("source"), t)

	} else {
		cmd = fmt.Sprintf("sudo -Hu postgres createdb -O %s \"%s\"",
			              data.GetStringOr("db-user", "postgres"), data.GetString("target"))
	}

	return utils.RunSingleSshCmd(data.GetString("host"), data.GetString("ssh-user"), cmd)
}

func (p DBCreatePostgresql) Drop(data manifest.Manifest) error {
	return utils.RunSingleSshCmd(
		data.GetString("host"),
		data.GetString("ssh-user"),
		fmt.Sprintf("sudo -Hu postgres psql -c \\\"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname='%s';\\\" && sudo -Hu postgres dropdb --if-exists \"%s\"",
			data.GetString("target"), data.GetString("target")),
	)
}
