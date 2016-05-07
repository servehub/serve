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
	ElasticAlertPlugin struct{}

	ElasticManifest struct {
		Alerts []ElasticAlerts `yaml:"alerts"`
	}

	ElasticAlerts struct {
		Name    string        `yaml:"name"`
		Elastic *ElasticAlert `yaml:"elastic"`
	}

	ElasticAlert struct {
		Query string `yaml:"query"`
		From  string `yaml:"from"`

		Warn interface{} `yaml:"warn"`
		Crit interface{} `yaml:"crit"`

		WarnMax interface{} `yaml:"warn.max"`
		CritMax interface{} `yaml:"crit.max"`

		WarnMin interface{} `yaml:"warn.min"`
		CritMin interface{} `yaml:"crit.min"`
	}
)

func (ea ElasticAlertPlugin) Run(conf *viper.Viper, manf *manifest.Manifest) error {
	log.Println("Run Elastic plugin")

	elmanf := ElasticManifest{}
	yaml.Unmarshal(manf.Source, &elmanf)

	checks := make([]string, 0)

	for _, alert := range elmanf.Alerts {
		if el := alert.Elastic; el != nil {
			log.Printf("Elastic: %v\n", el)

			warn := ea.threshold(conf, "w", el.WarnMin, el.WarnMax, el.Warn)
			crit := ea.threshold(conf, "c", el.CritMin, el.CritMax, el.Crit)

			if warn != "" || crit != "" {
				checks = append(checks, fmt.Sprintf(
					`result=$(check_json.pl %s %s --url '%s/logstash-*/_search?q=`+
						`(%s) AND timemillis:['$(( ($(date +%%s) * 1000) - %v ))' TO '$(($(date +%%s) * 1000))']&search_type=count' `+
						`--attribute '{hits}->{total}' `+
						"--perfvars '{hits}->{total}') \n"+
						`echo "$? services.%s.%s perfdata=$(echo $result | sed 's/.*- total: \([0-9]*\).*/\\1/').0 `+
						`total=$(echo $result | sed 's/.*total=\(.*\).*/\\1/')' by query %s %s';"`,
					warn,
					crit,
					conf.GetString("alerts.elastic.host"),
					strings.Replace(el.Query, `'`, `'\\''`, -1),
					durationMillis(envVar(conf, el.From, "15m")),
					manf.Info.Name,
					regexp.MustCompile(`\W+`).ReplaceAllString(strings.ToLower(alert.Name), "-"),
					regexp.MustCompile(`[^\w\s:\-\.\(\)]+`).ReplaceAllString(el.Query, ""),
					prepareChannel(conf.GetString("env"), manf.Notification.Channel),
				))
			}
		}
	}

	if len(checks) > 0 {
		return generateCheckMkFile(conf.GetString("alerts.elastic.filepath"), checks)
	} else {
		return nil
	}
}

func (ea ElasticAlertPlugin) threshold(conf *viper.Viper, level string, vmin interface{}, vmax interface{}, vdef interface{}) string {
	min := envVar(conf, vmin)
	max := envVar(conf, vmax, envVar(conf, vdef))

	if min != "" || max != "" {
		return fmt.Sprintf("-%s %v:%v", level, min, max)
	} else {
		return ""
	}
}
