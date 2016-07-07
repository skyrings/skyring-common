package models

const (
	CLUSTER_AVAILABILITY = "cluster_availability"
	QUORUM_LOSS          = "quorum_loss"
	HOST_AVAILABILITY    = "host_availability"
)

type NotificationSubscription struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

var Notifications = []NotificationSubscription{
	{
		Name:        CLUSTER_AVAILABILITY,
		Description: "Cluster Availability",
		Enabled:     true,
	},
	{
		Name:        HOST_AVAILABILITY,
		Description: "Host Availability",
		Enabled:     true,
	},
	{
		Name:        QUORUM_LOSS,
		Description: "Quorum Loss",
		Enabled:     false,
	},
}
