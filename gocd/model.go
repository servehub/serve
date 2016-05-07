package gocd

type (
	CreatePipline struct {
		Group    string   `json:"group"`
		Pipeline Pipeline `json:"pipeline"`
	}

	Pipeline struct {
		Name         string        `json:"name"`
		Materials    []Material    `json:"materials"`
		EnvVariables []EnvVariable `json:"environment_variables,omitempty"`
		Stages       []Stage       `json:"stages"`
	}

	Material struct {
		Type       string        `json:"type"`
		Attributes MaterialAttrs `json:"attributes"`
	}

	MaterialAttrs struct {
		Url         string `json:"url"`
		Destination string `json:"destination"`
	}

	EnvVariable struct {
		Name           string `json:"name"`
		Value          string `json:"value"`
		EncryptedValue string `json:"encrypted_value"`
		Secure         bool   `json:"secure"`
	}

	Stage struct {
		Name                  string   `json:"name"`
		CleanWorkingDirectory bool     `json:"clean_working_directory"`
		FetchMaterials        bool     `json:"fetch_materials"`
		Approval              Approval `json:"approval,omitempty"`
		Jobs                  []Job    `json:"jobs"`
	}

	Approval struct {
		Type          string        `json:"type,omitempty"`
		Authorization Authorization `json:"authorization,omitempty"`
	}

	Authorization struct {
		Roles []string `json:"roles,omitempty"`
	}

	Job struct {
		Name      string   `json:"name"`
		Resources []string `json:"resources"`
		Tasks     []Task   `json:"tasks"`
	}

	Task struct {
		Type       string         `json:"type"`
		Attributes TaskAttributes `json:"attributes"`
	}

	TaskAttributes struct {
		Command          string   `json:"type"`
		Arguments        []string `json:"arguments"`
		RunIf            []string `json:"run_if"`
		WorkingDirectory string   `json:"working_directory"`
	}
)
