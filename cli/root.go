package cli

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/uphy/coin-trade-history/services"
)

var root = &cobra.Command{
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var listCurrencies = &cobra.Command{
	Use:   "list-currencies",
	Short: "List the available currency pairs.",
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, service := range allServices {
			currencyPairs, err := service.GetCurrencyPairs()
			if err != nil {
				return err
			}
			fmt.Printf("[%s]\n", service.Name())
			for _, currencyPair := range currencyPairs {
				fmt.Printf("- %s\n", currencyPair)
			}
		}
		return nil
	},
}

var download = &cobra.Command{
	Use:   "download",
	Short: "Download the trade history.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("specify the destination file path")
		}
		file := args[0]
		allData := []services.TradeData{}
		for _, serviceConfig := range conf.Services {
			service, err := createService(&serviceConfig)
			if err != nil {
				return err
			}
			allData, err = getTradeData(service, allData, serviceConfig.Currencies)
			if err != nil {
				return err
			}
		}
		services.SortTradeData(allData)

		profit := 0.
		w := NewExcelWriter(file)
		defer w.Close()
		for _, data := range allData {
			profit += data.Profit
			w.Write(&data)
		}
		return nil
	},
}

func createService(conf *serviceConfig) (services.Service, error) {
	if conf.Service == "zaif" {
		return services.NewZaif(conf.Key, conf.Secret), nil
	} else if conf.Service == "bitflyer" {
		return services.NewBitFlyer(conf.Key, conf.Secret), nil
	}
	return nil, errors.New("unsupported service: " + conf.Service)
}

func getTradeData(service services.Service, allData []services.TradeData, currencyPairs []string) ([]services.TradeData, error) {
	if len(currencyPairs) == 0 {
		c, err := service.GetCurrencyPairs()
		if err != nil {
			return nil, err
		}
		currencyPairs = c
	}
	for _, currencyPair := range currencyPairs {
		data, err := service.GetTradeHistory(currencyPair, time.Now().Add(-time.Hour*24*365), time.Now())
		if err != nil {
			return nil, err
		}
		allData = append(allData, data...)
	}
	return allData, nil
}

var conf *config
var allServices []services.Service

func init() {
	c, err := readConfig("config.yml")
	if err != nil {
		panic(err)
	}
	conf = c
	for _, serviceConfig := range c.Services {
		service, err := createService(&serviceConfig)
		if err != nil {
			panic(err)
		}
		allServices = append(allServices, service)
	}
}

// Execute executes the user specified command.
func Execute() error {
	root.AddCommand(listCurrencies)
	root.AddCommand(download)
	return root.Execute()
}
