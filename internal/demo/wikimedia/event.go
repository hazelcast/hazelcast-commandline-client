package wikimedia

import (
	"encoding/json"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type event struct {
	Schema           string   `json:"$schema,omitempty"`
	Meta             Meta     `json:"meta,omitempty"`
	ID_              int64    `json:"id,omitempty"`
	Type             string   `json:"type,omitempty"`
	Namespace        int64    `json:"namespace,omitempty"`
	Title            string   `json:"title,omitempty"`
	TitleURL         string   `json:"title_url,omitempty"`
	Comment          string   `json:"comment,omitempty"`
	Timestamp        int64    `json:"timestamp,omitempty"`
	User             string   `json:"user,omitempty"`
	Bot              bool     `json:"bot,omitempty"`
	NotifyURL        string   `json:"notify_url,omitempty"`
	Minor            bool     `json:"minor,omitempty"`
	Length           Length   `json:"length,omitempty"`
	Revision         Revision `json:"revision,omitempty"`
	ServerURL        string   `json:"server_url,omitempty"`
	ServerName       string   `json:"server_name,omitempty"`
	ServerScriptPath string   `json:"server_script_path,omitempty"`
	Wiki             string   `json:"wiki,omitempty"`
	Parsedcomment    string   `json:"parsedcomment,omitempty"`
}

type Meta struct {
	URI       string    `json:"uri,omitempty"`
	RequestID string    `json:"request_id,omitempty"`
	ID        string    `json:"id,omitempty"`
	Dt        time.Time `json:"dt,omitempty"`
	Domain    string    `json:"domain,omitempty"`
	Stream    string    `json:"stream,omitempty"`
	Topic     string    `json:"topic,omitempty"`
	Partition int64     `json:"partition,omitempty"`
	Offset    int64     `json:"offset,omitempty"`
}

type Length struct {
	Old int64 `json:"old,omitempty"`
	New int64 `json:"new,omitempty"`
}

type Revision struct {
	Old int64 `json:"old,omitempty"`
	New int64 `json:"new,omitempty"`
}

func (ev event) ID() string {
	return ev.Meta.ID
}

func (ev event) Row() output.Row {
	row := output.Row{
		output.Column{
			Name:  "ID",
			Type:  serialization.TypeString,
			Value: ev.Meta.ID,
		},
		output.Column{
			Name:  "Timestamp",
			Type:  serialization.TypeJavaOffsetDateTime,
			Value: ev.Meta.Dt,
		},
		output.Column{
			Name:  "User",
			Type:  serialization.TypeString,
			Value: ev.User,
		},
		output.Column{
			Name:  "Comment",
			Type:  serialization.TypeString,
			Value: ev.Comment,
		},
	}
	return row
}

func (ev event) FlatMap() map[string]any {
	fe := flatEvent{
		Schema:           ev.Schema,
		MetaURI:          ev.Meta.URI,
		MetaRequestID:    ev.Meta.RequestID,
		MetaID:           ev.Meta.ID,
		MetaDt:           ev.Meta.Dt,
		MetaDomain:       ev.Meta.Domain,
		MetaStream:       ev.Meta.Stream,
		MetaTopic:        ev.Meta.Topic,
		MetaPartition:    ev.Meta.Partition,
		MetaOffset:       ev.Meta.Offset,
		ID:               ev.ID_,
		Type:             ev.Type,
		Namespace:        ev.Namespace,
		Title:            ev.Title,
		TitleURL:         ev.TitleURL,
		Comment:          ev.Comment,
		Timestamp:        ev.Timestamp,
		User:             ev.User,
		Bot:              ev.Bot,
		NotifyURL:        ev.NotifyURL,
		Minor:            ev.Minor,
		LengthOld:        ev.Length.Old,
		LengthNew:        ev.Length.New,
		RevisionOld:      ev.Revision.Old,
		RevisionNew:      ev.Revision.New,
		ServerURL:        ev.ServerURL,
		ServerName:       ev.ServerName,
		ServerScriptPath: ev.ServerScriptPath,
		Wiki:             ev.Wiki,
		Parsedcomment:    ev.Parsedcomment,
	}
	var sm map[string]any
	inrec, _ := json.Marshal(fe)
	json.Unmarshal(inrec, &sm)
	return sm
}

type flatEvent struct {
	Schema           string    `json:"schema"`
	MetaURI          string    `json:"meta_uri"`
	MetaRequestID    string    `json:"meta_request_id"`
	MetaID           string    `json:"meta_id"`
	MetaDt           time.Time `json:"meta_dt"`
	MetaDomain       string    `json:"meta_domain"`
	MetaStream       string    `json:"meta_stream"`
	MetaTopic        string    `json:"meta_topic"`
	MetaPartition    int64     `json:"meta_partition"`
	MetaOffset       int64     `json:"meta_offset"`
	ID               int64     `json:"id"`
	Type             string    `json:"type"`
	Namespace        int64     `json:"namespace"`
	Title            string    `json:"title"`
	TitleURL         string    `json:"title_url"`
	Comment          string    `json:"comment"`
	Timestamp        int64     `json:"time_stamp"`
	User             string    `json:"user_"`
	Bot              bool      `json:"bot"`
	NotifyURL        string    `json:"notify_url"`
	Minor            bool      `json:"minor"`
	LengthOld        int64     `json:"length_old"`
	LengthNew        int64     `json:"length_new"`
	RevisionOld      int64     `json:"revision_old"`
	RevisionNew      int64     `json:"revision_new"`
	ServerURL        string    `json:"server_url"`
	ServerName       string    `json:"server_name"`
	ServerScriptPath string    `json:"server_script_path"`
	Wiki             string    `json:"wiki"`
	Parsedcomment    string    `json:"parsedcomment"`
}
