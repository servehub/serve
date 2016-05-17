package app

import (
	"log"

	"github.com/codegangsta/cli"
	"github.com/InnovaCo/serve/manifest"
)

func ReleaseCommand() cli.Command {
	return cli.Command{
		Name:  "release",
		Usage: "Release service",
		Flags: []cli.Flag{
			cli.StringFlag{Name: "env"},
			cli.StringFlag{Name: "branch"},
			cli.StringFlag{Name: "build-number"},
			cli.StringFlag{Name: "route"},
		},
		Action: func(c *cli.Context) error {
			mf := manifest.LoadManifest(c)

			if mf.Has("deploy") {
				strategy, err := GetStrategy("release", mf.GetStringOr("deploy.type", "default"))

				if err != nil {
					log.Fatalf("Release error: %v", err)
				}

				return strategy.Run(mf, mf.Sub("deploy"))
			}

			return nil

			// вынести это все в отдельную стратегию — site
			// там для мастер и для фича-веток сделать одинаковое поведение — -v1.0.34 прибавлять, в марафоне искать старую версию и т.п.

			// находим текущий в консуле и убеждаемся что с ним все ок
			// добавляем ему роуты

			// ищем есть ли старый с такими же роутами:
			//    формируем массив роутов
			//    ищем сервис с таким-же именем но другой версии, и содержащий один из указанных роутов
			//    например в kv можно хранить /kv/services/{name-?branch}/v{version}} и там матчить через compareMaps
			//    если хотябы один роут полностью совпал — это наш кандидат на убивание
			// если есть — убиваем в консуле сразу и через 5 минут в марафоне

			println(compareMaps(map[string]string{"name": "dima", "version": "1.0"}, map[string]string{"version": "1.0", "name": "dima"}))

			log.Println("route")

			return nil
		},
	}
}

func compareMaps(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if w, ok := b[k]; !ok || v != w {
			return false
		}
	}

	return true
}
