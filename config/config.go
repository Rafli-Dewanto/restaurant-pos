package configs

type Config struct {
	DBName               string
	DBPassword           string
	DBUser               string
	DBPort               string
	DBHost               string
	JWT_SECRET           string
	MIDTRANS_MERCHANT_ID string
	MIDTRANS_CLIENT_KEY  string
	MIDTRANS_SERVER_KEY  string
	MIDTRANS_ENDPOINT    string
	SERVER_ENV           string
	SERVER_PORT          string
}

func LoadConfig() *Config {
	viper := NewViper()

	return &Config{
		DBName:               viper.GetString("MYSQL_DATABASE"),
		DBPassword:           viper.GetString("MYSQL_ROOT_PASSWORD"),
		DBUser:               viper.GetString("MYSQL_USER"),
		DBPort:               viper.GetString("MYSQL_PORT"),
		DBHost:               viper.GetString("MYSQL_HOST"),
		JWT_SECRET:           viper.GetString("JWT_SECRET"),
		MIDTRANS_MERCHANT_ID: viper.GetString("MIDTRANS_MERCHANT_ID"),
		MIDTRANS_CLIENT_KEY:  viper.GetString("MIDTRANS_CLIENT_KEY"),
		MIDTRANS_SERVER_KEY:  viper.GetString("MIDTRANS_SERVER_KEY"),
		MIDTRANS_ENDPOINT:    viper.GetString("MIDTRANS_ENDPOINT"),
		SERVER_ENV:           viper.GetString("SERVER_ENV"),
		SERVER_PORT:          viper.GetString("SERVER_PORT"),
	}
}
