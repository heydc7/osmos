/*
Copyright © 2024 OSMOS
*/
package cmd

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func checkParameterReflection(inputURL string) {
	// Parse the URL
	parsedURL, err := url.Parse(inputURL)
	if err != nil {
		fmt.Println("Error parsing URL:", err)
		return
	}

	// Extract host and parameters
	// host := parsedURL.Host
	parameters := parsedURL.Query()

	// Send HTTP request
	resp, err := http.Get(inputURL)
	if err != nil {
		fmt.Println("Error making HTTP request:", err)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}
	fmt.Println(resp.Body)
	// Check if any parameter value is reflected in the response
	for key, values := range parameters {
		for _, value := range values {
			if strings.Contains(string(body), value) {
				fmt.Printf("Parameter value %s is reflected in the response for parameter <%s>\n", value, key)
				fuzzParameter(key, value, inputURL)
			} else {
				fmt.Printf("Parameter value %s is not reflected in the response for parameter <%s>\n", value, key)
			}
		}
	}
}

func fuzzParameter(parameterKey, originalValue, inputURL string) {
	// Character sets
	characterSets := map[string]string{
		"Alphabets":   "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		"Numbers":     "0123456789",
		"Punctuation": "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~",
	}

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Fuzz the parameter with random sets of random lengths
	for setName, charSet := range characterSets {
		fmt.Printf("Fuzzing parameter %s with %s:\n", parameterKey, setName)
		for i := 0; i < 5; i++ { // fuzz with 5 random sets
			fuzzedValue := generateRandomSet(charSet)
			// Construct a new URL with the fuzzed parameter
			fuzzedURL := strings.Replace(inputURL, originalValue, fuzzedValue, -1)
			// Send a request with the fuzzed URL
			resp, err := http.Get(fuzzedURL)
			if err != nil {
				fmt.Println("Error making HTTP request:", err)
				continue
			}
			defer resp.Body.Close()
			// Check if the fuzzed value is reflected
			var body strings.Builder
			_, err = io.Copy(&body, resp.Body)
			if err != nil {
				fmt.Println("Error reading response body:", err)
				continue
			}
			if strings.Contains(body.String(), fuzzedValue) {
				fmt.Printf("Fuzzed value %s is reflected in the response for parameter <%s>\n", fuzzedValue, parameterKey)
			} else {
				fmt.Printf("Fuzzed value %s is not reflected in the response for parameter <%s>\n", fuzzedValue, parameterKey)
			}
		}
	}
}

func generateRandomSet(charSet string) string {
	length := rand.Intn(5) + 1 // random length between 1 and 5
	var sb strings.Builder
	for i := 0; i < length; i++ {
		randomIndex := rand.Intn(len(charSet))
		sb.WriteByte(charSet[randomIndex])
	}
	return sb.String()
}

var rootCmd = &cobra.Command{
	Use:   "osmos",
	Short: "A brief description of your application",
	Long: `
██████╗ ███████╗███╗   ███╗ ██████╗ ███████╗
██╔═══██╗██╔════╝████╗ ████║██╔═══██╗██╔════╝
██║   ██║███████╗██╔████╔██║██║   ██║███████╗
██║   ██║╚════██║██║╚██╔╝██║██║   ██║╚════██║
╚██████╔╝███████║██║ ╚═╝ ██║╚██████╔╝███████║
╚═════╝ ╚══════╝╚═╝     ╚═╝ ╚═════╝ ╚══════╝
     
Disclaimer: 
Usage of osmos for fuzzing targets without prior mutual consent is illegal. It is the end user's responsibility to obey all applicable local, state and federal laws. Developers assume no liability and are not responsible for any misuse or damage caused by this program.
     `,
	Run: func(cmd *cobra.Command, args []string) {
		url, _ := cmd.Flags().GetString("url")
		request, _ := cmd.Flags().GetString("request")

		if url != "" {
			fmt.Println("URL:", url)
			checkParameterReflection(url)
		}
		if request != "" {
			fmt.Println("Request:", request)
			// Handle request here
		}
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringP("url", "u", "", "URL")
	rootCmd.Flags().StringP("request", "r", "", "HTTP Request")
}
