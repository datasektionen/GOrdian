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
		LoginURL:   "https://login.datasektionen.se",
		LoginToken: "",
		PlsURL:     "https://pls.datasektionen.se",
		PlsSystem:  "gordian",
		ServerPort: "3000",
		ServerURL:  "http://localhost:3000",
	}

	return envConfig
}
