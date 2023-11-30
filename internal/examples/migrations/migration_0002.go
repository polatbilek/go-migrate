package migrations

import "github.com/polatbilek/gomigrator/migrator"

var Migration0002 = migrator.Migration{
	MigrationName:         "migration_0002",
	PreviousMigrationName: "migration_0001",
	SQL: []migrator.SQLOP{
		{
			Apply: `
					CREATE TABLE auth_tokens (
					token      VARCHAR(64) PRIMARY KEY,
					user_id    BIGINT NOT NULL);`,
			Revert: "DROP TABLE auth_tokens;",
		},
		{
			Apply:  "CREATE UNIQUE INDEX uidx_user_id ON auth_tokens(user_id);",
			Revert: "DROP INDEX uidx_user_id ON auth_tokens;",
		},
	}}

func init() {
	migrator.AddMigrationOps(Migration0002)
}
