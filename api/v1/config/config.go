package v1

// Config describes the schema for configuring Status.
type Config struct {

	// Sources are a list of event sources that produce statuses.
	Sources interface{}

	// Receivers are a list of receivers onto which statuses are posted.
	Receivers interface{}

	// Settings describe the configuration for Status itself.
	Settings Settings
}
