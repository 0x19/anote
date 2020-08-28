package config

import "github.com/spf13/viper"

func Setup() {
	viper.SetDefault("development_mode", true)
	viper.SetDefault("currency_exchange_rate_scheme", "https")
	viper.SetDefault("currency_exchange_host", "min-api.cryptocompare.com")
	viper.SetDefault("currency_exchange_rate_path", "data/price")
	viper.SetDefault("currency_exchange_rate_client_timeout_s", 10)
	viper.SetDefault("currency_exchange_rate_transport_timeout_s", 5)
	viper.SetDefault("currency_exchange_rate_transport_handshake_timeout_s", 5)
	viper.SetDefault("currency_exchange_rate_cache_enabled", true)
	viper.SetDefault("currency_exchange_rate_cache_expiry_s", 60)
}
