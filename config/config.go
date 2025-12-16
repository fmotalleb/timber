package config

type Config struct {
	Users  []User            `mapstructure:"users"`
	Access map[string]Access `mapstructure:"access"`
}
