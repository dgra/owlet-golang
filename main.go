package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/dgra/owlet-golang/client"
	"github.com/spf13/viper"
)

var attempts = 0

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

	// fmt.Printf("Device: %+v\n", client.Device)

	// 1. Set App Active Status
	// 2. Read last datapoints.
	// 	2a. If Base station is off, start exponential backoff to n minutes(Don't dos ayla if not needed)? (20 minutes?)
	// 3. Wait n miliseconds
	// 4. GOTO 1

	attempts = 1

	filename := time.Now().UTC().Format(time.RFC3339)
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Failed to create file. %s\n", err)
		return
	}
	defer file.Close()

	for {
		// 2.
		properties, err := client.GetProperties(client.Device.DSN)
		if err != nil {
			fmt.Printf("Failed to get properties for %s. %+v\n", client.Device.DSN, err)
			return
		}

		// fmt.Println("Props:")
		// fmt.Printf("%+v\n", properties)
		props_json, err := json.Marshal(properties)
		fmt.Fprintf(file, "%s,\n", string(props_json))
		fmt.Printf("%+v\n", properties["CHARGE_STATUS"])
		fmt.Printf("%+v\n", properties["BASE_STATION_ON"])
		fmt.Printf("%+v\n", properties["BATT_LEVEL"])
		fmt.Printf("%+v\n", properties["OXYGEN_LEVEL"])
		fmt.Printf("%+v\n", properties["HEART_RATE"])

		fmt.Println("Attempt:", attempts)

		attempts++
		// 2a.
		if properties["CHARGE_STATUS"].Value != "2" {
			// 1.
			attempts = 1
			work, err := client.SetAppActiveStatus(client.Device.DSN)
			if err != nil {
				fmt.Printf("Failed to set APP_ACTIVE: %+v\n", err)
				return
			}
			fmt.Println("Work:", work)
		} else {
			attempts++
		}

		// Min wait is 2 seconds to give APP_ACTIVE to cause a record write.
		// backoff_time := math.Pow(2000.0, float64(attempts))
		// time.Sleep(time.Duration(backoff_time) * time.Millisecond)
		// time.Sleep(15 * time.Minute)
		time.Sleep(5 * time.Second)
	}
}
