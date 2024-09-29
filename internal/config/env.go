package config

import "os"

type EnvVar struct {
	PsqlconnStringGOrdian  string
	PsqlconnStringCashflow string
	LoginURL               string
	LoginToken             string
	PlsURL                 string
	PlsSystem              string
	ServerPort             string
	ServerURL              string
}

func GetEnv() EnvVar {

	envConfig := EnvVar{
		//prod
		PsqlconnStringGOrdian:  os.Getenv("GO_CONN"),
		PsqlconnStringCashflow: os.Getenv("CF_CONN"),
		LoginURL:               os.Getenv("LOGIN_URL"),
		LoginToken:             os.Getenv("LOGIN_TOKEN"),
		PlsURL:                 os.Getenv("PLS_URL"),
		PlsSystem:              os.Getenv("PLS_SYSTEM"),
		ServerPort:             os.Getenv("SERVER_PORT"),
		ServerURL:              os.Getenv("SERVER_URL"),

		// local
		// PsqlconnStringGOrdian:  "host=localhost port=5432 user=alexander password=kopis dbname=budget_local sslmode=disable",
		// PsqlconnStringCashflow: "host=localhost port=54321 user=cashflow password=cashflow dbname=cashflow sslmode=disable",
		// LoginURL:               "https://login.datasektionen.se",
		// PlsURL:                 "https://pls.datasektionen.se",
		// PlsSystem:              "gordian",
		// ServerPort:             "3000",
		// ServerURL:              "http://localhost:3000",
		// LoginToken:             "this is secret fuck you",
	}

	return envConfig
}
