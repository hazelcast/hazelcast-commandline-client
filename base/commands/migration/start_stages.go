package migration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	clcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"
	"golang.org/x/exp/slices"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Stages struct {
	migrationID       string
	configDir         string
	ci                *hazelcast.ClientInternal
	startQueue        *hazelcast.Queue
	statusMap         *hazelcast.Map
	updateTopic       *hazelcast.Topic
	topicListenerID   types.UUID
	updateMessageChan chan UpdateMessage
}

var timeoutErr = fmt.Errorf("migration could not be completed: reached timeout while reading status: "+
	"please ensure that you are using Hazelcast's migration cluster distribution and your DMT config points to that cluster: %w",
	context.DeadlineExceeded)

func NewStages(migrationID, configDir string) *Stages {
	if migrationID == "" {
		panic("migrationID is required")
	}
	return &Stages{
		migrationID: migrationID,
		configDir:   configDir,
	}
}

func (st *Stages) Build(ctx context.Context, ec plug.ExecContext) []stage.Stage {
	return []stage.Stage{
		{
			ProgressMsg: "Connecting to the migration cluster",
			SuccessMsg:  "Connected to the migration cluster",
			FailureMsg:  "Could not connect to the migration cluster",
			Func:        st.connectStage(ctx, ec),
		},
		{
			ProgressMsg: "Starting the migration",
			SuccessMsg:  "Started the migration",
			FailureMsg:  "Could not start the migration",
			Func:        st.startStage(ctx, ec),
		},
		{
			ProgressMsg: "Migrating the cluster",
			SuccessMsg:  "Migrated the cluster",
			FailureMsg:  "Could not migrate the cluster",
			Func:        st.migrateStage(ctx, ec),
		},
	}
}

func (st *Stages) connectStage(ctx context.Context, ec plug.ExecContext) func(stage.Statuser) error {
	return func(status stage.Statuser) error {
		var err error
		st.ci, err = ec.ClientInternal(ctx)
		if err != nil {
			return err
		}
		st.startQueue, err = st.ci.Client().GetQueue(ctx, StartQueueName)
		if err != nil {
			return err
		}
		st.statusMap, err = st.ci.Client().GetMap(ctx, MakeStatusMapName(st.migrationID))
		if err != nil {
			return err
		}
		st.updateTopic, err = st.ci.Client().GetTopic(ctx, MakeUpdateTopicName(st.migrationID))
		if err != nil {
			return err
		}
		st.updateMessageChan = make(chan UpdateMessage)
		_, err = st.updateTopic.AddMessageListener(ctx, st.topicListener)
		return err
	}
}

func (st *Stages) topicListener(event *hazelcast.MessagePublished) {
	st.updateMessageChan <- event.Value.(UpdateMessage)
}

func (st *Stages) startStage(ctx context.Context, ec plug.ExecContext) func(stage.Statuser) error {
	return func(stage.Statuser) error {
		if err := st.statusMap.Delete(ctx, StatusMapEntryName); err != nil {
			return err
		}
		var cb ConfigBundle
		cb.MigrationID = st.migrationID
		if err := cb.Walk(st.configDir); err != nil {
			return err
		}
		b, err := json.Marshal(cb)
		if err != nil {
			return err
		}
		if err = st.startQueue.Put(ctx, serialization.JSON(b)); err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		msg := <-st.updateMessageChan // read the first message
		if slices.Contains([]status{StatusComplete, StatusFailed, statusCanceled}, msg.Status) {
			ms, err := st.readMigrationStatus(ctx)
			if ctx.Err() != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					return timeoutErr
				}
				return fmt.Errorf("migration failed: %w", err)
			}
			if err != nil {
				return fmt.Errorf("reading status: %w", err)
			}
			ec.PrintlnUnnecessary(msg.Message)
			ec.PrintlnUnnecessary(ms.Report)
			for _, l := range ms.Logs {
				ec.Logger().Info(l)
			}
			switch ms.Status {
			case StatusComplete:
				return nil
			case statusCanceled:
				return clcerrors.ErrUserCancelled
			case StatusFailed:
				return fmt.Errorf("migration failed with following error(s): %s", strings.Join(ms.Errors, "\n"))
			}
		} else {
			ec.PrintlnUnnecessary(msg.Message)
		}
		return nil
	}
}

func (st *Stages) migrateStage(ctx context.Context, ec plug.ExecContext) func(statuser stage.Statuser) error {
	return func(stage.Statuser) error {
		defer st.updateTopic.RemoveListener(ctx, st.topicListenerID)
		for {
			select {
			case msg := <-st.updateMessageChan:
				if slices.Contains([]status{StatusComplete, StatusFailed, statusCanceled}, msg.Status) {
					ms, err := st.readMigrationStatus(ctx)
					if err != nil {
						return fmt.Errorf("reading status: %w", err)
					}
					ec.PrintlnUnnecessary(msg.Message)
					ec.PrintlnUnnecessary(ms.Report)
					for _, l := range ms.Logs {
						ec.Logger().Info(l)
					}
					switch ms.Status {
					case StatusComplete:
						return nil
					case statusCanceled:
						return clcerrors.ErrUserCancelled
					case StatusFailed:
						return fmt.Errorf("migration failed with following error(s): %s", strings.Join(ms.Errors, "\n"))
					}
				} else {
					ec.PrintlnUnnecessary(msg.Message)
				}
			case <-ctx.Done():
				if err := ctx.Err(); err != nil {
					if errors.Is(err, context.DeadlineExceeded) {
						return timeoutErr
					}
					return fmt.Errorf("migration failed: %w", err)
				}
			}
		}
	}
}

func (st *Stages) readMigrationStatus(ctx context.Context) (MigrationStatus, error) {
	v, err := st.statusMap.Get(ctx, StatusMapEntryName)
	if err != nil {
		return migrationStatusNone, err
	}
	if v == nil {
		return migrationStatusNone, nil
	}
	var b []byte
	if vv, ok := v.(string); ok {
		b = []byte(vv)
	} else if vv, ok := v.(serialization.JSON); ok {
		b = vv
	} else {
		return migrationStatusNone, fmt.Errorf("invalid status value")
	}
	var ms MigrationStatus
	if err := json.Unmarshal(b, &ms); err != nil {
		return migrationStatusNone, fmt.Errorf("unmarshaling status: %w", err)
	}
	return ms, nil
}

func MakeStatusMapName(migrationID string) string {
	return "__datamigration_" + migrationID
}

func MakeUpdateTopicName(migrationID string) string {
	return updateTopic + migrationID
}

type status string

const (
	statusNone       status = ""
	StatusComplete   status = "COMPLETED"
	statusCanceled   status = "CANCELED"
	StatusFailed     status = "FAILED"
	StatusInProgress status = "IN_PROGRESS"
)

type MigrationStatus struct {
	Status status   `json:"status"`
	Logs   []string `json:"logs"`
	Errors []string `json:"errors"`
	Report string   `json:"report"`
}

var migrationStatusNone = MigrationStatus{
	Status: statusNone,
	Logs:   nil,
	Errors: nil,
	Report: "",
}

type UpdateMessage struct {
	Status               status  `json:"status"`
	CompletionPercentage float32 `json:"completionPercentage"`
	Message              string  `json:"message"`
}
