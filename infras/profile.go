package infras

type UserProfile struct {
	ID                 *string       `json:"id"`
	Username           *string       `json:"username,omitempty"`
	Name               *string       `json:"name"`
	About              *string       `json:"about,omitempty"` // Bio
	Email              *string       `json:"email,omitempty"`
	Birthday           *string       `json:"birthday,omitempty"`
	Gender             *string       `json:"gender,omitempty"`
	Quotes             *string       `json:"quotes,omitempty"`
	Link               *string       `json:"link"`
	Locale             *string       `json:"locale,omitempty"`
	RelationshipStatus *string       `json:"relationship_status,omitempty"`
	UpdatedTime        *string       `json:"updated_time"`
	Location           *EntityNameID `json:"location,omitempty"`
	Hometown           *EntityNameID `json:"hometown,omitempty"`
	Work               *[]Work       `json:"work,omitempty"`
	Education          *[]Education  `json:"education,omitempty"`
}

type Work struct {
	EndDate   *string       `json:"end_date,omitempty"`
	Employer  *EntityNameID `json:"employer,omitempty"`
	StartDate *string       `json:"start_date,omitempty"`
	Position  *EntityNameID `json:"position,omitempty"`
	Location  *EntityNameID `json:"location,omitempty"`
}

type Education struct {
	School        *EntityNameID   `json:"school,omitempty"`
	Type          *string         `json:"type,omitempty"`
	Id            *string         `json:"id,omitempty"`
	Concentration *[]EntityNameID `json:"concentration,omitempty"`
}
