package config

func Initialize() {
	initBaseVariables()
	initCORSVariables()
	initMongoVariables()
}

func Check() []error {
	errs := make([]error, 0)

	errs = append(errs, checkMongoEnvs()...)

	return errs
}
