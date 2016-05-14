package swagger

import ()

type NewsPreview struct {
	NewsId      int64   `json:"news_id,omitempty"`
	SourceId    int64   `json:"source_id,omitempty"`
	Headline    string  `json:"headline,omitempty"`
	Instruments []int32 `json:"instruments,omitempty"`
	Lang        string  `json:"lang,omitempty"`
	Type_       string  `json:"type,omitempty"`
	Timestamp   int64   `json:"timestamp,omitempty"`
}
