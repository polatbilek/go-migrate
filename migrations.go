package gomigrator

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
)

var migrationOps = make(map[string]Migration)
var genesis Migration

func Migrate(migrationTargetName string, gormDB *gorm.DB) {
	var migrationNames []string
	migrations := migrationOps
	err := validateLocalMigrationOrder(migrations, genesis)
	if err != nil {
		panic("Your local migration order is not valid. Reason: " + err.Error())
	}

	// fetch migrations
	tx := gormDB.Raw("SHOW TABLES LIKE 'migrations'").Scan(&migrationNames)
	if tx.Error != nil {
		panic("Couldn't get tables. Reason: " + tx.Error.Error())
	}

	if len(migrationNames) == 0 {
		if err := genesis.applyMigration(gormDB); err != nil {
			panic("Error while adding initial migrations to migrations table. Reason: " + err.Error())
		}
	}

	var dbMigrations []string
	err = gormDB.Table("migrations").Select("migration_name").Order("id asc").Scan(&dbMigrations).Error
	if err != nil {
		panic("Couldn't get migrations from table. Reason: " + err.Error())
	}

	err = validateLocalMigrationOrderWithDbMigrationOrder(dbMigrations, migrations, genesis)
	if err != nil {
		panic("Couldn't validate db migrations with local migration order. Reason: " + err.Error())
	}
	if migrationTargetName == "" {
		migrateAll(dbMigrations, gormDB)

	} else {
		migrateToTarget(migrationTargetName, dbMigrations, gormDB)
	}
	fmt.Println("Done.")

}

func AddMigrationOps(migrationOp Migration) {
	migrationOps[migrationOp.MigrationName] = migrationOp
}

func migrateAll(dbMigrations []string, gormDB *gorm.DB) {

	fmt.Println("Trying to migrate new changes...")
	migrations := migrationOps

	lastDbMigration := migrations[dbMigrations[len(dbMigrations)-1]]

	migrationName, err := getMigrationNameDependingOnIt(lastDbMigration, migrations)

	if err != nil {
		panic("Couldn't get migration name depending on last migration. Reason: " + err.Error())
	}

	if migrationName == "" {
		fmt.Println("Couldn't see any changes. Nothing applied.")
		return
	}

	seenLastMigration := false

	for !seenLastMigration {
		migrationName, err := getMigrationNameDependingOnIt(lastDbMigration, migrations)

		if err != nil {
			panic("Couldn't get migration name depending on last migration. Reason: " + err.Error())
		} else if migrationName == lastDbMigration.MigrationName {
			panic("Got same migration name back from getMigrationNameDependingOnIt. migrationName: " + migrationName)
		}

		if migrationName == "" {
			seenLastMigration = true
		} else {
			migration := migrations[migrationName]
			fmt.Println("Applying migration: " + migrationName + "...")
			err := migration.applyMigration(gormDB)

			if err != nil {
				panic("Couldn't apply migration (" + migrationName + "). Reason: " + err.Error())
			} else {
				fmt.Println("Applied migration: " + migrationName + " OK!\n")
			}
			lastDbMigration = migration
		}
	}
}

func migrateToTarget(migrationTargetName string, dbMigrations []string, gormDB *gorm.DB) {
	fmt.Println("Trying to migrate applied changes...")
	migrations := migrationOps
	var migrationsToRevert []Migration
	foundMigration := false

	for i := len(dbMigrations) - 1; i >= 0; i-- {
		if dbMigrations[i] == migrationTargetName {
			foundMigration = true
			break
		} else {
			migrationsToRevert = append(migrationsToRevert, migrations[dbMigrations[i]])
		}
	}

	if !foundMigration {
		fmt.Println("Couldn't find your migration among db applied migrations. migration name: " + migrationTargetName)
	} else {
		if len(migrationsToRevert) == len(dbMigrations) {
			fmt.Println("No revertable migration is found. Please check provided migration name")
		} else if len(migrationsToRevert) == 0 {
			fmt.Println("Your migration is the last applied migration. Nothing changed.")
		} else {
			for _, migration := range migrationsToRevert {
				fmt.Println("Reverting migration: " + migration.MigrationName + "...")
				err := migration.revertMigration(gormDB)

				if err != nil {
					panic("Couldn't revert migration (" + migration.MigrationName + "). Reason: " + err.Error())
				} else {
					fmt.Println("Reverted migration: " + migration.MigrationName + " OK!\n")
				}
			}
		}
	}
}

func validateLocalMigrationOrder(migrations map[string]Migration, genesis Migration) error {
	targetMigration := migrations[genesis.MigrationName]
	NotFoundLastMigration := true

	for NotFoundLastMigration {
		dependentMigrationName, err := getMigrationNameDependingOnIt(targetMigration, migrations)
		if err != nil {
			return err
		} else if dependentMigrationName == "" {
			NotFoundLastMigration = false
		} else if dependentMigrationName == targetMigration.MigrationName {
			return errors.New("Found the same target migration back. Some migrations look each other. dependentMigrationName: " + dependentMigrationName)
		} else {
			targetMigration = migrations[dependentMigrationName]
		}
	}
	return nil
}

func getMigrationNameDependingOnIt(targetMigration Migration, allMigrations map[string]Migration) (string, error) {
	var dependentMigrationName = ""

	for _, migration := range allMigrations {
		if migration.PreviousMigrationName == targetMigration.MigrationName {
			if dependentMigrationName == "" {
				dependentMigrationName = migration.MigrationName
			} else {
				return "", errors.New("More than 1 migrations are originated from same migration. origin:" + targetMigration.MigrationName)
			}
		}
	}

	return dependentMigrationName, nil
}

func validateLocalMigrationOrderWithDbMigrationOrder(dbMigrations []string, migrations map[string]Migration, genesis Migration) error {
	localMigration := genesis

	for _, dbMigrationName := range dbMigrations {
		dbMigration := migrations[dbMigrationName]
		if dbMigration.MigrationName != localMigration.MigrationName {
			return errors.New("Local migration order differs from db migration order on db migration name. db migration name: " + dbMigration.MigrationName + "local migration name: " + localMigration.MigrationName)
		} else {
			localMigrationName, err := getMigrationNameDependingOnIt(localMigration, migrations)

			if err != nil {
				return err
			}
			localMigration = migrations[localMigrationName]
		}
	}

	return nil
}

func init() {
	genesis = Migration{
		MigrationName:         "genesis",
		PreviousMigrationName: "",
		SQL: []SQLOP{
			{
				Apply: `
					CREATE TABLE migrations (
					id                INT AUTO_INCREMENT PRIMARY KEY NOT NULL,
					migration_name    VARCHAR(16) NOT NULL,
					created_at        DATETIME DEFAULT CURRENT_TIMESTAMP
					);`,
				Revert: "DROP TABLE migrations;",
			},
			{
				Apply:  "CREATE UNIQUE INDEX uidx_migration_name ON migrations(migration_name)",
				Revert: "DROP INDEX uidx_migration_name ON migrations;",
			},
		}}
	migrationOps[genesis.MigrationName] = genesis
}
