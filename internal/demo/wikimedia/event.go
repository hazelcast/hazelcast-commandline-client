package wikimedia

import (
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
	Timestamp time.Time `json:"dt,omitempty"`
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
			Value: ev.Meta.Timestamp,
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

func (ev event) KeyValues() map[string]any {
	return map[string]any{
		"schema":             ev.Schema,
		"meta_uri":           ev.Meta.URI,
		"meta_request_id":    ev.Meta.RequestID,
		"meta_id":            ev.Meta.ID,
		"meta_dt":            ev.Meta.Timestamp,
		"meta_domain":        ev.Meta.Domain,
		"meta_stream":        ev.Meta.Stream,
		"meta_topic":         ev.Meta.Topic,
		"meta_partition":     ev.Meta.Partition,
		"meta_offset":        ev.Meta.Offset,
		"id":                 ev.ID_,
		"type":               ev.Type,
		"namespace":          ev.Namespace,
		"title":              ev.Title,
		"title_url":          ev.TitleURL,
		"comment":            ev.Comment,
		"event_time":         ev.Timestamp,
		"user_name":          ev.User,
		"bot":                ev.Bot,
		"notify_url":         ev.NotifyURL,
		"minor":              ev.Minor,
		"length_old":         ev.Length.Old,
		"length_new":         ev.Length.New,
		"revision_old":       ev.Revision.Old,
		"revision_new":       ev.Revision.New,
		"server_url":         ev.ServerURL,
		"server_name":        ev.ServerName,
		"server_script_path": ev.ServerScriptPath,
		"wiki":               ev.Wiki,
		"parsed_comment":     ev.Parsedcomment,
	}
}
