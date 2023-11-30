package migrations

import (
	"github.com/polatbilek/gomigrator"
)

var Migration0001 = gomigrator.Migration{
	MigrationName:         "migration_0001",
	PreviousMigrationName: "migration_0000",
	SQL: []gomigrator.SQLOP{
		{
			Apply: `
					CREATE TABLE users (
					id            BIGINT AUTO_INCREMENT PRIMARY KEY,
					name          VARCHAR(128),
					surname       VARCHAR(128),
					nickname      VARCHAR(32) NOT NULL,
					created_at    DATETIME DEFAULT CURRENT_TIMESTAMP,
					public_id     VARCHAR(32) NOT NULL,
					password      VARCHAR(64));`,
			Revert: "DROP TABLE users;",
		},
		{
			Apply:  "CREATE UNIQUE INDEX uidx_public_id ON users(public_id);",
			Revert: "DROP INDEX uidx_public_id ON users;",
		},
		{
			Apply:  "CREATE UNIQUE INDEX uidx_nickname ON users(nickname);",
			Revert: "DROP INDEX uidx_nickname ON users;",
		},
	}}

func init() {
	gomigrator.AddMigrationOps(Migration0001)
}
