package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/zxc7563598/bilibili-live-assistant/internal/validation"
	"go.uber.org/zap"
)

// AdminInfo JWT 中管理员信息
type AdminInfo struct {
	AdminID  uint64 `json:"admin_id"`
	RoleID   uint64 `json:"role_id"`
	RoleCode string `json:"role_code"`
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
	adminID, ok := adminIDVal.(uint64)
	if !ok {
		return AdminInfo{}, false
	}
	roleID, ok := roleIDVal.(uint64)
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

// BindAndValidate 绑定请求参数并进行验证，失败将得到错误码
func BindAndValidate(c *gin.Context, req any) (int, bool, error) {
	if err := c.ShouldBindJSON(req); err != nil {
		code := validation.Parse(err, req)
		return code, false, err
	}
	return 0, true, nil
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
