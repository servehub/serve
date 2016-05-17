package supervisor

type Supervisor struct {
	MaxRetry   int
	OnError    func()
	OnComplete func()
}

func (s Supervisor) Run(cmdArgs []string) {
	//
}
