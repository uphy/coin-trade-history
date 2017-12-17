package cli

import "gopkg.in/yaml.v2"
import "io/ioutil"

type serviceConfig struct {
	Service    string
	Key        string
	Secret     string
	Currencies []string
}
type config struct {
	Services []serviceConfig
}

func readConfig(file string) (*config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var config config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
