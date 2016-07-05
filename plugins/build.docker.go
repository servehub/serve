package plugins

//import (
//	"github.com/InnovaCo/serve/manifest"
//	"github.com/InnovaCo/serve/utils"
//	"github.com/docker/engine-api/client"
//)
//
//type DockerBuild struct{}
//
//func (_ DockerBuild) Run(m *manifest.Manifest, sub *manifest.Manifest) error {
//	c, _ := client.NewClient("", nil, nil, nil)
//
//	if err := utils.RunCmdf(`docker build --build-arg SSH_KEY="$(< ~/.ssh/id_rsa)" .`); err != nil {
//		return err
//	}
//
//	return nil
//}
