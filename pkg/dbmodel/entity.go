package dbmodel

import (
	"github.com/sentrycloud/sentry/pkg/newlog"
	"reflect"
	"time"
)

// Entity all MySQL tables should include these columns, embed this structure to all table entity
type Entity struct {
	ID        uint32    `json:"id"`
	Created   time.Time `json:"created"`
	Updated   time.Time `json:"updated"`
	IsDeleted int       `json:"is_deleted"`
}

func (p *Entity) SetTimeNow() {
	now := time.Now()
	p.Created = now
	p.Updated = now
}

func getJsonTags(entity interface{}) []string {
	var fields []string
	entityType := reflect.TypeOf(entity)
	if entityType.Kind() != reflect.Ptr {
		return fields
	}

	objectType := entityType.Elem()
	if objectType.Kind() != reflect.Struct {
		return fields
	}
	numField := objectType.NumField()
	if numField > 1 {
		// skip the first Entity field, that is an empty string
		for i := 1; i < numField; i++ {
			field := objectType.Field(i)
			fields = append(fields, field.Tag.Get("json"))
		}
	}

	return fields
}

func QueryAllEntity(entities interface{}) error {
	result := db.Where("is_deleted = ?", 0).Find(entities)
	if result.Error != nil {
		newlog.Error("query alarm contacts failed: %v", result.Error)
		return result.Error
	}

	return nil
}

func AddEntity(entity interface{}) error {
	fields := getJsonTags(entity)
	result := db.Select(fields).Create(entity)
	return result.Error
}

func UpdateEntity(entity interface{}) error {
	fields := getJsonTags(entity)
	result := db.Model(entity).Select(fields).Updates(entity)
	return result.Error
}

func DeleteEntity(entity interface{}) error {
	// soft delete
	result := db.Model(entity).Update("is_deleted", 1)
	return result.Error
}
