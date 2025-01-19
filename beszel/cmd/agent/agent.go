package main

import (
	"beszel"
	"beszel/internal/agent"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	// Bayrakları / alt komutları işle
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v":
			fmt.Println(beszel.AppName+"-agent", beszel.Version)
		case "update":
			agent.Update()
		}
		os.Exit(0)
	}

	// Anahtarı KEY ortam değişkeninden almaya çalış.
	pubKey := []byte(os.Getenv("KEY"))

	// Eğer KEY ayarlanmamışsa, anahtarı KEY_FILE ile belirtilen dosyadan okumaya çalış.
	if len(pubKey) == 0 {
		keyFile, varMi := os.LookupEnv("KEY_FILE")
		if !varMi {
			log.Fatal("KEY veya KEY_FILE ortam değişkenini ayarlamalısınız")
		}
		var err error
		pubKey, err = os.ReadFile(keyFile)
		if err != nil {
			log.Fatal(err)
		}
	}

	addr := ":45876"
	if portEnvVar, varMi := os.LookupEnv("PORT"); varMi {
		// "127.0.0.1:45876" şeklinde bir adres geçilmesine izin ver
		if !strings.Contains(portEnvVar, ":") {
			portEnvVar = ":" + portEnvVar
		}
		addr = portEnvVar
	}

	agent.NewAgent().Run(pubKey, addr)
}
