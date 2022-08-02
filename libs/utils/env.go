package utils

import (
	"os"
	"strconv"

	"github.com/golang/glog"
)

func GetEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func GetJwtKey() []byte {
	privKey := GetEnv("PRIV_KEY", "badKey")
	if privKey == "badKey" {
		glog.Warningf("!!! DEFAULT JWT SIGNING KEY IS BEING USED, NOT SECURE !!!")
	}
	return []byte(os.Getenv("PRIV_KEY"))
}

type SmtpConnInformation struct {
	Server   string
	Port     int
	Username string
	Password string
}

func GetSmtpConnInformation() *SmtpConnInformation {
	server := GetEnv("SMTP_SERVER", "")
	portRaw := GetEnv("SMTP_PORT", "-1")
	port, err := strconv.Atoi(portRaw)
	if err != nil {
		port = -1
	}
	username := GetEnv("SMTP_USERNAME", "")
	password := GetEnv("SMTP_PASSWORD", "")
	if server == "" || username == "" || password == "" || port == -1 {
		return nil
	}
	return &SmtpConnInformation{
		Server:   server,
		Port:     port,
		Username: username,
		Password: password,
	}
}
