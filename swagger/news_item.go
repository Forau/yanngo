package swagger

type NewsItem struct {
	NewsId      int64   `json:"news_id,omitempty"`
	SourceId    int64   `json:"source_id,omitempty"`
	Headline    string  `json:"headline,omitempty"`
	Body        string  `json:"body,omitempty"`
	Instruments []int64 `json:"instruments,omitempty"`
	Lang        string  `json:"lang,omitempty"`
	Typ         string  `json:"type,omitempty"`
	Timestamp   int64   `json:"timestamp,omitempty"`
}
