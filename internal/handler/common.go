package handler

import (
	"context"
	"errors"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/i18n"
	"github.com/zxc7563598/bilibili-live-assistant/internal/response"
	"github.com/zxc7563598/bilibili-live-assistant/internal/validation"
	"go.uber.org/zap"
)

// AdminInfo JWT 中管理员信息
type AdminInfo struct {
	AdminID  int64  `json:"admin_id"`
	RoleID   int64  `json:"role_id"`
	RoleCode string `json:"role_code"`
}

// UserInfo JWT 中用户信息
type UserInfo struct {
	UserID int64 `json:"user_id"`
}

// GetAdminInfo 获取 JWT 携带的管理员信息
func GetAdminInfo(c *gin.Context) (AdminInfo, bool) {
	adminIDVal, ok := c.Get("adminID")
	if !ok {
		return AdminInfo{}, false
	}
	roleIDVal, ok := c.Get("roleID")
	if !ok {
		return AdminInfo{}, false
	}
	roleCodeVal, ok := c.Get("roleCode")
	if !ok {
		return AdminInfo{}, false
	}
	adminID, ok := adminIDVal.(int64)
	if !ok {
		return AdminInfo{}, false
	}
	roleID, ok := roleIDVal.(int64)
	if !ok {
		return AdminInfo{}, false
	}
	roleCode, ok := roleCodeVal.(string)
	if !ok {
		return AdminInfo{}, false
	}
	return AdminInfo{
		AdminID:  adminID,
		RoleID:   roleID,
		RoleCode: roleCode,
	}, true
}

// GetUserInfo 获取 JWT 携带的用户信息
func GetUserInfo(c *gin.Context) (UserInfo, bool) {
	userIDVal, ok := c.Get("userID")
	if !ok {
		return UserInfo{}, false
	}
	userID, ok := userIDVal.(int64)
	if !ok {
		return UserInfo{}, false
	}
	return UserInfo{
		UserID: userID,
	}, true
}

// bindAndValidate 绑定请求参数并进行验证；allowEmpty 时空请求体视为未传参直接放行
func bindAndValidate(c *gin.Context, req any, allowEmpty bool) (int, bool, error) {
	if err := c.ShouldBindJSON(req); err != nil {
		if allowEmpty && errors.Is(err, io.EOF) {
			return 0, true, nil
		}
		code := validation.Parse(err, req)
		return code, false, err
	}
	return 0, true, nil
}

// BindAndValidate 绑定请求参数并进行验证，失败将得到错误码
func BindAndValidate(c *gin.Context, req any) (int, bool, error) {
	return bindAndValidate(c, req, false)
}

// BindAndValidateAllowEmpty 绑定请求参数并进行验证；请求体为空（不含任何 JSON）时视为未传参直接放行。
// 适用于所有字段均可选的接口（如 Type *int 可选）。非空请求体仍走完整解码与校验。
func BindAndValidateAllowEmpty(c *gin.Context, req any) (int, bool, error) {
	return bindAndValidate(c, req, true)
}

// UserRequest 移动端认证接口统一前置的结果：上下文、语言与当前用户
type UserRequest struct {
	Ctx  context.Context
	Lang string
	User UserInfo
}

// BindUserRequest 移动端接口统一前置：取用户上下文 + 绑定校验请求参数。
// 任一步失败时内部已记录日志并写入错误响应，返回 ok=false。
func BindUserRequest(c *gin.Context, log *zap.Logger, action string, req any) (UserRequest, bool) {
	return bindUserRequest(c, log, action, req, false)
}

// BindUserRequestAllowEmpty 同 BindUserRequest，但允许空请求体（用于全可选字段的接口）。
func BindUserRequestAllowEmpty(c *gin.Context, log *zap.Logger, action string, req any) (UserRequest, bool) {
	return bindUserRequest(c, log, action, req, true)
}

func bindUserRequest(c *gin.Context, log *zap.Logger, action string, req any, allowEmpty bool) (UserRequest, bool) {
	ctx := c.Request.Context()
	lang := i18n.GetLang(ctx)
	user, ok := GetUserInfo(c)
	if !ok {
		response.Error(c, lang, 20001)
		return UserRequest{}, false
	}
	if code, ok, err := bindAndValidate(c, req, allowEmpty); !ok {
		ErrorLog(log, action+" 参数异常", code, err)
		response.Error(c, lang, code)
		return UserRequest{}, false
	}
	return UserRequest{Ctx: ctx, Lang: lang, User: user}, true
}

// ErrorLog 根据异常 Code 区分级别，封装异常日志信息
func ErrorLog(log *zap.Logger, msg string, code int, err error, fields ...zap.Field) {
	newFields := make([]zap.Field, 0, len(fields)+2)
	newFields = append(newFields, fields...)
	newFields = append(newFields, zap.Int("code", code))
	if err != nil {
		newFields = append(newFields, zap.Error(err))
	}
	switch {
	case code >= 50000:
		log.Error(msg, newFields...)
	case code >= 30000:
		log.Warn(msg, newFields...)
	default:
		log.Info(msg, newFields...)
	}
}
