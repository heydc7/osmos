# osmos
Osmos is a command-line tool written in Go for analyzing HTTP responses and checking if parameter values are reflected. It fuzzes parameter values with random sets of characters from various character sets and checks if the modified values are reflected in the HTTP response.
     

### Installation

```
git clone https://github.com/your-username/osmos.git
cd osmos
go build -o osmos cmd/main.go
```


### Usage:
  osmos [flags]

**Flags:**

    -h, --help             help for osmos
    -r, --request string   HTTP Request
    -u, --url string       URL

**Run the osmos executable with the desired URL or request as follows:**

```
./osmos -u <URL>
```


### Disclaimer: 
Usage of osmos for fuzzing targets without prior mutual consent is illegal. It is the end user's responsibility to obey all applicable local, state and federal laws. Developers assume no liability and are not responsible for any misuse or damage caused by this program.

### Contributing

Contributions are welcome! Please feel free to submit bug reports or feature requests.

### License

This project is licensed under the MIT License.