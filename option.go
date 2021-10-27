package untold

type Option func(l *vault)

func Environment(environment string) Option {
	return func(l *vault) {
		l.environment = environment
	}
}

func EnvVariable(name string) Option {
	return func(l *vault) {
		l.privateKeyEnv = name
	}
}

func PathPrefix(prefix string) Option {
	return func(l *vault) {
		l.pathPrefix = prefix
	}
}
