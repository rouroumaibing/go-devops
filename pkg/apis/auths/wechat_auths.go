package auths

// 微信登录响应
type WechatTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenID       string `json:"openid"`
	Scope        string `json:"scope"`
	UnionID      string `json:"unionid,omitempty"`
	Errcode      int    `json:"errcode,omitempty"`
	Errmsg       string `json:"errmsg,omitempty"`
}

// 用户信息响应
type WechatUserInfoResponse struct {
	OpenID    string   `json:"openid"`
	Nickname  string   `json:"nickname"`
	HeadImg   string   `json:"headimg"`
	UnionID   string   `json:"unionid,omitempty"`
	Privilege []string `json:"privilege"`
	Errcode   int      `json:"errcode,omitempty"`
	Errmsg    string   `json:"errmsg,omitempty"`
}

// 用户会话信息
type WechatUserSessionInfo struct {
	OpenID     string `json:"openid"`
	Nickname   string `json:"nickname"`
	HeadImg    string `json:"headimg"`
	UnionID    string `json:"unionid"`
	IsLoggedIn bool   `json:"isLoggedIn"`
}
