package alerts

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"

	"github.com/kulikov/serve/manifest"
)

type (
	GraphiteAlertPlugin struct{}

	GraphiteManifest struct {
		Alerts []GraphiteAlerts `yaml:"alerts"`
	}

	GraphiteAlerts struct {
		Name     string         `yaml:"name"`
		Graphite *GraphiteAlert `yaml:"graphite"`
	}

	GraphiteAlert struct {
		Metric string `yaml:"metric"`
		From   string `yaml:"from"`

		Warn interface{} `yaml:"warn"`
		Crit interface{} `yaml:"crit"`

		WarnMax interface{} `yaml:"warn.max"`
		CritMax interface{} `yaml:"crit.max"`

		WarnMin interface{} `yaml:"warn.min"`
		CritMin interface{} `yaml:"crit.min"`
	}
)

func (ga GraphiteAlertPlugin) Run(conf *viper.Viper, manf *manifest.Manifest) error {
	log.Println("Run Graphite plugin")

	grmanf := GraphiteManifest{}
	yaml.Unmarshal(manf.Source, &grmanf)

	checks := make([]string, 0)

	for _, alert := range grmanf.Alerts {
		if g := alert.Graphite; g != nil {
			log.Printf("Graphite: %v\n", g)

			warn := ga.threshold(conf, "w", g.WarnMin, g.WarnMax, g.Warn)
			crit := ga.threshold(conf, "c", g.CritMin, g.CritMax, g.Crit)

			if warn != "" || crit != "" {
				checks = append(checks, fmt.Sprintf(
					`get_graphite_metric.sh -N services.%s.%s -f %d %s %s -m "%s" -D "%s"`,
					manf.Info.Name,
					regexp.MustCompile(`\W+`).ReplaceAllString(strings.ToLower(alert.Name), "-"),
					durationMillis(envVar(conf, g.From, "15m")),
					warn,
					crit,
					regexp.MustCompile(`\s+`).ReplaceAllString(strings.Replace(g.Metric, `"`, `'`, -1), ""),
					prepareChannel(conf.GetString("env"), manf.Notification.Channel),
				))
			}
		}
	}

	if len(checks) > 0 {
		return generateCheckMkFile(conf.GetString("alerts.graphite.filepath"), checks)
	} else {
		return nil
	}
}

func (ga GraphiteAlertPlugin) threshold(conf *viper.Viper, level string, vmin interface{}, vmax interface{}, vdef interface{}) string {
	result := ""

	if min := envVar(conf, vmin, envVar(conf, vdef)); min != "" {
		result += fmt.Sprintf("-%s %s ", strings.ToLower(level), min)
	}

	if max := envVar(conf, vmax); max != "" {
		result += fmt.Sprintf("-%s %s ", strings.ToUpper(level), max)
	}

	return result
}
