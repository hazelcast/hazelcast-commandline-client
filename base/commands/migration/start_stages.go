package migration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	clcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/cluster"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"
	"golang.org/x/exp/slices"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type StartStages struct {
	migrationID     string
	configDir       string
	ci              *hazelcast.ClientInternal
	startQueue      *hazelcast.Queue
	statusMap       *hazelcast.Map
	updateTopic     *hazelcast.Topic
	topicListenerID types.UUID
	updateMsgChan   chan UpdateMessage
}

var timeoutErr = fmt.Errorf("migration could not be completed: reached timeout while reading status: "+
	"please ensure that you are using Hazelcast's migration cluster distribution and your DMT config points to that cluster: %w",
	context.DeadlineExceeded)

func NewStartStages(updateTopic *hazelcast.Topic, migrationID, configDir string) *StartStages {
	if migrationID == "" {
		panic("migrationID is required")
	}
	return &StartStages{
		updateTopic: updateTopic,
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
		st.updateMsgChan = make(chan UpdateMessage)
		st.topicListenerID, err = st.updateTopic.AddMessageListener(ctx, st.topicListener)
		return err
	}
}

func (st *StartStages) topicListener(event *hazelcast.MessagePublished) {
	var u UpdateMessage
	err := json.Unmarshal(event.Value.(serialization.JSON), &u)
	if err != nil {
		panic(fmt.Errorf("receiving update from migration cluster: %w", err))
	}
	st.updateMsgChan <- u
}

func (st *StartStages) startStage(ctx context.Context, ec plug.ExecContext) func(stage.Statuser) error {
	return func(status stage.Statuser) error {
		cb, err := makeConfigBundle(st.configDir, st.migrationID)
		if err != nil {
			return err
		}
		if err = st.startQueue.Put(ctx, cb); err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if isTerminal, err := st.handleUpdateMessage(ctx, ec, <-st.updateMsgChan, status); isTerminal {
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
	return func(status stage.Statuser) error {
		for {
			select {
			case msg := <-st.updateMsgChan:
				if isTerminal, err := st.handleUpdateMessage(ctx, ec, msg, status); isTerminal {
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

func (st *StartStages) handleUpdateMessage(ctx context.Context, ec plug.ExecContext, msg UpdateMessage, status stage.Statuser) (bool, error) {
	status.SetProgress(msg.CompletionPercentage)
	ec.PrintlnUnnecessary(msg.Message)
	if slices.Contains([]Status{StatusComplete, StatusFailed, StatusCanceled}, msg.Status) {
		ms, err := readMigrationStatus(ctx, st.statusMap)
		if err != nil {
			return true, fmt.Errorf("reading status: %w", err)
		}
		ec.PrintlnUnnecessary(ms.Report)
		name := fmt.Sprintf("migration_report_%s", st.migrationID)
		if err = saveReportToFile(name, ms.Report); err != nil {
			return true, fmt.Errorf("writing report to file: %w", err)
		}
		if err = st.saveDebugLogs(ctx, st.ci.OrderedMembers()); err != nil {
			return true, fmt.Errorf("writing debug logs to file: %w", err)
		}
		ec.PrintlnUnnecessary(fmt.Sprintf("migration report saved to file: %s", name))
		for _, l := range ms.Logs {
			ec.Logger().Info(l)
		}
		switch ms.Status {
		case StatusComplete:
			return true, nil
		case StatusCanceled:
			return true, clcerrors.ErrUserCancelled
		case StatusFailed:
			return true, fmt.Errorf("migration failed")
		}
	}
	return false, nil
}

func saveReportToFile(fileName, report string) error {
	f, err := os.Create(fmt.Sprintf(fileName))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(report)
	return err
}

func (st *StartStages) saveDebugLogs(ctx context.Context, members []cluster.MemberInfo) error {
	for _, m := range members {
		f, err := os.Create(fmt.Sprintf("%s%s.log", DebugLogsListPrefix, m.UUID.String()))
		if err != nil {
			return err
		}
		defer f.Close()
		l, err := st.ci.Client().GetList(ctx, DebugLogsListPrefix+m.UUID.String())
		if err != nil {
			return err
		}
		logs, err := l.GetAll(ctx)
		if err != nil {
			return err
		}
		for _, log := range logs {
			if _, err = fmt.Fprintf(f, log.(string)); err != nil {
				return err
			}
		}
	}
	return nil
}
