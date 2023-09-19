//go:build std || migration

package migration

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"golang.org/x/exp/slices"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
)

type StatusStages struct {
	migrationID              string
	ci                       *hazelcast.ClientInternal
	migrationsInProgressList *hazelcast.List
	statusMap                *hazelcast.Map
	updateTopic              *hazelcast.Topic
	updateMsgChan            chan UpdateMessage
	logger                   log.Logger
}

func NewStatusStages(logger log.Logger) *StatusStages {
	return &StatusStages{logger: logger}
}

func (st *StatusStages) Build(ctx context.Context, ec plug.ExecContext) []stage.Stage[any] {
	return []stage.Stage[any]{
		{
			ProgressMsg: "Connecting to the migration cluster",
			SuccessMsg:  "Connected to the migration cluster",
			FailureMsg:  "Could not connect to the migration cluster",
			Func:        st.connectStage(ec),
		},
		{
			ProgressMsg: "Fetching migration status",
			SuccessMsg:  "Fetched migration status",
			FailureMsg:  "Could not fetch migration status",
			Func:        st.fetchStage(ec),
		},
	}
}

type MigrationInProgress struct {
	MigrationID string `json:"migrationId"`
}

func (st *StatusStages) connectStage(ec plug.ExecContext) func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, status stage.Statuser[any]) (any, error) {
		var err error
		st.ci, err = ec.ClientInternal(ctx)
		if err != nil {
			return nil, err
		}
		st.migrationsInProgressList, err = st.ci.Client().GetList(ctx, MigrationsInProgressList)
		if err != nil {
			return nil, err
		}
		all, err := st.migrationsInProgressList.GetAll(ctx)
		if err != nil {
			return nil, err
		}
		if len(all) == 0 {
			return nil, fmt.Errorf("there are no migrations are in progress on migration cluster")
		}
		var mip MigrationInProgress
		m := all[0].(serialization.JSON)
		err = json.Unmarshal(m, &mip)
		if err != nil {
			return nil, fmt.Errorf("parsing migration in progress: %w", err)
		}
		st.migrationID = mip.MigrationID
		st.statusMap, err = st.ci.Client().GetMap(ctx, MakeStatusMapName(st.migrationID))
		if err != nil {
			return nil, err
		}
		st.updateTopic, err = st.ci.Client().GetTopic(ctx, MakeUpdateTopicName(st.migrationID))
		return nil, err
	}
}

func (st *StatusStages) fetchStage(ec plug.ExecContext) func(context.Context, stage.Statuser[any]) (any, error) {
	return func(ctx context.Context, status stage.Statuser[any]) (any, error) {
		ms, err := readMigrationStatus(ctx, st.statusMap)
		if err != nil {
			return nil, fmt.Errorf("reading status: %w", err)
		}
		if slices.Contains([]Status{StatusComplete, StatusFailed, StatusCanceled}, ms.Status) {
			ec.PrintlnUnnecessary(ms.Report)
			return nil, nil
		}
		st.updateMsgChan = make(chan UpdateMessage)
		id, err := st.updateTopic.AddMessageListener(ctx, st.topicListener)
		defer st.updateTopic.RemoveListener(ctx, id)
		for {
			select {
			case msg := <-st.updateMsgChan:
				ec.PrintlnUnnecessary(msg.Message)
				status.SetProgress(msg.CompletionPercentage)
				if slices.Contains([]Status{StatusComplete, StatusFailed, StatusCanceled}, msg.Status) {
					ms, err := readMigrationStatus(ctx, st.statusMap)
					if err != nil {
						return nil, fmt.Errorf("reading status: %w", err)
					}
					ec.PrintlnUnnecessary(ms.Report)
					return nil, nil
				}
			}
		}
	}
}

func (st *StatusStages) topicListener(event *hazelcast.MessagePublished) {
	var u UpdateMessage
	v, ok := event.Value.(serialization.JSON)
	if !ok {
		st.logger.Warn(fmt.Sprintf("update message type is unexpected"))
		return
	}
	err := json.Unmarshal(v, &u)
	if err != nil {
		st.logger.Warn(fmt.Sprintf("receiving update from migration cluster: %s", err.Error()))
	}
	st.updateMsgChan <- u
}
