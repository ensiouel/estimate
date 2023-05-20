package entity

type Metrics []Metric

type Metric struct {
	Endpoint string `json:"endpoint"`
	Count    int    `json:"count"`
}
