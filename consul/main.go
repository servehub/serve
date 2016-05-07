package consul

import (
	"regexp"

	"github.com/codegangsta/cli"
)

func ConsulCommand() cli.Command {
	return cli.Command{
		Name: "consul",
		Subcommands: []cli.Command{
			SupervisorCommand(),
			NginxTemplateContextCommand(),
		},
	}
}

var tagSplitRegex = regexp.MustCompile(":")

func ParseTags(tags []string) map[string]string {
	output := make(map[string]string)
	for _, t := range tags {
		tt := tagSplitRegex.Split(t, 2)
		if len(tt) > 1 {
			output[tt[0]] = tt[1]
		}
	}
	return output
}

func TagsFromFlags(c *cli.Context) map[string]string {
	tags := make(map[string]string, 0)

	if t := c.GlobalString("version"); t != "" {
		tags["version"] = t
	}

	if t := c.GlobalString("domain"); t != "" {
		tags["domain"] = t
	}

	if t := c.GlobalString("location"); t != "" {
		tags["location"] = t
	}

	if t := c.GlobalString("staging"); t != "" {
		tags["staging"] = t
	}

	return tags
}

func MapToList(m map[string]string) []string {
	out := make([]string, 0)
	for k, v := range m {
		out = append(out, k+":"+v)
	}
	return out
}
