package alerts

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var dayRegex = regexp.MustCompile(`(\d+)\s*(day|days)`)
var hourRegex = regexp.MustCompile(`(\d+)\s*(hr|hour|hours)`)
var minRegex = regexp.MustCompile(`(\d+)\s*(min|mins|minute|minutes)`)
var secondRegex = regexp.MustCompile(`(\d+)\s*(sec|secs|second|seconds)`)

func envVar(conf *viper.Viper, val interface{}, def ...string) string {
	res := ""

	switch v := val.(type) {
	case map[string]string:
		res = v[conf.GetString("env")]
	case map[interface{}]string:
		res = v[conf.GetString("env")]
	case map[string]int:
		res = fmt.Sprintf("%v", v[conf.GetString("env")])
	case map[interface{}]int:
		res = fmt.Sprintf("%v", v[conf.GetString("env")])
	case map[string]float32:
		res = fmt.Sprintf("%v", v[conf.GetString("env")])
	case map[interface{}]float32:
		res = fmt.Sprintf("%v", v[conf.GetString("env")])
	case map[string]interface{}:
		res = fmt.Sprintf("%v", v[conf.GetString("env")])
	case map[interface{}]interface{}:
		res = fmt.Sprintf("%v", v[conf.GetString("env")])
	default:
		if val != nil {
			res = fmt.Sprintf("%v", v)
		}
	}

	if res != "" && res != "<nil>" {
		return res
	} else if len(def) > 0 {
		return def[0]
	} else {
		return ""
	}
}

func durationMillis(duration string) int {
	duration = dayRegex.ReplaceAllStringFunc(duration, func(d string) string {
		di, _ := strconv.Atoi(d)
		return fmt.Sprintf("%vh", di*24)
	})

	duration = hourRegex.ReplaceAllString(duration, "${1}h")
	duration = minRegex.ReplaceAllString(duration, "${1}m")
	duration = secondRegex.ReplaceAllString(duration, "${1}s")

	from, err := time.ParseDuration(duration)
	if err != nil {
		log.Println("Error on parse duration: " + duration)
	}
	return int(from.Seconds() * 1000)
}

func generateCheckMkFile(path string, checks []string) error {
	return ioutil.WriteFile(path, []byte("#!/bin/bash\n\n"+strings.Join(checks, "\n\n")), 0755)
}

func prepareChannel(env string, channel string) string {
	if env == "live" && channel != "" {
		return "#" + strings.Trim(channel, " #")
	} else {
		return ""
	}
}
