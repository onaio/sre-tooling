package api_1_8_0

// APIResponse does..
type APIResponse struct {
	BulletinBoard BulletinBoard `json:"bulletinBoard"`
}

// BulletinBoard does..
type BulletinBoard struct {
	Bulletins []BulletinProcessor `json:"bulletins"`
	Generated string              `json:"generated"`
}

// BulletinProcessor does...
type BulletinProcessor struct {
	GroupID  string   `json:"groupId"`
	SourceID string   `json:"sourceId"`
	CanRead  bool     `json:"canRead"`
	Bulletin Bulletin `json:"bulletin"`
}

// Bulletin does...
type Bulletin struct {
	ID         int    `json:"id"`
	Category   string `json:"category"`
	SourceName string `json:"sourceName"`
	Level      string `json:"level"`
	Message    string `json:"message"`
	Timestamp  string `json:"timestamp"`
}

type Event struct {
	runtimeName *string
	runtime     *string
	id          *string
	category    *string
	sourceId    *string
	sourceName  *string
	groupId     *string
	timestamp   *string
}

func (event *Event) New(id *string, category *string, sourceId *string, sourceName *string, groupId *string, timestamp *string, runtime *string, runtimeName *string) {
	event.runtimeName = runtimeName
	event.runtime = runtime
	event.id = id
	event.category = category
	event.sour
}

func (event *Event) GetCategory() string {

}
func (event *Event) GetId() string {

}
func (event *Event) GetSourceId() string {

}
func (event *Event) GetGroupId() string {

}
func (event *Event) GetSourceName() string {

}
func (event *Event) GetTimestamp() string {

}
func (event *Event) GetRuntime() string {

}
func (event *Event) GetRuntimeName() string {

}
