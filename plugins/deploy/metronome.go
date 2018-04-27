package deploy

import (
	"fmt"

	"github.com/adobe-platform/go-metronome/metronome"

	"github.com/servehub/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("deploy.metronome", DeployMetronome{})
}

type DeployMetronome struct{}

func (p DeployMetronome) Run(data manifest.Manifest) error {
	config := metronome.NewDefaultConfig()
	config.URL = "http://" + data.GetString("metronome-address")
	client, err := metronome.NewClient(config)
	if err != nil {
		return err
	}

	job, err := client.CreateJob(&metronome.Job{
		ID: data.GetString("id"),
		Run: &metronome.Run{
			Cmd:  data.GetString("cmd"),
			Cpus: data.GetFloat("cpu"),
			Mem:  data.GetInt("mem"),
			Disk: data.GetInt("disk"),
			Docker: &metronome.Docker{
				Image: data.GetString("docker.image"),
			},
		},
	})

	for i, sch := range data.GetArray("schedules") {
		_, err := client.CreateSchedule(job.ID, &metronome.Schedule{
			ID:                      fmt.Sprintf("%s-%d", job.ID, i),
			Cron:                    sch.GetString("cron"),
			ConcurrencyPolicy:       sch.GetStringOr("concurrency-policy", ""),
			Enabled:                 true,
			StartingDeadlineSeconds: sch.GetIntOr("starting-deadline-seconds", 60),
			Timezone:                sch.GetStringOr("timezone", "UTC"),
		})

		if err != nil {
			return fmt.Errorf("metronome CreateSchedule error: %v", err)
		}
	}

	if err != nil {
		return fmt.Errorf("metronome CreateJob error: %v", err)
	}

	return nil
}
