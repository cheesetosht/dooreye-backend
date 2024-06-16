package utility

import (
	"log"
	"os"
)

func ValidateEnv(vars []string) {
	log.Println("> validating environment variables")
	var missing []string
	for _, key := range vars {
		_, exists := os.LookupEnv(key)
		if !exists {
			missing = append(missing, key)
			log.Println("?? missing environment variable: ", key)
		} else {
			log.Println("> load environment variable: ", key)
		}
	}
	if len(missing) != 0 {
		log.Fatalln("!! exiting due to missing environment variables")
	}
}
