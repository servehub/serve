package build

import (
	"testing"

	"github.com/servehub/serve/tests"
)

func TestDockerImageBuild(t *testing.T) {
	tests.RunAllMultiCmdTests(t,
		map[string]tests.TestCase{
			"simple": {
				In: `---
		      image: registry.superman.space/common/node:v1.0.34
		      tags: []
		      workdir: "."
		      build-args: "--pull"
		      environment: {}
		      login:
		        user: ""
		        password: ""
				`,
				Expects: []string{
					"docker build --pull -t registry.superman.space/common/node:v1.0.34 -t registry.superman.space/common/node:latest --cache-from=registry.superman.space/common/node:latest .",
					"docker push registry.superman.space/common/node:v1.0.34",
					"docker push registry.superman.space/common/node:latest",
				},
			},

			"multiple tags": {
				In: `---
		      image: registry.superman.space/common/node:v1.0.34
		      tags: [7, "7.10", latest]
		      workdir: 7
		      build-args: "--pull"
		      environment: {}
		      login:
		        user: ""
		        password: ""
				`,
				Expects: []string{
					"docker pull registry.superman.space/common/node:7",
					"docker pull registry.superman.space/common/node:7.10",
					"docker pull registry.superman.space/common/node:latest",
					"docker build --pull -t registry.superman.space/common/node:7 -t registry.superman.space/common/node:7.10 -t registry.superman.space/common/node:latest --cache-from=registry.superman.space/common/node:7 7",
					"docker push registry.superman.space/common/node:7",
					"docker push registry.superman.space/common/node:7.10",
					"docker push registry.superman.space/common/node:latest",
				},
			},

			"custom name": {
				In: `---
		      image: registry.superman.space/common/:v1.0.34
		      tags: [7, "7.0", latest]
		      name: php
		      workdir: 7
		      build-args: "--pull"
		      environment: {}
		      login:
		        user: ""
		        password: ""
				`,
				Expects: []string{
					"docker pull registry.superman.space/common/php:7",
					"docker pull registry.superman.space/common/php:7.0",
					"docker pull registry.superman.space/common/php:latest",
					"docker build --pull -t registry.superman.space/common/php:7 -t registry.superman.space/common/php:7.0 -t registry.superman.space/common/php:latest --cache-from=registry.superman.space/common/php:7 7",
					"docker push registry.superman.space/common/php:7",
					"docker push registry.superman.space/common/php:7.0",
					"docker push registry.superman.space/common/php:latest",
				},
			},

			"more complex custom name": {
				In: `---
		      image: registry.superman.space/web/common/:v0.0.0
		      name: "postgres-for-tests"
		      tags: [9.6, "latest"]
		      workdir: "postgres-for-tests/9.6"
		      build-args: "--pull"
		      environment: {}
		      login:
		        password: "${DOCKER_REGISTRY_PASSWORD}"
		        user: "${DOCKER_REGISTRY_USER}"
				`,
				Expects: []string{
					`docker login -u "${DOCKER_REGISTRY_USER}" -p "${DOCKER_REGISTRY_PASSWORD}" registry.superman.space`,
					"docker pull registry.superman.space/web/common/postgres-for-tests:9.6",
					"docker pull registry.superman.space/web/common/postgres-for-tests:latest",
					"docker build --pull -t registry.superman.space/web/common/postgres-for-tests:9.6 -t registry.superman.space/web/common/postgres-for-tests:latest --cache-from=registry.superman.space/web/common/postgres-for-tests:9.6 postgres-for-tests/9.6",
					"docker push registry.superman.space/web/common/postgres-for-tests:9.6",
					"docker push registry.superman.space/web/common/postgres-for-tests:latest",
				},
			},

			"custom category": {
				In: `---
		      image: registry.superman.space/web/common/:v0.0.0
		      name: "new-container"
		      category: "utility"
		      tags: [9.6, "latest"]
		      workdir: "new-container/9.6"
		      build-args: "--pull"
		      environment: {}
		      login:
		        user: ""
		        password: ""
				`,
				Expects: []string{
					"docker pull registry.superman.space/utility/new-container:9.6",
					"docker pull registry.superman.space/utility/new-container:latest",
					"docker build --pull -t registry.superman.space/utility/new-container:9.6 -t registry.superman.space/utility/new-container:latest --cache-from=registry.superman.space/utility/new-container:9.6 new-container/9.6",
					"docker push registry.superman.space/utility/new-container:9.6",
					"docker push registry.superman.space/utility/new-container:latest",
				},
			},

			"custom category 2": {
				In: `---
		      image: registry.superman.space/web/common/front/some-app:v0.0.0
		      name: "newyear"
		      category: "library/utils"
		      tags: latest
		      workdir: "."
		      build-args: "--pull"
		      environment: {}
		      login:
		        user: ""
		        password: ""
				`,
				Expects: []string{
					"docker pull registry.superman.space/library/utils/newyear:latest",
					"docker build --pull -t registry.superman.space/library/utils/newyear:latest --cache-from=registry.superman.space/library/utils/newyear:latest .",
					"docker push registry.superman.space/library/utils/newyear:latest",
				},
			},

			"docker login": {
				In: `---
		      image: registry.superman.space/common/node:v1.0.34
		      tags: []
		      workdir: "."
		      build-args: "--pull"
		      environment: {}
		      login:
		        user: "${DOCKER_REGISTRY_USER}"
		        password: "${DOCKER_REGISTRY_PASSWORD}"
				`,
				Expects: []string{
					`docker login -u "${DOCKER_REGISTRY_USER}" -p "${DOCKER_REGISTRY_PASSWORD}" registry.superman.space`,
					"docker build --pull -t registry.superman.space/common/node:v1.0.34 -t registry.superman.space/common/node:latest --cache-from=registry.superman.space/common/node:latest .",
					"docker push registry.superman.space/common/node:v1.0.34",
					"docker push registry.superman.space/common/node:latest",
				},
			},

			"custom dockerfile": {
				In: `---
		      image: registry.superman.space/common/node:v1.0.34
		      dockerfile: Dockerfile.stt
		      tags: []
		      workdir: "."
		      build-args: "--pull"
		      environment: {}
		      login:
		        user: ""
		        password: ""
				`,
				Expects: []string{
					"docker build --pull --file ./Dockerfile.stt -t registry.superman.space/common/node:v1.0.34 -t registry.superman.space/common/node:latest --cache-from=registry.superman.space/common/node:latest .",
					"docker push registry.superman.space/common/node:v1.0.34",
					"docker push registry.superman.space/common/node:latest",
				},
			},

			"custom dockerfile nested": {
				In: `---
		      image: registry.superman.space/common/node:v1.0.34
		      tags: []
		      workdir: "this_is_subdirectory"
		      current-branch: "master"
		      images:
		         - branch: "master"
		           dockerfile: "Dockerfile.stt"
		           repository: "docker.custom-repo.com/subdir/path"
		         - branch: "master"
		           dockerfile: "Dockerfile.stt"
		           repository: "new-other-repo.text.com"
		      build-args: "--pull"
		      environment: {}
		      login:
		        user: ""
		        password: ""
				`,
				Expects: []string{
					"docker build --pull -t registry.superman.space/common/node:v1.0.34 -t registry.superman.space/common/node:latest --cache-from=registry.superman.space/common/node:latest this_is_subdirectory",
					"docker push registry.superman.space/common/node:v1.0.34",
					"docker push registry.superman.space/common/node:latest",
				},
			},
		},
		BuildDockerImage{})
}
