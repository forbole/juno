package database

// Migrator represents a generic database migrator that should be used to migrate from one database version to another
type Migrator interface {
	// Migrate performs the migrations and returns any error
	Migrate() error
}
