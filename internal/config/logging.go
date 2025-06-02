package config

// SamplingConfig задаёт параметры sample-логирования для высокой пропускной способности.
type SamplingConfig struct {
	Initial    int `mapstructure:"initial"`
	Thereafter int `mapstructure:"thereafter"`
}

// LoggingConfig описывает параметры конфигурации логгера.
type LoggingConfig struct {
	Mode              string                 `mapstructure:"mode"`
	Level             string                 `mapstructure:"level"`
	Encoding          string                 `mapstructure:"encoding"`
	Sampling          *SamplingConfig        `mapstructure:"sampling"`
	InitialFields     map[string]interface{} `mapstructure:"initialFields"`
	DisableCaller     bool                   `mapstructure:"disableCaller"`
	DisableStacktrace bool                   `mapstructure:"disableStacktrace"`
	OutputPaths       []string               `mapstructure:"outputPaths"`
	ErrorOutputPaths  []string               `mapstructure:"errorOutputPaths"`
	TimestampKey      string                 `mapstructure:"timestampKey"`
	CapitalizeLevel   bool                   `mapstructure:"capitalizeLevel"`
}
