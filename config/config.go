package config

type Config struct {
	Listen string            `mapstructure:"listen" env:"LISTEN" default:"127.0.0.1:8080"`
	Users  []User            `mapstructure:"users"`
	Access map[string]Access `mapstructure:"access"`
}
