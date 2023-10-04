package health

type HealthCheck interface {
	GetRabbitStatus() bool
	SetRabbitStatus(status bool)
}
