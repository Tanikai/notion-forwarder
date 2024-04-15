package models

type ForwardedDatabaseConfig struct {
	Name              string `koanf:"name" validate:"required"`
	DatabaseId        string `koanf:"database_id" validate:"required"`
	ForwardColumnName string `koanf:"forward_column_name" validate:"required"`
}

type NotionForwarderConfig struct {
	LogLevel         string                    `koanf:"log_level"`
	IntegrationToken string                    `koanf:"integration_token" validate:"required"`
	LazyLoad         bool                      `koanf:"lazy_load"`
	Databases        []ForwardedDatabaseConfig `koanf:"forwarded_databases" validate:"required"`
}
