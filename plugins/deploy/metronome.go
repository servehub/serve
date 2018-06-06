package deploy

import (
	"fmt"
	"log"
	"strings"

	"github.com/adobe-platform/go-metronome/metronome"

	"github.com/servehub/serve/manifest"
)

func init() {
	manifest.PluginRegestry.Add("deploy.metronome", DeployMetronome{})
}

type DeployMetronome struct{}

func (p DeployMetronome) Run(data manifest.Manifest) error {
	if !data.GetBool("enabled") {
		return nil
	}

	config := metronome.NewDefaultConfig()
	config.URL = data.GetString("metronome-address")

	client, err := metronome.NewClient(config)
	if err != nil {
		return fmt.Errorf("Error on create metronome client: %v", err)
	}

	jobReq := &metronome.Job{
		ID: data.GetString("id"),
		Run: &metronome.Run{
			Cmd:            data.GetString("cmd"),
			Cpus:           data.GetFloat("cpu"),
			Mem:            data.GetInt("mem"),
			Disk:           data.GetInt("disk"),
			Volumes:        make([]metronome.Volume, 0),
			MaxLaunchDelay: 1200,
			Docker: &metronome.Docker{
				Image: data.GetString("docker.image"),
			},
		},
	}

	job, err := client.CreateJob(jobReq)

	if err != nil && strings.Contains(err.Error(), "Job with this id already exists") {
		updatedJob, err := client.UpdateJob(jobReq.ID, jobReq)

		if err != nil {
			return fmt.Errorf("Error on update job: %v", err)
		}

		log.Printf("Updated job: %s", updatedJob)
	} else if err != nil {
		return fmt.Errorf("Error on create new job: %v", err)
	} else {
		log.Printf("Created job: %v", job)
	}

	for i, sch := range data.GetArray("schedules") {
		schReq := &metronome.Schedule{
			ID:                      fmt.Sprintf("%s-%d", jobReq.ID, i),
			Cron:                    sch.GetString("cron"),
			ConcurrencyPolicy:       sch.GetStringOr("concurrency-policy", "ALLOW"),
			Enabled:                 true,
			StartingDeadlineSeconds: sch.GetIntOr("starting-deadline-seconds", 60),
			Timezone:                sch.GetStringOr("timezone", "UTC"),
		}

		schedule, err := client.CreateSchedule(jobReq.ID, schReq)

		if err != nil && strings.Contains(err.Error(), "A schedule with id "+schReq.ID+" already exists") {
			updatedSchedule, err := client.UpdateSchedule(jobReq.ID, schReq.ID, schReq)

			if err != nil {
				return fmt.Errorf("Error on update schedule: %v", err)
			}

			log.Printf("Updated schedule: %s", updatedSchedule)

		} else if err != nil {
			return fmt.Errorf("metronome CreateSchedule error: %v", err)
		} else {
			log.Printf("Created schedule: %v", schedule)
		}
	}

	return nil
}
