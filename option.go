package untold

type Option func(v *vault)

func Environment(environment string) Option {
	return func(v *vault) {
		v.environment = environment
	}
}

func EnvVariable(name string) Option {
	return func(v *vault) {
		v.privateKeyEnv = name
	}
}

func PathPrefix(prefix string) Option {
	return func(v *vault) {
		v.pathPrefix = prefix
	}
}
