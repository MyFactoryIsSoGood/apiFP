package db

import (
	"github.com/jinzhu/gorm"
)

type ISOTemplatesSTR struct {
	gorm.Model
	TemplateData string `gorm:"type:bytea"`
	Meta         string `gorm:"type:varchar(255)"`
}

type ISOTemplates struct {
	gorm.Model
	TemplateData []byte `gorm:"type:bytea"`
}

func Push(template []byte, fname string) {
	DB.Create(&ISOTemplatesSTR{TemplateData: string(template), Meta: fname})
	var newTemplate ISOTemplatesSTR
	DB.Last(&newTemplate)
}

func Take(id int) ([]byte, string) {
	var newTemplate ISOTemplatesSTR
	DB.Where("id = ?", id).First(&newTemplate)

	return []byte(newTemplate.TemplateData), newTemplate.Meta
}
