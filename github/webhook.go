package github

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/codegangsta/cli"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/kulikov/hookup"
	"github.com/kulikov/serve/gocd"
	"github.com/kulikov/serve/manifest"
	"github.com/kulikov/serve/manifest/alerts"
	"github.com/kulikov/serve/utils"
)

type (
	Plugin interface {
		Run(conf *viper.Viper, manf *manifest.Manifest) error
	}
)

var manifestPlugins = []Plugin{
	gocd.DeployPlugin{},
	alerts.GraphiteAlertPlugin{},
	alerts.ElasticAlertPlugin{},
}

func WebhookServerCommand() cli.Command {
	return cli.Command{
		Name:  "webhook-server",
		Usage: "Start webhook http sever and handle github hook event for check manifest changes",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "port",
				Value: "9090",
			},
			cli.StringFlag{
				Name:  "config",
				Value: "config.yml",
				Usage: "Path to config.yml file",
			},
		},
		Action: func(c *cli.Context) {
			conf, err := InitConfig(c.String("config"))
			if err != nil {
				log.Fatalf("Error load config: %v", err)
			}

			hookup.StartWebhookServer(c.Int("port"), func(source string, eventType string, payload string) {
				if source == "github" && eventType == "push" {
					if err := HandleGithubChanges(conf, manifestPlugins, payload); err != nil {
						log.Printf("Error %v", err)
					}
				}
			})
		},
	}
}

func HandleGithubChanges(conf *viper.Viper, plugins []Plugin, payload string) error {
	event := &Push{}
	json.Unmarshal([]byte(payload), event)

	modified := false
	for _, commit := range event.Commits {
		log.Println("Changes: ", append(commit.Added, commit.Modified...))

		if utils.Contains("manifest.yml", append(commit.Added, commit.Modified...)) {
			modified = true
		}
	}

	if modified {
		resp, err := http.Get(strings.Replace(event.Repository.ContentUrl, "{+path}", "manifest.yml", 1))
		defer resp.Body.Close()

		if err != nil {
			return err
		}

		file := &FileContent{}
		data, _ := ioutil.ReadAll(resp.Body)

		err = json.Unmarshal(data, file)
		if err != nil {
			return err
		}

		data, err = base64.StdEncoding.DecodeString(file.Content)
		if err != nil {
			return err
		}

		manf := &manifest.Manifest{Sha: file.Sha, GitSshUrl: event.Repository.SshUrl, Source: data}
		yaml.Unmarshal(data, manf)

		RunPlugins(conf, plugins, manf)
	} else {
		log.Println("manifest.yml not changed")
	}

	return nil
}

func RunPlugins(conf *viper.Viper, plugins []Plugin, manf *manifest.Manifest) {
	wg := sync.WaitGroup{}

	for _, plugin := range plugins {
		wg.Add(1)

		go func(p Plugin) {
			defer wg.Done()

			err := p.Run(conf, manf)

			if err != nil {
				log.Printf("%T: %s\n", p, err)
			}
		}(plugin)
	}

	wg.Wait()
}

func InitConfig(configFile string) (*viper.Viper, error) {
	conf := viper.New()
	conf.SetConfigType("yml")

	for _, file := range strings.Split(configFile, ",") {
		conf.SetConfigFile(file)

		if err := conf.MergeInConfig(); err != nil {
			return nil, err
		}
	}

	return conf, nil
}
