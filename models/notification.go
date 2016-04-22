package models

var NOTIFICATIONS_SUPPORTED = []string{
	CLUSTER_AVAILABILITY,
	QUORUM_LOSS,
	HOST_AVAILABILITY,
}

const (
	CLUSTER_AVAILABILITY = "cluster_availability"
	QUORUM_LOSS          = "quorum_loss"
	HOST_AVAILABILITY    = "host_availability"
)

type NotificationSubscription struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}