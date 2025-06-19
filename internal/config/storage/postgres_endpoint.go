package storage

type PostgresEndpoint struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}
