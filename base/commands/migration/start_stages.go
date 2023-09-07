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

type StartStages struct {
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

func NewStartStages(migrationID, configDir string) *StartStages {
	if migrationID == "" {
		panic("migrationID is required")
	}
	return &StartStages{
		migrationID: migrationID,
		configDir:   configDir,
	}
}

func (st *StartStages) Build(ctx context.Context, ec plug.ExecContext) []stage.Stage {
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

func (st *StartStages) connectStage(ctx context.Context, ec plug.ExecContext) func(stage.Statuser) error {
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

func (st *StartStages) topicListener(event *hazelcast.MessagePublished) {
	st.updateMessageChan <- event.Value.(UpdateMessage)
}

func (st *StartStages) startStage(ctx context.Context, ec plug.ExecContext) func(stage.Statuser) error {
	return func(stage.Statuser) error {
		cb, err := makeConfigBundle(st.configDir, st.migrationID)
		if err != nil {
			return err
		}
		if err = st.startQueue.Put(ctx, cb); err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if isTerminal, err := st.handleUpdateMessage(ctx, ec, <-st.updateMessageChan); isTerminal {
			return err
		}
		return nil
	}
}

func makeConfigBundle(configDir, migrationID string) (serialization.JSON, error) {
	var cb ConfigBundle
	cb.MigrationID = migrationID
	if err := cb.Walk(configDir); err != nil {
		return nil, err
	}
	b, err := json.Marshal(cb)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (st *StartStages) migrateStage(ctx context.Context, ec plug.ExecContext) func(statuser stage.Statuser) error {
	return func(stage.Statuser) error {
		defer st.updateTopic.RemoveListener(ctx, st.topicListenerID)
		for {
			select {
			case msg := <-st.updateMessageChan:
				if isTerminal, err := st.handleUpdateMessage(ctx, ec, msg); isTerminal {
					return err
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

func (st *StartStages) handleUpdateMessage(ctx context.Context, ec plug.ExecContext, msg UpdateMessage) (bool, error) {
	ec.PrintlnUnnecessary(msg.Message)
	if slices.Contains([]Status{StatusComplete, StatusFailed, StatusCanceled}, msg.Status) {
		ms, err := readMigrationStatus(ctx, st.statusMap)
		if err != nil {
			return true, fmt.Errorf("reading status: %w", err)
		}
		ec.PrintlnUnnecessary(ms.Report)
		for _, l := range ms.Logs {
			ec.Logger().Info(l)
		}
		switch ms.Status {
		case StatusComplete:
			return true, nil
		case StatusCanceled:
			return true, clcerrors.ErrUserCancelled
		case StatusFailed:
			return true, fmt.Errorf("migration failed with following error(s):\n%s", strings.Join(ms.Errors, "\n"))
		}
	}
	return false, nil
}
