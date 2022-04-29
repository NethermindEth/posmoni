package eth2

const (
	NoValidatorsFoundError = "no validator address or public index was found. Please check your configuration settings (file, enviroment variables, etc.)"
	NoConsensusFoundError  = "no validator address or public index was found. Please check your configuration settings (file, enviroment variables, etc.)"
	ValidatorBalancesError = "something went wrong while fetching validator balances. Skiping current checkpoint. Error: %v"
	SQLiteCreationError    = "sqlite creation failed. Error %v"
	ParseUintError         = "something went wrong while parsing uint. Skiping current validator. Error: %v"
	ValidatorNotFoundError = "validator not found. Skiping current validator. Error: %v"
	MigrationError         = "failed to migrate database. Error: %v"
	SetupError             = "an error occurred while configurating the monitor. Error: %v"
)
