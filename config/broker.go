package config

type Broker struct {
	Stan BrokerStan `json:"stan"`
}

type BrokerStan struct {
	Enable      bool     `json:"enable"`
	Addresses   []string `json:"addresses"`
	ClusterID   string   `json:"cluster_id"`
	DurableName string   `json:"durable_name"`
}

func DefaultBroker() *Broker {
	return &Broker{
		Stan: BrokerStan{
			Enable: true,
			Addresses: []string{
				"nats://localhost:4222",
			},
			ClusterID:   "test-cluster",
			DurableName: "mercury-durable",
		},
	}
}
