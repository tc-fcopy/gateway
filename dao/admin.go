package dao

import (
	"errors"
	"fcopy_gateway/dto"
	"fcopy_gateway/public"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"time"
)

type Admin struct {
	ID        int64     `json:"id" gorm:"primaryKey;autoIncrement;column:id;comment:自增主键"`
	UserName  string    `json:"user_name" gorm:"column:user_name;type:varchar(255);not null;default:'';comment:用户�?`
	Salt      string    `json:"salt" gorm:"column:salt;type:varchar(50);not null;default:'';comment:盐�?`
	Password  string    `json:"password" gorm:"column:password;type:varchar(255);not null;default:'';comment:加密密码"`
	CreatedAt time.Time `json:"create_at" gorm:"column:create_at;type:datetime;not null;default:'1971-01-01 00:00:00';comment:创建时间"`
	UpdatedAt time.Time `json:"update_at" gorm:"column:update_at;type:datetime;not null;default:'1971-01-01 00:00:00';comment:更新时间"`
	IsDelete  int       `json:"is_delete" gorm:"column:is_delete;type:tinyint;not null;default:0;comment:是否删除(0-正常,1-删除)"`
}

func (t *Admin) TableName() string {
	return "gateway_admin"
}

func (t *Admin) LoginCheck(c *gin.Context, tx *gorm.DB, param *dto.AdminLoginInput) (*Admin, error) {
	adminInfo, err := t.Find(c, tx, &Admin{
		UserName: param.Username,
		IsDelete: 0})
	if err != nil {
		return nil, errors.New("用户信息不存�?)
	}
	saltPassword := public.GenSaltPassword(adminInfo.Salt, param.Password)
	if adminInfo.Password != saltPassword {
		return nil, errors.New("密码错误，请重新输入")
	}
	return adminInfo, nil
}

func (t *Admin) Find(c *gin.Context, tx *gorm.DB, search *Admin) (*Admin, error) {
	out := &Admin{}
	if err := tx.WithContext(c.Request.Context()).Where(search).First(out).Error; err != nil {
		return nil, err
	}
	print(out)
	return out, nil
}
func (t *Admin) Save(c *gin.Context, tx *gorm.DB) error {
	return tx.WithContext(c.Request.Context()).Save(t).Error
}
