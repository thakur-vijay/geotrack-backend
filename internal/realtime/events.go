package realtime

const (
	// Location
	EventLocationUpdated = "location.updated"

	// Groups
	EventGroupCreated      = "group.created"
	EventGroupUpdated      = "group.updated"
	EventGroupDeleted      = "group.deleted"
	EventGroupMemberJoined = "group.member_joined"
	EventGroupMemberLeft   = "group.member_left"

	// Geofence
	EventGeofenceCreated = "geofence.created"
	EventGeofenceUpdated = "geofence.updated"
	EventGeofenceDeleted = "geofence.deleted"

	// Geofence transitions
	EventGeofenceEntered = "geofence.entered"
	EventGeofenceExited  = "geofence.exited"
)
