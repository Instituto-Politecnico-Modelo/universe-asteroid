package types

type Config struct {
	RabbitMQServer struct {
		URI      string `yaml:"uri" envconfig:"RABBITMQ_URI"`
		Port     string `yaml:"port" envconfig:"RABBITMQ_PORT"`
		Host     string `yaml:"host" envconfig:"RABBITMQ_HOST"`
		User     string `yaml:"user" envconfig:"RABBITMQ_USER"`
		Password string `yaml:"password" envconfig:"RABBITMQ_PASSWORD"`
	} `yaml:"rabbitmq_server"`
}
