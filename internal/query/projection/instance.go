package projection

import (
	"context"

	"github.com/zitadel/zitadel/internal/errors"
	"github.com/zitadel/zitadel/internal/eventstore"
	"github.com/zitadel/zitadel/internal/eventstore/handler"
	"github.com/zitadel/zitadel/internal/eventstore/handler/crdb"
	"github.com/zitadel/zitadel/internal/repository/instance"
)

const (
	InstanceProjectionTable = "projections.instances"

	InstanceColumnID              = "id"
	InstanceColumnName            = "name"
	InstanceColumnChangeDate      = "change_date"
	InstanceColumnCreationDate    = "creation_date"
	InstanceColumnGlobalOrgID     = "global_org_id"
	InstanceColumnProjectID       = "iam_project_id"
	InstanceColumnConsoleID       = "console_client_id"
	InstanceColumnConsoleAppID    = "console_app_id"
	InstanceColumnSequence        = "sequence"
	InstanceColumnDefaultLanguage = "default_language"
)

type InstanceProjection struct {
	crdb.StatementHandler
}

func NewInstanceProjection(ctx context.Context, config crdb.StatementHandlerConfig) *InstanceProjection {
	p := new(InstanceProjection)
	config.ProjectionName = InstanceProjectionTable
	config.Reducers = p.reducers()
	config.InitCheck = crdb.NewTableCheck(
		crdb.NewTable([]*crdb.Column{
			crdb.NewColumn(InstanceColumnID, crdb.ColumnTypeText),
			crdb.NewColumn(InstanceColumnName, crdb.ColumnTypeText, crdb.Default("")),
			crdb.NewColumn(InstanceColumnChangeDate, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(InstanceColumnCreationDate, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(InstanceColumnGlobalOrgID, crdb.ColumnTypeText, crdb.Default("")),
			crdb.NewColumn(InstanceColumnProjectID, crdb.ColumnTypeText, crdb.Default("")),
			crdb.NewColumn(InstanceColumnConsoleID, crdb.ColumnTypeText, crdb.Default("")),
			crdb.NewColumn(InstanceColumnConsoleAppID, crdb.ColumnTypeText, crdb.Default("")),
			crdb.NewColumn(InstanceColumnSequence, crdb.ColumnTypeInt64),
			crdb.NewColumn(InstanceColumnDefaultLanguage, crdb.ColumnTypeText, crdb.Default("")),
		},
			crdb.NewPrimaryKey(InstanceColumnID),
		),
	)
	p.StatementHandler = crdb.NewStatementHandler(ctx, config)
	return p
}

func (p *InstanceProjection) reducers() []handler.AggregateReducer {
	return []handler.AggregateReducer{
		{
			Aggregate: instance.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  instance.InstanceAddedEventType,
					Reduce: p.reduceInstanceAdded,
				},
				{
					Event:  instance.GlobalOrgSetEventType,
					Reduce: p.reduceGlobalOrgSet,
				},
				{
					Event:  instance.ProjectSetEventType,
					Reduce: p.reduceIAMProjectSet,
				},
				{
					Event:  instance.ConsoleSetEventType,
					Reduce: p.reduceConsoleSet,
				},
				{
					Event:  instance.DefaultLanguageSetEventType,
					Reduce: p.reduceDefaultLanguageSet,
				},
			},
		},
	}
}

func (p *InstanceProjection) reduceInstanceAdded(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*instance.InstanceAddedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-29nlS", "reduce.wrong.event.type %s", instance.InstanceAddedEventType)
	}
	return crdb.NewCreateStatement(
		e,
		[]handler.Column{
			handler.NewCol(InstanceColumnID, e.Aggregate().InstanceID),
			handler.NewCol(InstanceColumnCreationDate, e.CreationDate()),
			handler.NewCol(InstanceColumnChangeDate, e.CreationDate()),
			handler.NewCol(InstanceColumnSequence, e.Sequence()),
			handler.NewCol(InstanceColumnName, e.Name),
		},
	), nil
}

func (p *InstanceProjection) reduceGlobalOrgSet(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*instance.GlobalOrgSetEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-2n9f2", "reduce.wrong.event.type %s", instance.GlobalOrgSetEventType)
	}
	return crdb.NewUpdateStatement(
		e,
		[]handler.Column{
			handler.NewCol(InstanceColumnChangeDate, e.CreationDate()),
			handler.NewCol(InstanceColumnSequence, e.Sequence()),
			handler.NewCol(InstanceColumnGlobalOrgID, e.OrgID),
		},
		[]handler.Condition{
			handler.NewCond(InstanceColumnID, e.Aggregate().InstanceID),
		},
	), nil
}

func (p *InstanceProjection) reduceIAMProjectSet(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*instance.ProjectSetEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-30o0e", "reduce.wrong.event.type %s", instance.ProjectSetEventType)
	}
	return crdb.NewUpdateStatement(
		e,
		[]handler.Column{
			handler.NewCol(InstanceColumnChangeDate, e.CreationDate()),
			handler.NewCol(InstanceColumnSequence, e.Sequence()),
			handler.NewCol(InstanceColumnProjectID, e.ProjectID),
		},
		[]handler.Condition{
			handler.NewCond(InstanceColumnID, e.Aggregate().InstanceID),
		},
	), nil
}

func (p *InstanceProjection) reduceConsoleSet(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*instance.ConsoleSetEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-Dgf11", "reduce.wrong.event.type %s", instance.ConsoleSetEventType)
	}
	return crdb.NewUpdateStatement(
		e,
		[]handler.Column{
			handler.NewCol(InstanceColumnChangeDate, e.CreationDate()),
			handler.NewCol(InstanceColumnSequence, e.Sequence()),
			handler.NewCol(InstanceColumnConsoleID, e.ClientID),
			handler.NewCol(InstanceColumnConsoleAppID, e.AppID),
		},
		[]handler.Condition{
			handler.NewCond(InstanceColumnID, e.Aggregate().InstanceID),
		},
	), nil
}

func (p *InstanceProjection) reduceDefaultLanguageSet(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*instance.DefaultLanguageSetEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "HANDL-30o0e", "reduce.wrong.event.type %s", instance.DefaultLanguageSetEventType)
	}
	return crdb.NewUpdateStatement(
		e,
		[]handler.Column{
			handler.NewCol(InstanceColumnChangeDate, e.CreationDate()),
			handler.NewCol(InstanceColumnSequence, e.Sequence()),
			handler.NewCol(InstanceColumnDefaultLanguage, e.Language.String()),
		},
		[]handler.Condition{
			handler.NewCond(InstanceColumnID, e.Aggregate().InstanceID),
		},
	), nil
}
