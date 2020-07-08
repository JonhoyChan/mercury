package orm

import (
	. "github.com/jinzhu/gorm"
)

func init() {
	DefaultCallback.Create().Replace("gorm:update_time_stamp", updateTimeStampForCreateCallback)
	DefaultCallback.Update().Replace("gorm:update_time_stamp", updateTimeStampForUpdateCallback)
}

// updateTimeStampForCreateCallback will set `create_at`, `update_at` when creating
func updateTimeStampForCreateCallback(scope *Scope) {
	if !scope.HasError() {
		now := NowFunc().Unix()

		if createdAtField, ok := scope.FieldByName("create_at"); ok {
			if createdAtField.IsBlank {
				createdAtField.Set(now)
			}
		}

		if updatedAtField, ok := scope.FieldByName("update_at"); ok {
			if updatedAtField.IsBlank {
				updatedAtField.Set(now)
			}
		}
	}
}

func updateTimeStampForUpdateCallback(scope *Scope) {
	if _, ok := scope.Get("gorm:update_column"); !ok {
		scope.SetColumn("update_at", NowFunc().Unix())
	}
}
