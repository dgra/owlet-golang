package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/dgra/owlet-golang/client"
	"github.com/spf13/viper"
)

type config struct {
	Email    string `map_structure:"email"`
	Password string `map_structure:"password"`
}

// Load in config file(from the current directory) and marshalls into a config struct.
func LoadConfig() *config {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	config := &config{}
	err = viper.Unmarshal(config)
	if err != nil { // Handle errors unmarshalling the config values
		panic(fmt.Errorf("Fatal error config file unmarshalling: %s \n", err))
	}
	return config
}

func backoff(msg string, attempt int) int {
	maxAttempts := 20
	backoff_time := math.Pow(2.0, float64(attempt))
	fmt.Printf("Backing off for %.2f milliseconds. Attempt %d of %d\n", backoff_time, attempt, maxAttempts)
	fmt.Println(msg)

	if attempt >= maxAttempts {
		attempt = maxAttempts - 1
	}

	time.Sleep(time.Duration(backoff_time) * time.Millisecond)
	attempt++
	return attempt
}

func main() {
	config := LoadConfig()
	client, err := client.New(config.Email, config.Password)
	if err != nil {
		fmt.Printf("Failed to create client. %+v\n", err)
		return
	}

	err = client.SetFirstDevice()
	if err != nil {
		fmt.Printf("Failed to get devices. %+v\n", err)
		return
	}

	// Log out to file instead of new one each day.
	file, err := os.OpenFile("owlet_data.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed to create file. %s\n", err)
		return
	}
	defer file.Close()

	attempts := 1
	propAttempts := 1
	for {
		properties, err := client.GetProperties(client.Device.DSN)
		if err != nil {
			propAttempts = backoff(fmt.Sprintf("Failed to get properties for %s. %+v\n", client.Device.DSN, err), propAttempts)
			continue
		}
		propAttempts = 1

		props_json, err := json.Marshal(properties)
		if err != nil {
			fmt.Printf("Failed to unmarshal properties for %s. %+v\n", client.Device.DSN, err)
			continue
		}

		chargeStatus := properties["CHARGE_STATUS"].Value

		if chargeStatus == "2" {
			attempts = backoff(fmt.Sprintf("Backing off due to CHARGE_STATUS"), attempts)
			continue
		}
		attempts = 1

		// Append to our file.
		fmt.Fprintf(file, "%s\n", string(props_json))

		// STDOUT logging of specific stats.
		fmt.Printf("%+v\n", properties["CHARGE_STATUS"])
		fmt.Printf("%+v\n", properties["BASE_STATION_ON"])
		fmt.Printf("%+v\n", properties["BATT_LEVEL"])
		fmt.Printf("%+v\n", properties["OXYGEN_LEVEL"])
		fmt.Printf("%+v\n", properties["HEART_RATE"])

		_, err = client.SetAppActiveStatus(client.Device.DSN)
		if err != nil {
			fmt.Printf("Failed to set APP_ACTIVE: %+v\n", err)
		}

		time.Sleep(2 * time.Second)
	}
}
