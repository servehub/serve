package manifest

type (
	Manifest struct {
		Sha       string
		GitSshUrl string
		Source    []byte

		Info         Info         `yaml:"info"`
		Notification Notification `yaml:"notification"`
	}

	Info struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
		Owner   Owner  `yaml:"owner"`
	}

	Owner struct {
		Name  string `yaml:"name"`
		Email string `yaml:"email"`
	}

	Notification struct {
		Channel string `yaml:"channel"`
	}
)
