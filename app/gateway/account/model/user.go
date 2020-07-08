package model

type RegisterReq struct {
	Mobile string `json:"mobile" binding:"required"`
	//Captcha  string `json:"captcha" binding:"required"`
	//AreaCode string `json:"area_code"`
}

type RegisterResp struct {
	VID   string `json:"vid"`
	Token string `json:"token"`
}

type LoginReq struct {
	Input    string `json:"input" binding:"required"`
	Password string `json:"password"`
	Captcha  string `json:"captcha"`
}

type LoginResp struct {
	VID   string `json:"vid"`
	Token string `json:"token"`
}
