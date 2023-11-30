package gomigrator

import (
	"fmt"
	"gorm.io/gorm"
)

type SQLOP struct {
	Apply  string
	Revert string
}
type Migration struct {
	MigrationName         string
	PreviousMigrationName string
	SQL                   []SQLOP
}

func (m Migration) applyMigration(gormDB *gorm.DB) error {
	if err := gormDB.Transaction(func(tx *gorm.DB) error {
		for _, statement := range m.SQL {
			if err := tx.Exec(statement.Apply).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	// adding migration name after successful transaction
	err := gormDB.Table("migrations").Create([]map[string]interface{}{{"migration_name": m.MigrationName}}).Error
	return err
}

func (m Migration) revertMigration(gormDB *gorm.DB) error {
	if err := gormDB.Transaction(func(tx *gorm.DB) error {
		// iterate backwards because the most recent execution was the last element in the list
		for i := len(m.SQL) - 1; i >= 0; i-- {
			statement := m.SQL[i]
			if err := tx.Exec(statement.Revert).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	// removing migration name after successful transaction
	err := gormDB.Exec(fmt.Sprintf("DELETE FROM `migrations` WHERE migration_name = '%s'", m.MigrationName)).Error
	return err
}
