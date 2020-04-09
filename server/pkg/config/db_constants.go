package config

// constants associated with DB

const (
	// Bolt : Database Option
	Bolt DbTarget = "BOLT"

	// Postgres : Database option
	Postgres DbTarget = "POSTGRES"

	// Badger : DB Option
	Badger DbTarget = "BADGER"

	// DefaultDBEventBuffer : Change Listern Buffer length
	DefaultDBEventBuffer = 500

	// DefaultDataDirPath : Used by Bolt/Badger as data file path
	DefaultDataDirPath = "/data/badger"

	// DefaultDBBucket : Used by Bolt/Badger to name the bucket
	DefaultDBBucket = "hflow_master"
)
