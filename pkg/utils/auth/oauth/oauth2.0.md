# OAuth2.0相关开发记录

### wechat

微信公众平台接口测试帐号申请

无需公众帐号、快速申请接口测试号

直接体验和测试公众平台所有高级接口

https://mp.weixin.qq.com/debug/cgi-bin/sandbox?t=sandbox/login


步骤1：扫码登录

步骤2：登录成功后微信会分配appID和appsecret

步骤3：URL填后端服务接口的访问URL，Token可以随便填

步骤4：微信会去访问我们的后端服务，访问地址就是URL。

请求会传四个参数，其中关键的是echostr参数，我们必须原样返回。

| 参数      | 描述                                                      |
| --------- | --------------------------------------------------------- |
| signature | 微信加密签名。结合token参数、timestamp时间戳、nonce参数。 |
| timestamp | 时间戳                                                    |
| nonce     | 随机数                                                    |
| echostr   | 随机字符串                                                |
