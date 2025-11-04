package oauth2_0

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/rouroumaibing/go-devops/pkg/apis/auths"
)

func WeChatCallback(code string) string {
	WechatAppID := os.Getenv("WECHAT_APPID")
	WechatAppSecret := os.Getenv("WECHAT_APPSECRET")

	tokenURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		WechatAppID,
		WechatAppSecret,
		code,
	)

	return tokenURL
}

// getWechatAccessToken 获取微信访问令牌
func GetWechatAccessToken(tokenURL string) (*auths.WechatTokenResponse, error) {
	// 发送请求获取access_token
	resp, err := http.Get(tokenURL)
	if err != nil {
		return nil, fmt.Errorf("获取access_token失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	// 解析JSON
	var tokenResp auths.WechatTokenResponse
	if err = json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析access_token响应失败: %v", err)
	}

	// 检查是否有错误
	if tokenResp.Errcode != 0 {
		return nil, fmt.Errorf("微信接口错误(%d): %s", tokenResp.Errcode, tokenResp.Errmsg)
	}

	return &tokenResp, nil
}

// getWechatUserInfo 获取微信用户信息
func GetWechatUserInfo(accessToken, openID string) (*auths.WechatUserInfoResponse, error) {
	// 构建用户信息请求URL
	userInfoURL := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/userinfo?access_token=%s&openid=%s",
		accessToken,
		openID,
	)

	// 发送请求获取用户信息
	resp, err := http.Get(userInfoURL)
	if err != nil {
		return nil, fmt.Errorf("获取用户信息失败: %v", err)
	}
	defer resp.Body.Close()

	// 解析用户信息响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("解析用户信息响应失败: %v", err)
	}

	// 解析用户信息JSON
	var userInfo auths.WechatUserInfoResponse
	if err := json.Unmarshal(body, &userInfo); err != nil {
		return nil, fmt.Errorf("解析用户信息失败: %v", err)
	}

	// 检查是否有错误
	if userInfo.Errcode != 0 {
		return nil, fmt.Errorf("微信接口错误(%d): %s", userInfo.Errcode, userInfo.Errmsg)
	}

	return &userInfo, nil
}
