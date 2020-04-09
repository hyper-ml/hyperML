package config

// constants associated with flow engine and Object storage

const (
	// GCS : store objects / log on google cloud
	GCS StorageTarget = "GCS"

	// S3 : store objects / log on AWS
	S3 StorageTarget = "S3"

	// DefaultStorageBucket : Default bucket for object storage
	DefaultStorageBucket = "hyperflow"

	// DefaultStorageRegion : Default region used for object storage
	DefaultStorageRegion = "us-west-2"

	// DefaultStorageTarget : Default Backend for object storage
	DefaultStorageTarget = GCS

	// DefaultStorageBaseDir : Default base directory for backend. Used by Object storage
	DefaultStorageBaseDir = "hyperflow"
)
