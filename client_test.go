package gofalcon_test

import "os"

type testConfig struct {
	user  string
	token string

	verbose bool
}

var cfg testConfig

func init() {
	cfg = testConfig{
		user:    os.Getenv("FALCON_USER"),
		token:   os.Getenv("FALCON_TOKEN"),
		verbose: (os.Getenv("FALCON_TEST_VERBOSE") != ""),
	}
}
