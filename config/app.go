package config

const (
	ConfigAppGracefulShutDownTimeout = "app.gracefulShutDownTimeout"
	ConfigServerWriteTimeout         = "app.writeTimeout"
	ConfigServerReadTimeout          = "app.readTimeout"

	defaultAppGracefulShutDownTimeout int64 = 10000
	defaultserverWriteTimeout         int64 = 60000
	defaultserverReadTimeout          int64 = 60000
)

type AppConfig struct {
	GracefulShutDownTimeout int64 // in milliseconds
	ServerWriteTimeout      int64 // in milliseconds
	ServerReadTimeout       int64 // in milliseconds
}

func loadAppConfig() *AppConfig {
	return &AppConfig{
		GracefulShutDownTimeout: loadDefaultConfig(ConfigAppGracefulShutDownTimeout, defaultAppGracefulShutDownTimeout),
		ServerWriteTimeout:      loadDefaultConfig(ConfigServerWriteTimeout, defaultserverWriteTimeout),
		ServerReadTimeout:       loadDefaultConfig(ConfigServerReadTimeout, defaultserverReadTimeout),
	}
}
