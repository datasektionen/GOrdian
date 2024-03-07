package config

type EnvVar struct {
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string
}

func GetEnv() EnvVar {

	envConfig := EnvVar{
		DBHost: "localhost",
		DBPort: "5432",
		DBUser: "alexander",
		DBPass: "kopis",
		DBName: "budget_local",
	}

	return envConfig
}
