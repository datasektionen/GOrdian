package config

import "os"

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
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPass:     os.Getenv("DB_PASS"),
		DBName:     os.Getenv("DB_NAME"),
		LoginURL:   os.Getenv("LOGIN_URL"),
		LoginToken: os.Getenv("LOGIN_TOKEN"),
		PlsURL:     os.Getenv("PLS_URL"),
		PlsSystem:  os.Getenv("PLS_SYSTEM"),
		ServerPort: os.Getenv("SERVER_PORT"),
		ServerURL:  os.Getenv("SERVER_URL"),
	}

	return envConfig
}

//example
//		DBHost:     localhost,
//		DBPort:     "5432",
//		DBUser:     "alexander",
//		DBPass:     "kopis",
//		DBName:     "budget_local",
//		LoginURL:   "https://login.datasektionen.se",
//		LoginToken: "this is secret fuck you",
//		PlsURL:     "https://pls.datasektionen.se",
//		PlsSystem:  "gordian",
//		ServerPort: "3000",
//		ServerURL:  "http://localhost:3000",
