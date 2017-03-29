package marathon

import "encoding/json"

// GroupsResponse json returned from Marathon with apps definitions
type GroupsResponse struct {
	Groups []*Group `json:"groups"`
}

// Group represents application returned in Marathon json
type Group struct {
	Apps    []*App   `json:"apps"`
	Groups  []*Group `json:"groups"`
	ID      GroupID  `json:"id"`
	Version string   `json:"version"`
}

// IsEmpty checks if group is an empty group
func (g *Group) IsEmpty() bool {
	return len(g.Apps) == 0 && len(g.Groups) == 0
}

// GroupID represents group id from marathon
type GroupID string

//// String stringer for group
func (id GroupID) String() string {
	return string(id)
}

// ParseGroups json
func ParseGroups(jsonBlob []byte) ([]*Group, error) {
	groups := &GroupsResponse{}
	err := json.Unmarshal(jsonBlob, groups)

	return groups.Groups, err
}
