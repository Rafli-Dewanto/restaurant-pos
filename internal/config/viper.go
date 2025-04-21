package configs

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

func NewViper() *viper.Viper {
	config := viper.New()
	config.SetConfigName(".env")   
	config.SetConfigType("env")    
	config.AddConfigPath(".")      
	config.AddConfigPath("..")     
	config.AddConfigPath("../../") 

	err := config.ReadInConfig()
	if err != nil {
		log.Printf("No .env file found, using system environment variables %v", err.Error())
		panic(fmt.Errorf("fatal error config file: %v", err.Error()))
	}

	return config
}
