/*
Copyright © 2024 OSMOS
*/
package cmd

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cobra"
)

type paramCheck struct {
	url   string
	param string
}

var transport = &http.Transport{
	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: time.Second,
		DualStack: true,
	}).DialContext,
}

var httpClient = &http.Client{
	Transport: transport,
}

func checkReflected(targetURL string) ([]string, error) {

	out := make([]string, 0)

	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return out, err
	}

	req.Header.Add("User-Agent", "User-Agent: Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.100 Safari/537.36")

	resp, err := httpClient.Do(req)
	if err != nil {
		return out, err
	}
	if resp.Body == nil {
		return out, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return out, err
	}

	if strings.HasPrefix(resp.Status, "3") {
		return out, nil
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "" && !strings.Contains(ct, "html") {
		return out, nil
	}

	body := string(b)

	u, err := url.Parse(targetURL)
	if err != nil {
		return out, err
	}

	for key, vv := range u.Query() {
		for _, v := range vv {
			if !strings.Contains(body, v) {
				continue
			}

			out = append(out, key)
		}
	}

	return out, nil
}

func checkAppend(targetURL, param, payload string) (bool, error) {
	
	u, err := url.Parse(targetURL)
	if err != nil {
		return false, err
	}

	qs := u.Query()
	//val := qs.Get(param)
	//if val == "" {
	//return false, nil
	//return false, fmt.Errorf("can't append to non-existant param %s", param)
	//}

	qs.Set(param, payload)
	u.RawQuery = qs.Encode()
	reflected, err := checkReflected(u.String())
	if err != nil {
		return false, err
	}

	for _, r := range reflected {
		if r == param {
			return true, nil
		}
	}

	return false, nil
}

type workerFunc func(paramCheck, chan paramCheck)

func makePool(input chan paramCheck, fn workerFunc) chan paramCheck {
	var wg sync.WaitGroup

	output := make(chan paramCheck)
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func() {
			for c := range input {
				fn(c, output)
			}
			wg.Done()
		}()
	}

	go func() {
		wg.Wait()
		close(output)
	}()

	return output
}

func main(url string) {
	httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	initialChecks := make(chan paramCheck, 40)

	appendChecks := makePool(initialChecks, func(c paramCheck, output chan paramCheck) {
		reflected, err := checkReflected(c.url)
		if err != nil {
			//fmt.Fprintf(os.Stderr, "error from checkReflected: %s\n", err)
			return
		}

		if len(reflected) == 0 {
			// TODO: wrap in verbose mode
			fmt.Printf("[*] No Reflection Detected")
			return
		}

		for _, param := range reflected {
			fmt.Printf("[*] Reflected Parameter: %s\n", param)
			fmt.Println("[*] Unfiltered: Alphabets")
			fmt.Println("[*] Unfiltered: Numbers")
			output <- paramCheck{c.url, param}
		}
	})

	charChecks := makePool(appendChecks, func(c paramCheck, output chan paramCheck) {
		wasReflected, err := checkAppend(c.url, c.param, "iy3j4h234hjb23234")
		if err != nil {
			fmt.Fprintf(os.Stderr, "error from checkAppend for url %s with param %s: %s", c.url, c.param, err)
			return
		}

		if wasReflected {
			output <- paramCheck{c.url, c.param}
		}
	})

	done := makePool(charChecks, func(c paramCheck, output chan paramCheck) {
		output_of_url := []string{c.url, c.param}
		for _, char := range []string{"\"", "'", "<", ">", "$", "|", "(", ")", "`", ":", ";", "{", "}"} {
			wasReflected, err := checkAppend(c.url, c.param, "pfx"+char+"sfx")
			if err != nil {
				fmt.Fprintf(os.Stderr, "error from checkAppend for url %s with param %s with %s: %s", c.url, c.param, char, err)
				continue
			}

			if wasReflected {
				output_of_url = append(output_of_url, char)
			}
		}
		if len(output_of_url) >= 2 {
			fmt.Printf("[*] Unfiltered: Special Chars %v \n", output_of_url[2:])
		}
	})

	initialChecks <- paramCheck{url: url}

	close(initialChecks)
	<-done
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
			main(url)
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
