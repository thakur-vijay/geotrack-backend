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

type UserDisconnectedEvent struct {
	Type string `json:"type"`
	Data struct {
		UserID  string `json:"user_id"`
		GroupID string `json:"group_id"`
	} `json:"data"`
}

func NewUserDisconnectedEvent(userID, groupID string) UserDisconnectedEvent {
	event := UserDisconnectedEvent{
		Type: "user.disconnected",
	}

	event.Data.UserID = userID
	event.Data.GroupID = groupID

	return event
}
