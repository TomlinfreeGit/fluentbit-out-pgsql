package entity

import "time"

var tableName string

type Record struct {
	Timestamp time.Time `gorm:"column:timestamp;primaryKey"`
	Tag       string    `gorm:"column:tag;type:text"`
	Data      []byte    `gorm:"column:data;type:jsonb"`
}

func (Record) TableName() string {
	return tableName
}

func SetTableName(name string) {
	tableName = name
}
