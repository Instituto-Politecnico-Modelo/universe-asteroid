package types

type Config struct {
	RabbitMQServer struct {
		URI            string `yaml:"uri" env:"RABBITMQ_URI"`
		Port           string `yaml:"port" env:"RABBITMQ_PORT"`
		Host           string `yaml:"host" env:"RABBITMQ_HOST"`
		User           string `yaml:"user" env:"RABBITMQ_USER"`
		Password       string `yaml:"password" env:"RABBITMQ_PASSWORD"`
		MaxQueueLength int    `yaml:"max_queue_length" env:"RABBITMQ_MAX_QUEUE_LENGTH"`
	} `yaml:"rabbitmq_server"`
	MongoDB struct {
		URI string `yaml:"uri" env:"MONGODB_URI"`
		DB  string `yaml:"db" env:"MONGODB_DB"`
	} `yaml:"mongodb"`
	Snapshot struct {
		Directory  string `yaml:"directory" env:"SNAPSHOT_DIRECTORY"`
		URL_Prefix string `yaml:"url_prefix" env:"SNAPSHOT_URL_PREFIX"`
	} `yaml:"snapshot"`
}
