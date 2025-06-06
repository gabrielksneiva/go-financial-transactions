package config

type Config struct {
	APIPort      string
	FrontendPort string
	KafkaBroker  string
	KafkaTopic   string
	KafkaGroupID string
	DBHost       string
	DBPort       string
	DBUser       string
	DBPassword   string
	DBName       string
	RedisHost    string
	RedisDB      int
	JwtSecret    string
	TronWallet   string
}
