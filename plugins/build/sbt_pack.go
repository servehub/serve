package build

//import (
//	"github.com/InnovaCo/serve/manifest"
//	"github.com/InnovaCo/serve/utils"
//)
//
//type SbtPackBuild struct{}
//
//func (_ SbtPackBuild) Run(m *manifest.Manifest, sub *manifest.Manifest) error {
//	if err := utils.RunCmdf("sbt ';set version := \"%s\"' clean test pack", m.BuildVersion()); err != nil {
//		return err
//	}
//
//	return nil
//}
