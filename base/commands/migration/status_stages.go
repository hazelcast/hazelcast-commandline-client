package migration

import (
	"context"
	"encoding/json"
	"fmt"

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
}

func NewStatusStages() *StatusStages {
	return &StatusStages{}
}

func (st *StatusStages) Build(ctx context.Context, ec plug.ExecContext) []stage.Stage {
	return []stage.Stage{
		{
			ProgressMsg: "Connecting to the migration cluster",
			SuccessMsg:  "Connected to the migration cluster",
			FailureMsg:  "Could not connect to the migration cluster",
			Func:        st.connectStage(ctx, ec),
		},
		{
			ProgressMsg: "Fetching migration status",
			SuccessMsg:  "Fetched migration status",
			FailureMsg:  "Could not fetch migration status",
			Func:        st.fetchStage(ctx, ec),
		},
	}
}

type MigrationInProgress struct {
	MigrationID string `json:"migrationId"`
}

func (st *StatusStages) connectStage(ctx context.Context, ec plug.ExecContext) func(statuser stage.Statuser) error {
	return func(status stage.Statuser) error {
		var err error
		st.ci, err = ec.ClientInternal(ctx)
		if err != nil {
			return err
		}
		st.migrationsInProgressList, err = st.ci.Client().GetList(ctx, MigrationsInProgressList)
		if err != nil {
			return err
		}
		all, err := st.migrationsInProgressList.GetAll(ctx)
		if err != nil {
			return err
		}
		if len(all) == 0 {
			return fmt.Errorf("there are no migrations are in progress on migration cluster")
		}
		var mip MigrationInProgress
		m := all[0].(serialization.JSON)
		err = json.Unmarshal(m, &mip)
		if err != nil {
			return fmt.Errorf("parsing migration in progress: %w", err)
		}
		st.migrationID = mip.MigrationID
		st.statusMap, err = st.ci.Client().GetMap(ctx, MakeStatusMapName(st.migrationID))
		if err != nil {
			return err
		}
		st.updateTopic, err = st.ci.Client().GetTopic(ctx, MakeUpdateTopicName(st.migrationID))
		return err
	}
}

func (st *StatusStages) fetchStage(ctx context.Context, ec plug.ExecContext) func(statuser stage.Statuser) error {
	return func(stage.Statuser) error {
		ms, err := readMigrationStatus(ctx, st.statusMap)
		if err != nil {
			return fmt.Errorf("reading status: %w", err)
		}
		if slices.Contains([]Status{StatusComplete, StatusFailed, StatusCanceled}, ms.Status) {
			ec.PrintlnUnnecessary(ms.Report)
			return nil
		}
		st.updateMsgChan = make(chan UpdateMessage)
		id, err := st.updateTopic.AddMessageListener(ctx, st.topicListener)
		defer st.updateTopic.RemoveListener(ctx, id)
		for {
			select {
			case msg := <-st.updateMsgChan:
				ec.PrintlnUnnecessary(msg.Message)
				ec.PrintlnUnnecessary(fmt.Sprintf("Completion Percentage: %f", msg.CompletionPercentage))
				if slices.Contains([]Status{StatusComplete, StatusFailed, StatusCanceled}, msg.Status) {
					ms, err := readMigrationStatus(ctx, st.statusMap)
					if err != nil {
						return fmt.Errorf("reading status: %w", err)
					}
					ec.PrintlnUnnecessary(ms.Report)
					return nil
				}
			}
		}
	}
}

func (st *StatusStages) topicListener(event *hazelcast.MessagePublished) {
	var u UpdateMessage
	err := json.Unmarshal(event.Value.(serialization.JSON), &u)
	if err != nil {
		panic(fmt.Errorf("receiving update from migration cluster: %w", err))
	}
	st.updateMsgChan <- u
}
