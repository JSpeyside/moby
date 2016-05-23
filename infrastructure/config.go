package infrastructure

type Config struct {
	Version string
}

func LoadConfig() *Config {
	return &Config{
		Version: "0.0.1",
	}
}
