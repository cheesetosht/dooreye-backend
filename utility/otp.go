package utility

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"log"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

const (
	otpDigits    = 4
	otpPeriod    = 300 // 5 minutes
	otpAlgorithm = "SHA1"
)

func GenerateVerificationCode() (string, string, error) {
	// Generate a random secret
	secret := make([]byte, 20)
	_, err := rand.Read(secret)
	if err != nil {
		return "", "", err
	}
	secretBase32 := base32.StdEncoding.EncodeToString(secret)

	// Generate OTP
	otp, err := totp.GenerateCodeCustom(secretBase32, time.Now(), totp.ValidateOpts{
		Digits:    otpDigits,
		Algorithm: otp.AlgorithmSHA1,
		Period:    otpPeriod,
	})
	if err != nil {
		log.Println("!! failed to generate otp: ", err)
		return "", "", err
	}

	return otp, secretBase32, nil
}

func VerifyOTP(otpString, secret string) (bool, error) {
	isValid, err := totp.ValidateCustom(otpString, secret, time.Now(), totp.ValidateOpts{
		Digits:    otpDigits,
		Algorithm: otp.AlgorithmSHA1,
		Period:    otpPeriod,
	})
	if err != nil {
		return false, fmt.Errorf("!! failed to validate otp: %w", err)
	}
	return isValid, nil
}
