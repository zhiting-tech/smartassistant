package entity

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
	"strings"
)

type Dialector struct {
	sqlite.Dialector
}

func (dialector Dialector) DataTypeOf(field *schema.Field) string {
	// 使用自定义tag，返回INTEGER PRIMARY KEY AUTOINCREMENT，确保云端sa使用的pg数据库不受影响
	if v, ok := field.TagSettings["SQLITETYPE"]; ok {
		return v
	}
	return dialector.Dialector.DataTypeOf(field)
}

func Open(dsn string) gorm.Dialector {
	return &Dialector{sqlite.Dialector{
		DSN: dsn,
	}}
}

func (dialector Dialector) Migrator(db *gorm.DB) gorm.Migrator {
	return Migrator{Migrator: sqlite.Migrator{migrator.Migrator{Config: migrator.Config{
		DB:                          db,
		Dialector:                   dialector,
		CreateIndexAfterCreateTable: true}}}}
}

type Migrator struct {
	sqlite.Migrator
}

// CreateTable
/*
gorm Migrator包的CreateTable() 方法拼接建表语句时，默认会将dataType中不含"PRIMARY KEY"的主键字段加上primary Key
如果使用gorm的自定义type标签，当sqlite类型为INTEGER PRIMARY KEY AUTOINCREMENT 时gorm无法正常建表，且会导致云端sa使用的pg数据库失去bigserial特性，
所以兼容两个数据库，重写CreateTable方法，且加入strings.Contains(strings.ToUpper(field.TagSettings["SQLITETYPE"]), "PRIMARY KEY")判断，保证
数据库运行正常
*/
func (m Migrator) CreateTable(values ...interface{}) error {
	for _, value := range m.Migrator.ReorderModels(values, false) {
		tx := m.DB.Session(&gorm.Session{})
		if err := m.RunWithValue(value, func(stmt *gorm.Statement) (errr error) {
			var (
				createTableSQL          = "CREATE TABLE ? ("
				values                  = []interface{}{m.CurrentTable(stmt)}
				hasPrimaryKeyInDataType bool
			)

			for _, dbName := range stmt.Schema.DBNames {
				field := stmt.Schema.FieldsByDBName[dbName]
				createTableSQL += "? ?"
				// strings.Contains(strings.ToUpper(field.TagSettings["SQLITETYPE"]), "PRIMARY KEY") 为true时建表语句不需要默认添加PRIMARY KEY
				hasPrimaryKeyInDataType = hasPrimaryKeyInDataType || strings.Contains(strings.ToUpper(string(field.DataType)), "PRIMARY KEY") || strings.Contains(strings.ToUpper(field.TagSettings["SQLITETYPE"]), "PRIMARY KEY")
				values = append(values, clause.Column{Name: dbName}, m.DB.Migrator().FullDataTypeOf(field))
				createTableSQL += ","
			}

			if !hasPrimaryKeyInDataType && len(stmt.Schema.PrimaryFields) > 0 {
				createTableSQL += "PRIMARY KEY ?,"
				primaryKeys := []interface{}{}
				for _, field := range stmt.Schema.PrimaryFields {
					primaryKeys = append(primaryKeys, clause.Column{Name: field.DBName})
				}

				values = append(values, primaryKeys)
			}

			for _, idx := range stmt.Schema.ParseIndexes() {
				if m.CreateIndexAfterCreateTable {
					defer func(value interface{}, name string) {
						if errr == nil {
							errr = tx.Migrator().CreateIndex(value, name)
						}
					}(value, idx.Name)
				} else {
					if idx.Class != "" {
						createTableSQL += idx.Class + " "
					}
					createTableSQL += "INDEX ? ?"

					if idx.Option != "" {
						createTableSQL += " " + idx.Option
					}

					createTableSQL += ","
					values = append(values, clause.Expr{SQL: idx.Name}, tx.Migrator().(migrator.BuildIndexOptionsInterface).BuildIndexOptions(idx.Fields, stmt))
				}
			}

			for _, rel := range stmt.Schema.Relationships.Relations {
				if !m.DB.DisableForeignKeyConstraintWhenMigrating {
					if constraint := rel.ParseConstraint(); constraint != nil {
						if constraint.Schema == stmt.Schema {
							sql, vars := buildConstraint(constraint)
							createTableSQL += sql + ","
							values = append(values, vars...)
						}
					}
				}
			}

			for _, chk := range stmt.Schema.ParseCheckConstraints() {
				createTableSQL += "CONSTRAINT ? CHECK (?),"
				values = append(values, clause.Column{Name: chk.Name}, clause.Expr{SQL: chk.Constraint})
			}

			createTableSQL = strings.TrimSuffix(createTableSQL, ",")

			createTableSQL += ")"

			if tableOption, ok := m.DB.Get("gorm:table_options"); ok {
				createTableSQL += fmt.Sprint(tableOption)
			}

			errr = tx.Exec(createTableSQL, values...).Error
			return errr
		}); err != nil {
			return err
		}
	}
	return nil
}

func buildConstraint(constraint *schema.Constraint) (sql string, results []interface{}) {
	sql = "CONSTRAINT ? FOREIGN KEY ? REFERENCES ??"
	if constraint.OnDelete != "" {
		sql += " ON DELETE " + constraint.OnDelete
	}

	if constraint.OnUpdate != "" {
		sql += " ON UPDATE " + constraint.OnUpdate
	}

	var foreignKeys, references []interface{}
	for _, field := range constraint.ForeignKeys {
		foreignKeys = append(foreignKeys, clause.Column{Name: field.DBName})
	}

	for _, field := range constraint.References {
		references = append(references, clause.Column{Name: field.DBName})
	}
	results = append(results, clause.Table{Name: constraint.Name}, foreignKeys, clause.Table{Name: constraint.ReferenceSchema.Table}, references)
	return
}
