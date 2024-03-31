package config

type EnvVar struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPass     string
	DBName     string
	LoginURL   string
	LoginToken string
	PlsURL     string
	PlsSystem  string
	ServerPort string
	ServerURL  string
}

func GetEnv() EnvVar {

	envConfig := EnvVar{
		DBHost:     "localhost",
		DBPort:     "5432",
		DBUser:     "alexander",
		DBPass:     "kopis",
		DBName:     "budget_local",
		LoginURL:   "login2798.datasektionen.se",
		LoginToken: "",
		PlsURL:     "pls.datasektionen.se",
		PlsSystem:  "GOrdian",
		ServerPort: "3000",
		ServerURL:  "http://localhost:3000",
	}

	return envConfig
}
