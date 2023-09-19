//go:build std || migration

package migration

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	clcerrors "github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
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
	reportOutputDir string
	logger          log.Logger
}

var timeoutErr = fmt.Errorf("migration could not be completed: reached timeout while reading status: "+
	"please ensure that you are using Hazelcast's migration cluster distribution and your DMT config points to that cluster: %w",
	context.DeadlineExceeded)

func NewStartStages(logger log.Logger, updateTopic *hazelcast.Topic, migrationID, configDir, reportOutputDir string) *StartStages {
	if migrationID == "" {
		panic("migrationID is required")
	}
	return &StartStages{
		updateTopic:     updateTopic,
		migrationID:     migrationID,
		configDir:       configDir,
		reportOutputDir: reportOutputDir,
		logger:          logger,
	}
}

func (st *StartStages) Build(ctx context.Context, ec plug.ExecContext) []stage.Stage[any] {
	return []stage.Stage[any]{
		{
			ProgressMsg: "Connecting to the migration cluster",
			SuccessMsg:  "Connected to the migration cluster",
			FailureMsg:  "Could not connect to the migration cluster",
			Func:        st.connectStage(ec),
		},
		{
			ProgressMsg: "Starting the migration",
			SuccessMsg:  "Started the migration",
			FailureMsg:  "Could not start the migration",
			Func:        st.startStage(ec),
		},
		{
			ProgressMsg: "Migrating the cluster",
			SuccessMsg:  "Migrated the cluster",
			FailureMsg:  "Could not migrate the cluster",
			Func:        st.migrateStage(ec),
		},
	}
}

func (st *StartStages) connectStage(ec plug.ExecContext) func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, status stage.Statuser[any]) (any, error) {
		var err error
		st.ci, err = ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		st.startQueue, err = st.ci.Client().GetQueue(ctx, StartQueueName)
		if err != nil {
			return nil, fmt.Errorf("retrieving the start Queue: %w", err)
		}
		st.statusMap, err = st.ci.Client().GetMap(ctx, MakeStatusMapName(st.migrationID))
		if err != nil {
			return nil, fmt.Errorf("retrieving the status Map: %w", err)
		}
		st.updateTopic, err = st.ci.Client().GetTopic(ctx, MakeUpdateTopicName(st.migrationID))
		if err != nil {
			return nil, fmt.Errorf("retrieving the update Topic: %w", err)
		}
		st.updateMsgChan = make(chan UpdateMessage)
		st.topicListenerID, err = st.updateTopic.AddMessageListener(ctx, st.topicListener)
		if err != nil {
			return nil, fmt.Errorf("adding message listener to update Topic: %w", err)

		}
		return nil, nil
	}
}

func (st *StartStages) topicListener(event *hazelcast.MessagePublished) {
	var u UpdateMessage
	err := json.Unmarshal(event.Value.(serialization.JSON), &u)
	if err != nil {
		st.logger.Warn(fmt.Sprintf("receiving update from migration cluster: %s", err.Error()))
	}
	st.updateMsgChan <- u
}

func (st *StartStages) startStage(ec plug.ExecContext) func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, status stage.Statuser[any]) (any, error) {
		cb, err := makeConfigBundle(st.configDir, st.migrationID)
		if err != nil {
			return nil, fmt.Errorf("making configuration bundle: %w", err)
		}
		if err = st.startQueue.Put(ctx, cb); err != nil {
			return nil, fmt.Errorf("updating start Queue: %w", err)
		}
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		if isTerminal, err := st.handleUpdateMessage(ctx, ec, <-st.updateMsgChan, status); isTerminal {
			return nil, err
		}
		return nil, nil
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

func (st *StartStages) migrateStage(ec plug.ExecContext) func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, status stage.Statuser[any]) (any, error) {
		for {
			select {
			case msg := <-st.updateMsgChan:
				if isTerminal, err := st.handleUpdateMessage(ctx, ec, msg, status); isTerminal {
					return nil, err
				}
			case <-ctx.Done():
				if err := ctx.Err(); err != nil {
					if errors.Is(err, context.DeadlineExceeded) {
						return nil, timeoutErr
					}
					return nil, fmt.Errorf("migration failed: %w", err)
				}
			}
		}
	}
}

func (st *StartStages) handleUpdateMessage(ctx context.Context, ec plug.ExecContext, msg UpdateMessage, status stage.Statuser[any]) (bool, error) {
	status.SetProgress(msg.CompletionPercentage)
	ec.PrintlnUnnecessary(msg.Message)
	if slices.Contains([]Status{StatusComplete, StatusFailed, StatusCanceled}, msg.Status) {
		ms, err := readMigrationStatus(ctx, st.statusMap)
		if err != nil {
			return true, fmt.Errorf("reading status: %w", err)
		}
		ec.PrintlnUnnecessary(ms.Report)
		var name string
		if st.reportOutputDir == "" {
			name = fmt.Sprintf("migration_report_%s.txt", st.migrationID)
		}
		if err = saveReportToFile(name, ms.Report); err != nil {
			return true, fmt.Errorf("writing report to file: %w", err)
		}
		if err = st.saveDebugLogs(ctx, ec, st.migrationID, st.ci.OrderedMembers()); err != nil {
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
	return os.WriteFile(fileName, []byte(report), 0600)
}

func (st *StartStages) saveDebugLogs(ctx context.Context, ec plug.ExecContext, migrationID string, members []cluster.MemberInfo) error {
	for _, m := range members {
		l, err := st.ci.Client().GetList(ctx, DebugLogsListPrefix+m.UUID.String())
		if err != nil {
			return err
		}
		logs, err := l.GetAll(ctx)
		if err != nil {
			return err
		}
		for _, line := range logs {
			ec.Logger().Debugf(fmt.Sprintf("[%s_%s] %s", migrationID, m.UUID.String(), line.(string)))
		}
	}
	return nil
}
