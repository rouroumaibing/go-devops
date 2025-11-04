package errcode

import (
	"github.com/rouroumaibing/go-devops/pkg/apis/versions"
)

func init() {
	initErrorCodes()
}

const (
	StatusFailure = "Failure"
	StatusSuccess = "Success"
)

var E struct {
	Base struct {
		AuthPWD       versions.Response
		AuthSMS       versions.Response
		Token         versions.Response
		TokenRefresh  versions.Response
		MarshalFailed versions.Response
	}
	Internal struct {
		Internal versions.Response
	}
	User struct {
		ParamError   versions.Response
		IDInvalid    versions.Response
		NotFound     versions.Response
		CreateFailed versions.Response
		UpdateFailed versions.Response
		DeleteFailed versions.Response
		QueryFailed  versions.Response
	}
	Register struct {
		RegisterFaild    versions.Response
		RegisterConflict versions.Response
	}
	Service struct {
		ParamError   versions.Response
		IDInvalid    versions.Response
		NotFound     versions.Response
		CreateFailed versions.Response
		UpdateFailed versions.Response
		DeleteFailed versions.Response
		QueryFailed  versions.Response
	}
	Pipeline struct {
		ParamError   versions.Response
		IDInvalid    versions.Response
		NotFound     versions.Response
		CreateFailed versions.Response
		UpdateFailed versions.Response
		DeleteFailed versions.Response
		QueryFailed  versions.Response
		RunFailed    versions.Response
	}
	Component struct {
		ParamError   versions.Response
		NotFound     versions.Response
		CreateFailed versions.Response
		UpdateFailed versions.Response
		DeleteFailed versions.Response
		QueryFailed  versions.Response
	}
	Environment struct {
		ParamError   versions.Response
		NotFound     versions.Response
		CreateFailed versions.Response
		UpdateFailed versions.Response
		DeleteFailed versions.Response
		QueryFailed  versions.Response
	}
	Change struct {
		ParamError   versions.Response
		IDInvalid    versions.Response
		NotFound     versions.Response
		CreateFailed versions.Response
		UpdateFailed versions.Response
		DeleteFailed versions.Response
		QueryFailed  versions.Response
	}
	Product struct {
		ParamError   versions.Response
		NotFound     versions.Response
		CreateFailed versions.Response
		UpdateFailed versions.Response
		DeleteFailed versions.Response
		QueryFailed  versions.Response
	}
}

func initErrorCodes() {
	// 基础错误码 (ERR.01系列)
	E.Base.AuthPWD = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 401, Reason: "账户或密码错误", ErrorCode: "ERR.01401001", ErrorMessage: "Authorize failed"}
	E.Base.AuthSMS = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 401, Reason: "验证失败", ErrorCode: "ERR.01401002", ErrorMessage: "Authorize failed"}
	E.Base.Token = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 401, Reason: "Token解析失败", ErrorCode: "ERR.01401003", ErrorMessage: "Authorize failed"}
	E.Base.TokenRefresh = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 401, Reason: "Token续期失败", ErrorCode: "ERR.01401004", ErrorMessage: "Authorize failed"}
	E.Base.MarshalFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 400, Reason: "JSON解析失败", ErrorCode: "ERR.01400000", ErrorMessage: "JSON parse failed"}
	E.Internal.Internal = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "内部错误", ErrorCode: "ERR.0150001", ErrorMessage: "Internal Error"}

	// 用户相关错误码 (ERR.02系列)
	E.User.ParamError = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 400, Reason: "请求参数错误", ErrorCode: "ERR.02400001", ErrorMessage: "Invalid parameter"}
	E.User.IDInvalid = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 400, Reason: "用户ID无效", ErrorCode: "ERR.02400002", ErrorMessage: "Invalid user ID"}
	E.User.NotFound = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 404, Reason: "用户不存在", ErrorCode: "ERR.02404001", ErrorMessage: "User not found"}
	E.User.CreateFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "用户创建失败", ErrorCode: "ERR.02500001", ErrorMessage: "Failed to create user"}
	E.User.UpdateFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "用户更新失败", ErrorCode: "ERR.02500002", ErrorMessage: "Failed to update user"}
	E.User.DeleteFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "用户删除失败", ErrorCode: "ERR.02500003", ErrorMessage: "Failed to delete user"}
	E.User.QueryFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "用户查询失败", ErrorCode: "ERR.02500004", ErrorMessage: "Failed to query user"}
	E.Register.RegisterFaild = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 400, Reason: "注册失败", ErrorCode: "ERR.02400003", ErrorMessage: "Register failed"}
	E.Register.RegisterConflict = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 400, Reason: "注册失败", ErrorCode: "ERR.02400004", ErrorMessage: "Register failed"}

	// 服务树相关错误码 (ERR.03系列)
	E.Service.ParamError = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 400, Reason: "请求参数错误", ErrorCode: "ERR.03400001", ErrorMessage: "Invalid parameter"}
	E.Service.IDInvalid = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 400, Reason: "服务ID无效", ErrorCode: "ERR.03400002", ErrorMessage: "Invalid service ID"}
	E.Service.NotFound = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 404, Reason: "服务不存在", ErrorCode: "ERR.03404001", ErrorMessage: "Service not found"}
	E.Service.CreateFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "服务创建失败", ErrorCode: "ERR.03500001", ErrorMessage: "Failed to create service"}
	E.Service.UpdateFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "服务更新失败", ErrorCode: "ERR.03500002", ErrorMessage: "Failed to update service"}
	E.Service.DeleteFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "服务删除失败", ErrorCode: "ERR.03500003", ErrorMessage: "Failed to delete service"}
	E.Service.QueryFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "服务查询失败", ErrorCode: "ERR.03500004", ErrorMessage: "Failed to query service"}

	// 流水线相关错误码（ERR.04系列）
	E.Pipeline.ParamError = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 400, Reason: "参数错误", ErrorCode: "ERR.04400001", ErrorMessage: "Invalid parameter"}
	E.Pipeline.IDInvalid = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 400, Reason: "流水线ID无效", ErrorCode: "ERR.04400002", ErrorMessage: "Invalid pipeline ID"}
	E.Pipeline.NotFound = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 404, Reason: "流水线不存在", ErrorCode: "ERR.04404001", ErrorMessage: "Pipeline not found"}
	E.Pipeline.CreateFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "流水线创建失败", ErrorCode: "ERR.04500001", ErrorMessage: "Failed to create pipeline"}
	E.Pipeline.UpdateFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "流水线更新失败", ErrorCode: "ERR.04500002", ErrorMessage: "Failed to update pipeline"}
	E.Pipeline.DeleteFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "流水线删除失败", ErrorCode: "ERR.04500003", ErrorMessage: "Failed to delete pipeline"}
	E.Pipeline.QueryFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "流水线查询失败", ErrorCode: "ERR.04500004", ErrorMessage: "Failed to query pipeline"}
	E.Pipeline.RunFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "流水线运行失败", ErrorCode: "ERR.04500005", ErrorMessage: "Failed to run pipeline"}
	// 组件相关错误码（ERR.05系列）
	E.Component.QueryFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "组件查询失败", ErrorCode: "ERR.05500001", ErrorMessage: "Failed to query component"}
	E.Component.ParamError = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 400, Reason: "参数错误", ErrorCode: "ERR.05400001", ErrorMessage: "Invalid parameter"}
	E.Component.NotFound = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 404, Reason: "组件不存在", ErrorCode: "ERR.05404001", ErrorMessage: "Component not found"}
	E.Component.CreateFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "组件创建失败", ErrorCode: "ERR.05500002", ErrorMessage: "Failed to create component"}
	E.Component.UpdateFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "组件更新失败", ErrorCode: "ERR.05500003", ErrorMessage: "Failed to update component"}
	E.Component.DeleteFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "组件删除失败", ErrorCode: "ERR.05500004", ErrorMessage: "Failed to delete component"}
	// 环境相关错误码（ERR.06系列）
	E.Environment.ParamError = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 400, Reason: "参数错误", ErrorCode: "ERR.06400001", ErrorMessage: "Invalid parameter"}
	E.Environment.NotFound = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 404, Reason: "未找到", ErrorCode: "ERR.06400002", ErrorMessage: "Not found"}
	E.Environment.CreateFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "环境创建失败", ErrorCode: "ERR.06400003", ErrorMessage: "Internal error"}
	E.Environment.UpdateFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "环境更新失败", ErrorCode: "ERR.06400004", ErrorMessage: "Internal error"}
	E.Environment.DeleteFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "环境删除失败", ErrorCode: "ERR.06400005", ErrorMessage: "Internal error"}
	E.Environment.QueryFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "环境查询失败", ErrorCode: "ERR.06400006", ErrorMessage: "Internal error"}

	// 变更相关错误码（ERR.07系列）
	E.Change.ParamError = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 400, Reason: "参数错误", ErrorCode: "ERR.07400001", ErrorMessage: "Invalid parameter"}
	E.Change.NotFound = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 404, Reason: "未找到", ErrorCode: "ERR.07404001", ErrorMessage: "Not found"}
	E.Change.CreateFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "变更创建失败", ErrorCode: "ERR.07500001", ErrorMessage: "Internal error"}
	E.Change.UpdateFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "变更更新失败", ErrorCode: "ERR.07500002", ErrorMessage: "Internal error"}
	E.Change.DeleteFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "变更删除失败", ErrorCode: "ERR.07500003", ErrorMessage: "Internal error"}
	E.Change.QueryFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "变更查询失败", ErrorCode: "ERR.07500004", ErrorMessage: "Internal error"}

	// 产物相关错误码（ERR.08系列）
	E.Product.ParamError = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 400, Reason: "参数错误", ErrorCode: "ERR.08400001", ErrorMessage: "Invalid parameter"}
	E.Product.NotFound = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 404, Reason: "未找到", ErrorCode: "ERR.08404001", ErrorMessage: "Not found"}
	E.Product.CreateFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "产物创建失败", ErrorCode: "ERR.08500001", ErrorMessage: "Internal error"}
	E.Product.UpdateFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "产物更新失败", ErrorCode: "ERR.08500002", ErrorMessage: "Internal error"}
	E.Product.DeleteFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "产物删除失败", ErrorCode: "ERR.08500003", ErrorMessage: "Internal error"}
	E.Product.QueryFailed = versions.Response{TypeMeta: versions.TypeMeta{Kind: versions.KindStatus, APIVersion: versions.APIVersionV1}, Status: StatusFailure, Code: 500, Reason: "产物查询失败", ErrorCode: "ERR.08500004", ErrorMessage: "Internal error"}
}
