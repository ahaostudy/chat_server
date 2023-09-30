package config

import (
	"flag"
)

var (
	HOST string
	PORT int64

	PROXY   string
	MODEL   string = "gpt-3.5-turbo"
)

func init() {
	flag.StringVar(&HOST, "host", "0.0.0.0", "host")
	flag.Int64Var(&PORT, "port", 8080, "port")
	flag.StringVar(&PROXY, "proxy", "https://openai.ahao.ink/", "openai proxy")
	flag.Parse()
}
