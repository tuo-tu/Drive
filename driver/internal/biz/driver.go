package biz

import (
	"context"
	"database/sql"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
	"regexp"
	"strings"
	"time"
)

const SecretKey = "driver secret key"

// 司机业务逻辑
type DriverBiz struct {
	DI DriverInterface
}

// 司机表模型
type Driver struct {
	// 基础模型
	gorm.Model
	// 业务模型
	DriverWork
	// 关联部分
}

// 司机的业务模型
type DriverWork struct {
	Telephone     string         `gorm:"type:varchar(16);uniqueIndex;" json:"telephone"`
	Token         sql.NullString `gorm:"type:varchar(2047);" json:"token"`
	Status        sql.NullString `gorm:"type:enum('out', 'in', 'listen', 'stop');" json:"status"`
	Name          sql.NullString `gorm:"type:varchar(255);index;" json:"name"`
	IdNumber      sql.NullString `gorm:"type:char(18);uniqueIndex;" json:"id_number"`
	IdImageA      sql.NullString `gorm:"type:varchar(255);" json:"id_image_a"`
	LicenseImageA sql.NullString `gorm:"type:varchar(255);" json:"license_image_a"`
	LicenseImageB sql.NullString `gorm:"type:varchar(255);" json:"license_image_b"`
	DistinctCode  sql.NullString `gorm:"type:varchar(16);index;" json:"distinct_code"`
	TelephoneBak  sql.NullString `gorm:"type:varchar(16);index;" json:"telephone_bak"`
	AuditAt       sql.NullTime   `gorm:"index;" json:"audit_at"`
}

// 司机状态常量
const DriverStatusOut = "out"
const DriverStatusIn = "in"
const DriverStatusListen = "listen"
const DriverStatusStop = "stop"
const DriverTokenLife = 1 * 30 * 24 * 3600

// 司机相关的资源操作接口
type DriverInterface interface {
	GetVerifyCode(context.Context, string) (string, error)
	FetchVerifyCode(context.Context, string) (string, error)
	FetchInfoByTel(context.Context, string) (*Driver, error)
	InitDriverInfo(context.Context, string) (*Driver, error)
	GetSavedVerifyCode(context.Context, string) (string, error)
	SaveToken(context.Context, string, string) error
	GetToken(context.Context, string) (string, error)
}

// DriverBiz 构造器
func NewDriverBiz(di DriverInterface) *DriverBiz {
	return &DriverBiz{
		DI: di,
	}
}

// 验证登录信息方法
func (db *DriverBiz) CheckLogin(ctx context.Context, tel, verifyCode string) (string, error) {
	// 验证验证码是否正确
	code, err := db.DI.GetSavedVerifyCode(ctx, tel)
	if err != nil {
		return "", err
	}
	if verifyCode != code {
		return "", errors.New(1, "verify code error", "'")
	}
	// 生成token
	token, err := generateJWT(tel)
	if err != nil {
		return "", err
	}
	// 存储到driver表中
	if err := db.DI.SaveToken(ctx, tel, token); err != nil {
		return "", err
	}
	// 返回token
	return token, nil
}

// 生成JWT token
func generateJWT(tel string) (string, error) {
	// 构建token类型
	claims := jwt.RegisteredClaims{
		Issuer:    "LaomaDJ",
		Subject:   "driver authentication",
		Audience:  []string{"driver"},
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(DriverTokenLife * time.Second)),
		NotBefore: nil,
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ID:        tel,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 签名
	// 生成token字符串
	tokenString, err := token.SignedString([]byte(SecretKey))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// 比对验证码是否一致
// 将司机信息入库
func (db *DriverBiz) InitDriverInfo(ctx context.Context, tel string) (*Driver, error) {
	// 校验验证码（略）
	// 司机是否已经注册的校验（略）？
	// 司机是否在黑名单中校验（略）？
	if tel == "" {
		return nil, errors.New(1, "telephone is empty", "")
	}

	return db.DI.InitDriverInfo(ctx, tel)
}

// 实现获取验证码的业务逻辑
func (db *DriverBiz) GetVerifyCode(ctx context.Context, tel string) (string, error) {
	// 一，校验手机号
	pattern := `^(13\d|14[01456879]|15[0-35-9]|16[2567]|17[0-8]|18\d|19[0-35-9])\d{8}$`
	// 返回一个编译后的正则表达式对象
	regexpPattern := regexp.MustCompile(pattern)
	if !regexpPattern.MatchString(tel) {
		return "", errors.New(200, "DRIVER", "driver telephone error")
	}
	// 二，调用data/获取验证码
	return db.DI.GetVerifyCode(ctx, tel)
}

func (db *DriverBiz) CheckVerifyCode(ctx context.Context, tel, code string) bool {
	// 一，校验手机号
	pattern := `^(13\d|14[01456879]|15[0-35-9]|16[2567]|17[0-8]|18\d|19[0-35-9])\d{8}$`
	regexpPattern := regexp.MustCompile(pattern)
	if !regexpPattern.MatchString(tel) {
		return false
	}

	code = strings.TrimSpace(code)
	if len(code) == 0 {
		return false
	}

	verifyCode, err := db.DI.FetchVerifyCode(ctx, tel)
	if err != nil {
		return false
	}
	if verifyCode == code {
		return true
	}

	return false
}

func (db *DriverBiz) GetInfoByTel(ctx context.Context, tel string) (*Driver, error) {
	driver, err := db.DI.FetchInfoByTel(ctx, tel)
	if err != nil {
		return nil, err
	}
	return driver, nil
}
