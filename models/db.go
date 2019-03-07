package models

import (
	"database/sql"
	"fmt"
	"github.com/EDDYCJY/go-gin-example/pkg/setting"
	"github.com/jinzhu/gorm"
	"log"
	"time"
)

//how to stub these without modify code in this file?
var Db IDB
var Db2 = Open()

func init() {
	Db = Open()
}

type IDB interface {
	SingularTable(enable bool)
	Callback() *gorm.Callback
	DB() *sql.DB
	Select(query interface{}, args ...interface{}) IDB
	Where(query interface{}, args ...interface{}) IDB
	First(out interface{}, where ...interface{}) IDB
	Count(value interface{}) IDB
	Model(value interface{}) IDB
	Preload(column string, conditions ...interface{}) IDB
	Offset(offset interface{}) IDB
	Limit(limit interface{}) IDB
	Find(out interface{}, where ...interface{}) IDB
	Related(value interface{}, foreignKeys ...string) IDB
	Updates(values interface{}, ignoreProtectedAttrs ...bool) IDB
	Create(value interface{}) IDB
	Delete(value interface{}, where ...interface{}) IDB
	Connected() bool
	Unscoped() IDB
	Close() error
	Error() error
}

type dbWrap struct {
	db *gorm.DB
}

func newDbWrap(gorm *gorm.DB) *dbWrap {
	return &dbWrap{
		db: gorm,
	}
}

func (w *dbWrap) exec(action func(db *gorm.DB) *gorm.DB) *dbWrap {
	return newDbWrap(action(w.db))
}

func (w *dbWrap) Connected() bool {
	return w.db.DB().Stats().OpenConnections > 0
}

func (w *dbWrap) Close() error {
	return w.db.Close()
}

func (w *dbWrap) Unscoped() IDB {
	return w.exec(func(db *gorm.DB) *gorm.DB {
		return db.Unscoped()
	})
}

func (w *dbWrap) Delete(value interface{}, where ...interface{}) IDB {
	return w.exec(func(db *gorm.DB) *gorm.DB {
		return db.Delete(value, where...)
	})
}

func (w *dbWrap) Create(value interface{}) IDB {
	return w.exec(func(db *gorm.DB) *gorm.DB {
		return db.Create(value)
	})
}

func (w *dbWrap) Updates(values interface{}, ignoreProtectedAttrs ...bool) IDB {
	return w.exec(func(db *gorm.DB) *gorm.DB {
		return db.Updates(values, ignoreProtectedAttrs...)
	})
}

func (w *dbWrap) Related(value interface{}, foreignKeys ...string) IDB {
	return w.exec(func(db *gorm.DB) *gorm.DB {
		return db.Related(value, foreignKeys...)
	})
}

func (w *dbWrap) Find(out interface{}, where ...interface{}) IDB {
	return w.exec(func(db *gorm.DB) *gorm.DB {
		return db.First(out, where...)
	})
}

func (w *dbWrap) Limit(limit interface{}) IDB {
	return w.exec(func(db *gorm.DB) *gorm.DB {
		return db.Limit(limit)
	})
}

func (w *dbWrap) Offset(offset interface{}) IDB {
	return w.exec(func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset)
	})
}

func (w *dbWrap) Preload(column string, conditions ...interface{}) IDB {
	return newDbWrap(w.db.Preload(column, conditions...))
}

func (w *dbWrap) Model(value interface{}) IDB {
	return newDbWrap(w.db.Model(value))
}

func (w *dbWrap) Count(value interface{}) IDB {
	return newDbWrap(w.db.Count(value))
}

func (w *dbWrap) Error() error {
	return w.db.Error
}

func (w *dbWrap) First(out interface{}, where ...interface{}) IDB {
	return newDbWrap(w.db.First(out, where...))
}

func (w *dbWrap) Where(query interface{}, args ...interface{}) IDB {
	return newDbWrap(w.db.Where(query, args...))
}

func (w *dbWrap) Select(query interface{}, args ...interface{}) IDB {
	return newDbWrap(w.db.Select(query, args...))
}

func (w *dbWrap) SingularTable(enable bool) {
	w.db.SingularTable(enable)
}

func (w *dbWrap) Callback() *gorm.Callback {
	return w.db.Callback()
}

func (w *dbWrap) DB() *sql.DB {
	return w.db.DB()
}

func Open() IDB {
	gormDb, err := gorm.Open(setting.DatabaseSetting.Type, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local",
		setting.DatabaseSetting.User,
		setting.DatabaseSetting.Password,
		setting.DatabaseSetting.Host,
		setting.DatabaseSetting.Name))

	if err != nil {
		log.Fatalf("models.Open err: %v", err)
	}

	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return setting.DatabaseSetting.TablePrefix + defaultTableName
	}

	db := newDbWrap(gormDb)

	db.SingularTable(true)
	db.Callback().Create().Replace("gorm:update_time_stamp", updateTimeStampForCreateCallback)
	db.Callback().Update().Replace("gorm:update_time_stamp", updateTimeStampForUpdateCallback)
	db.Callback().Delete().Replace("gorm:delete", deleteCallback)
	db.DB().SetMaxIdleConns(10)
	db.DB().SetMaxOpenConns(100)
	return db
}

// updateTimeStampForCreateCallback will set `CreatedOn`, `ModifiedOn` when creating
func updateTimeStampForCreateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		nowTime := time.Now().Unix()
		if createTimeField, ok := scope.FieldByName("CreatedOn"); ok {
			if createTimeField.IsBlank {
				createTimeField.Set(nowTime)
			}
		}

		if modifyTimeField, ok := scope.FieldByName("ModifiedOn"); ok {
			if modifyTimeField.IsBlank {
				modifyTimeField.Set(nowTime)
			}
		}
	}
}

// updateTimeStampForUpdateCallback will set `ModifiedOn` when updating
func updateTimeStampForUpdateCallback(scope *gorm.Scope) {
	if _, ok := scope.Get("gorm:update_column"); !ok {
		scope.SetColumn("ModifiedOn", time.Now().Unix())
	}
}

func deleteCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		var extraOption string
		if str, ok := scope.Get("gorm:delete_option"); ok {
			extraOption = fmt.Sprint(str)
		}

		deletedOnField, hasDeletedOnField := scope.FieldByName("DeletedOn")

		if !scope.Search.Unscoped && hasDeletedOnField {
			scope.Raw(fmt.Sprintf(
				"UPDATE %v SET %v=%v%v%v",
				scope.QuotedTableName(),
				scope.Quote(deletedOnField.DBName),
				scope.AddToVars(time.Now().Unix()),
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				addExtraSpaceIfExist(extraOption),
			)).Exec()
		} else {
			scope.Raw(fmt.Sprintf(
				"DELETE FROM %v%v%v",
				scope.QuotedTableName(),
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				addExtraSpaceIfExist(extraOption),
			)).Exec()
		}
	}
}

func addExtraSpaceIfExist(str string) string {
	if str != "" {
		return " " + str
	}
	return ""
}
