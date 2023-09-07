package migration

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	serialization2 "github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
	"github.com/hazelcast/hazelcast-go-client"
	"github.com/hazelcast/hazelcast-go-client/serialization"
	"github.com/hazelcast/hazelcast-go-client/types"
)

type StatusStages struct {
	migrationID              string
	ci                       *hazelcast.ClientInternal
	migrationsInProgressList *hazelcast.List
	statusMap                *hazelcast.Map
	updateTopic              *hazelcast.Topic
	topicListenerID          types.UUID
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
		if err != nil {
			return err
		}
		st.updateMsgChan = make(chan UpdateMessage)
		_, err = st.updateTopic.AddMessageListener(ctx, st.topicListener)
		return err
	}
}

func (st *StatusStages) fetchStage(ctx context.Context, ec plug.ExecContext) func(statuser stage.Statuser) error {
	return func(stage.Statuser) error {
		defer st.updateTopic.RemoveListener(ctx, st.topicListenerID)
		for {
			select {
			case msg := <-st.updateMsgChan:
				ms, err := readMigrationStatus(ctx, st.statusMap)
				if err != nil {
					return fmt.Errorf("reading status: %w", err)
				}
				ec.PrintlnUnnecessary(msg.Message)
				if slices.Contains([]Status{StatusComplete, StatusFailed, StatusCanceled}, msg.Status) {
					ec.PrintlnUnnecessary(fmt.Sprintf("Completion Percentage: %f", ms.CompletionPercentage))
					ec.PrintlnUnnecessary(ms.Report)
					if len(ms.Errors) > 0 {
						ec.PrintlnUnnecessary(fmt.Sprintf("migration failed with following error(s):\n%s", strings.Join(ms.Errors, "\n")))
					}
					if len(ms.Migrations) > 0 {
						var rows []output.Row
						for _, m := range ms.Migrations {
							rows = append(rows, output.Row{
								output.Column{
									Name:  "Name",
									Type:  serialization2.TypeString,
									Value: m.Name,
								},
								output.Column{
									Name:  "Type",
									Type:  serialization2.TypeString,
									Value: m.Type,
								},
								output.Column{
									Name:  "Status",
									Type:  serialization2.TypeString,
									Value: string(m.Status),
								},
								output.Column{
									Name:  "Start Timestamp",
									Type:  serialization2.TypeJavaLocalDateTime,
									Value: types.LocalDateTime(m.StartTimestamp),
								},
								output.Column{
									Name:  "Entries Migrated",
									Type:  serialization2.TypeInt32,
									Value: int32(m.EntriesMigrated),
								},
								output.Column{
									Name:  "Total Entries",
									Type:  serialization2.TypeInt32,
									Value: int32(m.TotalEntries),
								},
								output.Column{
									Name:  "Completion Percentage",
									Type:  serialization2.TypeFloat32,
									Value: float32(m.CompletionPercentage),
								},
							})
						}
						return ec.AddOutputRows(ctx, rows...)
					}
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
