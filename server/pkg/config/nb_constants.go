package config

// Constants associated with notebook processor
const (
	// NbDefaultCommand : Default command for launching notebooks
	NbDefaultCommand = "jupyter notebook --no-browser --port={port} --ip={ip}  --NotebookApp.port_retries=0 --NotebookApp.disable_check_xsrf=True" // "--NotebookApp.token=abc"

	// NbDefaultBackgroundCmd : Default command used for background notebooks
	NbDefaultBackgroundCmd = "papermill "

	// NbDefaultSweepInterval : Default Notebook scheduler sweep interval
	NbDefaultSweepInterval = 1 // minutes

	// NbDefaultSaveInterval : Default Notebook scheduler save interval
	NbDefaultSaveInterval = 60 //seconds

	// JobDefaultBackoff : Job Default bckoff
	JobDefaultBackoff = 1

	// JobDefaultDeadline : Job Default active deadline
	JobDefaultDeadline = 54000

	// DefaultNotebookPort : Default port used for hosting notebook in the container
	DefaultNotebookPort = "8888"

	// DefaultNotebookIP : Default IP used by notebook container
	DefaultNotebookIP = "0.0.0.0"

	// DefaultNotebookBasePath: Used as base path by Notebook container
	DefaultNotebookBasePath = "/"
)
