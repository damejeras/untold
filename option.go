package untold

type Option func(l *loader)

func Environment(environment string) Option {
	return func(l *loader) {
		l.environment = environment
	}
}

func EnvVariable(name string) Option {
	return func(l *loader) {
		l.privateKeyEnv = name
	}
}

func PathPrefix(prefix string) Option {
	return func(l *loader) {
		l.pathPrefix = prefix
	}
}
