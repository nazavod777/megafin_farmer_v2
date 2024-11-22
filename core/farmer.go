package core

import (
	"encoding/json"
	"fmt"
	"github.com/corpix/uarand"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/valyala/fasthttp"
	"log"
	"megafin_farmer/customTypes"
	"megafin_farmer/global"
	"strings"
	"time"
)

func doRequest(client *fasthttp.Client,
	url string,
	method string,
	payload interface{},
	headers map[string]string) ([]byte, int, error) {

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod(strings.ToUpper(method))
	req.SetRequestURI(url)
	req.Header.SetContentType("application/json")

	if payload != nil {
		jsonData, err := json.Marshal(payload)

		if err != nil {
			return nil, 0, fmt.Errorf("failed to marshal JSON: %w", err)
		}
		req.SetBody(jsonData)
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)

	if err := client.Do(req, resp); err != nil {
		return nil, 0, fmt.Errorf("request failed: %w", err)
	}

	respBody := make([]byte, len(resp.Body()))
	copy(respBody, resp.Body())

	return respBody, resp.StatusCode(), nil
}

func profileRequest(client *fasthttp.Client,
	privateKeyHex string,
	headers map[string]string) (int, map[string]string, float64, float64) {
	for {
		var responseData customTypes.ProfileResponseStruct

		respBody, statusCode, err := doRequest(client, "https://api.megafin.xyz/users/profile", "GET", nil, headers)

		if statusCode == 401 {
			delete(headers, "X-Recaptcha-Response")
			return 401, nil, 0, 0
		}

		if err != nil {
			log.Printf("%s | Error When Profile: %s | Status Code: %d", privateKeyHex, err, statusCode)
			headers["user-agent"] = uarand.GetRandom()
			continue
		}

		if strings.Contains(string(respBody), "title>Access denied | api.megafin.xyz used Cloudflare to restrict access</title>") || strings.Contains(string(respBody), "<title>Just a moment...</title>") || strings.Contains(string(respBody), "<title>Attention Required! ") {
			log.Printf("%s | CloudFlare", privateKeyHex)
			headers["user-agent"] = uarand.GetRandom()
			continue
		}

		if err = json.Unmarshal(respBody, &responseData); err != nil {
			log.Printf("%s | Failed To Parse JSON Response When Profile: %s | Status Code: %d", privateKeyHex, string(respBody), statusCode)
			headers["user-agent"] = uarand.GetRandom()
			continue
		}

		return 0, headers, responseData.Result.Balance.MGF, responseData.Result.Balance.USDC
	}
}

func loginAccount(client *fasthttp.Client,
	privateKeyHex string,
	headers map[string]string) (map[string]string, string) {
	captchaResponse := SolveCaptcha(privateKeyHex, headers["user-agent"])

	headers["X-Recaptcha-Response"] = captchaResponse
	headers["accept"] = "application/json"

	privateKey, err := crypto.HexToECDSA(privateKeyHex)

	if err != nil {
		log.Panicf("%s | Failed to parse private key: %v", privateKeyHex, err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	signText := fmt.Sprintf("megafin.xyz requests you to sign in with your wallet address: %s", address.Hex())
	data := accounts.TextHash([]byte(signText))
	signature, err := crypto.Sign(data, privateKey)

	if err != nil {
		log.Panicf("%s | Failed to sign message: %v", privateKeyHex, err)
	}

	signature[64] += 27
	signHash := fmt.Sprintf("0x%x", signature)

	payload := map[string]interface{}{
		"invite_code": "133d76e4",
		"key":         address.String(),
		"wallet_hash": signHash,
	}

	for {
		var responseData customTypes.LoginResponseStruct
		respBody, statusCode, err := doRequest(client, "https://api.megafin.xyz/auth", "POST", payload, headers)

		if err != nil {
			log.Printf("%s | Error When Auth: %s | Status Code: %d", privateKeyHex, err, statusCode)
			headers["user-agent"] = uarand.GetRandom()
			continue
		}

		if strings.Contains(string(respBody), "title>Access denied | api.megafin.xyz used Cloudflare to restrict access</title>") || strings.Contains(string(respBody), "<title>Just a moment...</title>") || strings.Contains(string(respBody), "<title>Attention Required! ") {
			log.Printf("%s | CloudFlare", privateKeyHex)
			headers["user-agent"] = uarand.GetRandom()
			continue
		}

		if err = json.Unmarshal(respBody, &responseData); err != nil {
			log.Printf("%s | Failed To Parse JSON Response When Logging: %s | Status Code: %d", privateKeyHex, string(respBody), statusCode)
			headers["user-agent"] = uarand.GetRandom()
			continue
		}

		if responseData.Result.Token == "" {
			log.Printf("%s | Wrong Response When Auth: %s | Status Code: %d", privateKeyHex, err, statusCode)
			continue
		}

		delete(headers, "X-Recaptcha-Response")
		return headers, responseData.Result.Token
	}
}

func sendConnectRequest(client *fasthttp.Client,
	privateKeyHex string,
	headers map[string]string) (int, map[string]string, float64, float64) {
	for {
		var responseData customTypes.PingResponseStruct

		respBody, statusCode, err := doRequest(client, "https://api.megafin.xyz/users/connect", "GET", nil, headers)

		if statusCode == 401 {
			delete(headers, "X-Recaptcha-Response")
			return 401, headers, 0, 0
		}

		if err != nil {
			log.Printf("%s | Error When Pinging: %s | Status Code: %d", privateKeyHex, err, statusCode)
			headers["user-agent"] = uarand.GetRandom()
			continue
		}

		if strings.Contains(string(respBody), "title>Access denied | api.megafin.xyz used Cloudflare to restrict access</title>") || strings.Contains(string(respBody), "<title>Just a moment...</title>") || strings.Contains(string(respBody), "<title>Attention Required! ") {
			log.Printf("%s | CloudFlare", privateKeyHex)
			headers["user-agent"] = uarand.GetRandom()
			continue
		}

		if err = json.Unmarshal(respBody, &responseData); err != nil {
			log.Printf("%s | Failed To Parse JSON Response When Pinging: %s | Status Code: %d", privateKeyHex, string(respBody), statusCode)
			headers["user-agent"] = uarand.GetRandom()
			continue
		}

		return 0, headers, responseData.Result.Balance.MGF, responseData.Result.Balance.USDC
	}
}

func StartFarmAccount(privateKey string,
	proxy string) {
	headers := map[string]string{
		"accept":          "*/*",
		"accept-language": "ru,en;q=0.9,vi;q=0.8,es;q=0.7,cy;q=0.6",
		"origin":          "https://app.megafin.xyz",
		"referer":         "https://app.megafin.xyz",
		"connection":      "close",
		"user-agent":      uarand.GetRandom(),
	}

	client := GetClient(proxy)

	for {
		var authToken string

		global.Semaphore <- struct{}{}
		for {
			headers, authToken = loginAccount(client, privateKey, headers)
			headers["Authorization"] = "Bearer " + authToken
			statusCode, _, _, _ := profileRequest(client, privateKey, headers)

			if statusCode == 401 {
				log.Printf("%s | Unathorized, Relogining...", privateKey)
				continue
			}

			break
		}
		<-global.Semaphore

		for {
			var statusCode int
			var mgfBalance, usdcBalance float64
			statusCode, headers, mgfBalance, usdcBalance = sendConnectRequest(client, privateKey, headers)

			if statusCode == 401 {
				log.Printf("%s | Unathorized, Relogining...", privateKey)
				break
			}

			log.Printf("%s | MGF Balance: %f | USDC Balance: %f | Sleeping 120 secs.", privateKey, mgfBalance, usdcBalance)
			time.Sleep(time.Second * time.Duration(120))
		}
	}
}

func ParseAccountBalance(privateKey string,
	proxy string) (float64, float64) {
	headers := map[string]string{
		"accept":          "*/*",
		"accept-language": "ru,en;q=0.9,vi;q=0.8,es;q=0.7,cy;q=0.6",
		"origin":          "https://app.megafin.xyz",
		"referer":         "https://app.megafin.xyz",
		"connection":      "close",
		"user-agent":      uarand.GetRandom(),
	}

	client := GetClient(proxy)

	var mgfBalance, usdcBalance float64
	var authToken string
	var statusCode int

	global.Semaphore <- struct{}{}
	for {
		headers, authToken = loginAccount(client, privateKey, headers)
		headers["Authorization"] = "Bearer " + authToken

		statusCode, headers, mgfBalance, usdcBalance = profileRequest(client, privateKey, headers)

		if statusCode == 401 {
			log.Printf("%s | Unathorized, Relogining...", privateKey)
			continue
		}

		break
	}
	<-global.Semaphore

	log.Printf("%s | MGF Balance: %f | USDC Balance: %f", privateKey, mgfBalance, usdcBalance)

	return mgfBalance, usdcBalance
}
