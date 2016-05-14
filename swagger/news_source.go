package swagger

type NewsSource struct {
	Name      string   `json:"name,omitempty"`
	SourceId  int64    `json:"source_id,omitempty"`
	Level     string   `json:"level,omitempty"`
	Countries []string `json:"countries,omitempty"`
}
