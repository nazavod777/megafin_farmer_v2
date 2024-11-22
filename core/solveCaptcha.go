package core

import (
	"encoding/json"
	"github.com/valyala/fasthttp"
	"log"
	"megafin_farmer/global"
	"strconv"
	"time"
)

type CreateTaskResult struct {
	ErrorID int `json:"errorId"`
	TaskID  int `json:"taskId"`
}

type GetTaskResultResponse struct {
	ErrorID  int    `json:"errorId"`
	Status   string `json:"status"`
	Solution struct {
		Token string `json:"token"`
	} `json:"solution"`
}

func createTask(client *fasthttp.Client,
	privateKeyHex string,
	userAgent string) string {
	payload := map[string]interface{}{
		"clientKey": global.ConfigFile.TwoCaptchaApiKey,
		"soft_id":   "4744",
		"task": map[string]interface{}{
			"soft_id":    "4744",
			"type":       "TurnstileTaskProxyless",
			"websiteURL": "https://app.megafin.xyz/",
			"websiteKey": "0x4AAAAAAA0SGzxWuGl6kriB",
			"userAgent":  userAgent,
		},
	}

	jsonData, err := json.Marshal(payload)

	if err != nil {
		log.Panicf("%s | Error Encoding Payload: %s", privateKeyHex, err)
	}

	for {
		req := fasthttp.AcquireRequest()
		defer fasthttp.ReleaseRequest(req)

		req.SetRequestURI("https://api.2captcha.com/createTask?soft_id=4744")
		req.Header.SetMethod("POST")
		req.Header.SetContentType("application/json")
		req.SetBody(jsonData)

		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(resp)

		err = client.Do(req, resp)
		if err != nil {
			log.Printf("%s | Error Sending Request When Create Task: %s", privateKeyHex, err)
			continue
		}

		body := resp.Body()

		var createTaskResponse CreateTaskResult
		if err = json.Unmarshal(body, &createTaskResponse); err != nil {
			log.Printf("%s | Error Unmarshalling Json When Create Task: %s", privateKeyHex, err)
			continue
		}

		if createTaskResponse.ErrorID != 0 {
			log.Printf("%s | Error in Response When Create Task: %s", privateKeyHex, string(body))
			continue
		}

		return strconv.Itoa(createTaskResponse.TaskID)
	}
}

func getTaskResult(client *fasthttp.Client,
	privateKeyHex string,
	taskID string) *string {
	payload := map[string]interface{}{
		"soft_id":   "4744",
		"clientKey": global.ConfigFile.TwoCaptchaApiKey,
		"taskId":    taskID,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		log.Panicf("%s | Error Encoding Payload: %s", privateKeyHex, err)
	}

	for {
		req := fasthttp.AcquireRequest()
		defer fasthttp.ReleaseRequest(req)

		req.SetRequestURI("https://api.2captcha.com/getTaskResult?soft_id=4744")
		req.Header.SetMethod("POST")
		req.Header.SetContentType("application/json")
		req.SetBody(jsonData)

		resp := fasthttp.AcquireResponse()
		defer fasthttp.ReleaseResponse(resp)

		err = client.Do(req, resp)
		if err != nil {
			log.Printf("%s | Error Sending Request When Get Task Result: %s", privateKeyHex, err)
			continue
		}

		body := resp.Body()

		var result GetTaskResultResponse
		if err = json.Unmarshal(body, &result); err != nil {
			log.Printf("%s | Error Unmarshalling Json When Get Task Result: %s", privateKeyHex, err)
			continue
		}

		if result.ErrorID != 0 {
			log.Printf("%s | Error in Response When Get Task Result: %s", privateKeyHex, string(body))
			return nil
		}

		if result.Status == "ready" {
			return &result.Solution.Token
		}

		log.Printf("%s | Captcha is still processing... Sleeping 5 secs.", privateKeyHex)
		time.Sleep(time.Second * time.Duration(5))
	}
}

func SolveCaptcha(privateKeyHex string,
	userAgent string) string {
	client := GetClient("")

	for {
		taskId := createTask(client, privateKeyHex, userAgent)
		captchaResult := getTaskResult(client, privateKeyHex, taskId)

		if captchaResult != nil {
			return *captchaResult
		}
	}
}
