package customTypes

type LoginResponseStruct struct {
	Result struct {
		Address string `json:"address"`
		Token   string `json:"token"`
	} `json:"result"`
}

type ProfileResponseStruct struct {
	Result struct {
		Address    string `json:"address"`
		InviteCode string `json:"invite_code"`
		Balance    struct {
			MGF  float64 `json:"MGF"`
			USDC float64 `json:"USDC"`
		} `json:"balance"`
		NFTConfig struct {
			BuffSpeed float64 `json:"buff_speed"`
			Quantity  struct {
				Basic int `json:"basic"`
			} `json:"quantity"`
			Speed struct {
				MGF  float64 `json:"MGF"`
				USDC float64 `json:"USDC"`
			} `json:"speed"`
		} `json:"nft_config"`
	} `json:"result"`
}

type PingResponseStruct struct {
	Result struct {
		Balance struct {
			MGF  float64 `json:"MGF"`
			USDC float64 `json:"USDC"`
		} `json:"balance"`
	} `json:"result"`
}

type ConfigStruct struct {
	TwoCaptchaApiKey string `json:"two_captcha_api_key"`
}
