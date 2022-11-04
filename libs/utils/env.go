package utils

import (
	"os"
	"strconv"
	"strings"

	"k8s.io/klog/v2"
)

func GetEnv(key string, fallback string) string {
	value := os.Getenv(key)
	if len(value) == 0 {
		return fallback
	}
	return value
}

func GetBannedRewards() []string {
	raw := GetEnv("BPOW_BANNED_REWARDS", "")
	return strings.Split(raw, ",")
}

func GetAllowedEmails() []string {
	raw := GetEnv("BPOW_ALLOWED_EMAILS", "")
	return strings.Split(raw, ",")
}

func GetServiceTokens() []string {
	raw := GetEnv("BPOW_SERVICE_TOKENS", "")
	return strings.Split(raw, ",")
}

func GetJwtKey() []byte {
	privKey := GetEnv("PRIV_KEY", "badKey")
	if privKey == "badKey" {
		klog.Warningf("!!! DEFAULT JWT SIGNING KEY IS BEING USED, NOT SECURE !!!")
	}
	return []byte(privKey)
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

func GetTotalPrizePool() int {
	totalPrizeRaw := GetEnv("BPOW_PRIZE_POOL", "10000")
	totalPrize, err := strconv.Atoi(totalPrizeRaw)
	if err != nil {
		totalPrize = 0
	}
	return totalPrize
}

func GetWalletID() string {
	return GetEnv("BPOW_WALLET_ID", "wallet_id_not_set")
}

func GetWalletAddress() string {
	return GetEnv("BPOW_WALLET_ADDRESS", "wallet_address_not_set")
}
