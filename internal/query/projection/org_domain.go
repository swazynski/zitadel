package projection

import (
	"context"

	"github.com/zitadel/zitadel/internal/domain"
	"github.com/zitadel/zitadel/internal/errors"
	"github.com/zitadel/zitadel/internal/eventstore"
	"github.com/zitadel/zitadel/internal/eventstore/handler"
	"github.com/zitadel/zitadel/internal/eventstore/handler/crdb"
	"github.com/zitadel/zitadel/internal/repository/org"
)

const (
	OrgDomainTable = "projections.org_domains"

	OrgDomainOrgIDCol          = "org_id"
	OrgDomainInstanceIDCol     = "instance_id"
	OrgDomainCreationDateCol   = "creation_date"
	OrgDomainChangeDateCol     = "change_date"
	OrgDomainSequenceCol       = "sequence"
	OrgDomainDomainCol         = "domain"
	OrgDomainIsVerifiedCol     = "is_verified"
	OrgDomainIsPrimaryCol      = "is_primary"
	OrgDomainValidationTypeCol = "validation_type"
)

type OrgDomainProjection struct {
	crdb.StatementHandler
}

func NewOrgDomainProjection(ctx context.Context, config crdb.StatementHandlerConfig) *OrgDomainProjection {
	p := new(OrgDomainProjection)
	config.ProjectionName = OrgDomainTable
	config.Reducers = p.reducers()
	config.InitCheck = crdb.NewTableCheck(
		crdb.NewTable([]*crdb.Column{
			crdb.NewColumn(OrgDomainOrgIDCol, crdb.ColumnTypeText),
			crdb.NewColumn(OrgDomainInstanceIDCol, crdb.ColumnTypeText),
			crdb.NewColumn(OrgDomainCreationDateCol, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(OrgDomainChangeDateCol, crdb.ColumnTypeTimestamp),
			crdb.NewColumn(OrgDomainSequenceCol, crdb.ColumnTypeInt64),
			crdb.NewColumn(OrgDomainDomainCol, crdb.ColumnTypeText),
			crdb.NewColumn(OrgDomainIsVerifiedCol, crdb.ColumnTypeBool),
			crdb.NewColumn(OrgDomainIsPrimaryCol, crdb.ColumnTypeBool),
			crdb.NewColumn(OrgDomainValidationTypeCol, crdb.ColumnTypeEnum),
		},
			crdb.NewPrimaryKey(OrgDomainInstanceIDCol, OrgDomainOrgIDCol),
		),
	)
	p.StatementHandler = crdb.NewStatementHandler(ctx, config)
	return p
}

func (p *OrgDomainProjection) reducers() []handler.AggregateReducer {
	return []handler.AggregateReducer{
		{
			Aggregate: org.AggregateType,
			EventRedusers: []handler.EventReducer{
				{
					Event:  org.OrgDomainAddedEventType,
					Reduce: p.reduceDomainAdded,
				},
				{
					Event:  org.OrgDomainVerificationAddedEventType,
					Reduce: p.reduceDomainVerificationAdded,
				},
				{
					Event:  org.OrgDomainVerifiedEventType,
					Reduce: p.reduceDomainVerified,
				},
				{
					Event:  org.OrgDomainPrimarySetEventType,
					Reduce: p.reducePrimaryDomainSet,
				},
				{
					Event:  org.OrgDomainRemovedEventType,
					Reduce: p.reduceDomainRemoved,
				},
			},
		},
	}
}

func (p *OrgDomainProjection) reduceDomainAdded(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*org.DomainAddedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-DM2DI", "reduce.wrong.event.type %s", org.OrgDomainAddedEventType)
	}
	return crdb.NewCreateStatement(
		e,
		[]handler.Column{
			handler.NewCol(OrgDomainCreationDateCol, e.CreationDate()),
			handler.NewCol(OrgDomainChangeDateCol, e.CreationDate()),
			handler.NewCol(OrgDomainSequenceCol, e.Sequence()),
			handler.NewCol(OrgDomainDomainCol, e.Domain),
			handler.NewCol(OrgDomainOrgIDCol, e.Aggregate().ID),
			handler.NewCol(OrgDomainInstanceIDCol, e.Aggregate().InstanceID),
			handler.NewCol(OrgDomainIsVerifiedCol, false),
			handler.NewCol(OrgDomainIsPrimaryCol, false),
			handler.NewCol(OrgDomainValidationTypeCol, domain.OrgDomainValidationTypeUnspecified),
		},
	), nil
}

func (p *OrgDomainProjection) reduceDomainVerificationAdded(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*org.DomainVerificationAddedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-EBzyu", "reduce.wrong.event.type %s", org.OrgDomainVerificationAddedEventType)
	}
	return crdb.NewUpdateStatement(
		e,
		[]handler.Column{
			handler.NewCol(OrgDomainChangeDateCol, e.CreationDate()),
			handler.NewCol(OrgDomainSequenceCol, e.Sequence()),
			handler.NewCol(OrgDomainValidationTypeCol, e.ValidationType),
		},
		[]handler.Condition{
			handler.NewCond(OrgDomainDomainCol, e.Domain),
			handler.NewCond(OrgDomainOrgIDCol, e.Aggregate().ID),
		},
	), nil
}

func (p *OrgDomainProjection) reduceDomainVerified(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*org.DomainVerifiedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-3Rvkr", "reduce.wrong.event.type %s", org.OrgDomainVerifiedEventType)
	}
	return crdb.NewUpdateStatement(
		e,
		[]handler.Column{
			handler.NewCol(OrgDomainChangeDateCol, e.CreationDate()),
			handler.NewCol(OrgDomainSequenceCol, e.Sequence()),
			handler.NewCol(OrgDomainIsVerifiedCol, true),
		},
		[]handler.Condition{
			handler.NewCond(OrgDomainDomainCol, e.Domain),
			handler.NewCond(OrgDomainOrgIDCol, e.Aggregate().ID),
		},
	), nil
}

func (p *OrgDomainProjection) reducePrimaryDomainSet(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*org.DomainPrimarySetEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-aIuei", "reduce.wrong.event.type %s", org.OrgDomainPrimarySetEventType)
	}
	return crdb.NewMultiStatement(
		e,
		crdb.AddUpdateStatement(
			[]handler.Column{
				handler.NewCol(OrgDomainChangeDateCol, e.CreationDate()),
				handler.NewCol(OrgDomainSequenceCol, e.Sequence()),
				handler.NewCol(OrgDomainIsPrimaryCol, false),
			},
			[]handler.Condition{
				handler.NewCond(OrgDomainOrgIDCol, e.Aggregate().ID),
				handler.NewCond(OrgDomainIsPrimaryCol, true),
			},
		),
		crdb.AddUpdateStatement(
			[]handler.Column{
				handler.NewCol(OrgDomainChangeDateCol, e.CreationDate()),
				handler.NewCol(OrgDomainSequenceCol, e.Sequence()),
				handler.NewCol(OrgDomainIsPrimaryCol, true),
			},
			[]handler.Condition{
				handler.NewCond(OrgDomainDomainCol, e.Domain),
				handler.NewCond(OrgDomainOrgIDCol, e.Aggregate().ID),
			},
		),
	), nil
}

func (p *OrgDomainProjection) reduceDomainRemoved(event eventstore.Event) (*handler.Statement, error) {
	e, ok := event.(*org.DomainRemovedEvent)
	if !ok {
		return nil, errors.ThrowInvalidArgumentf(nil, "PROJE-gh1Mx", "reduce.wrong.event.type %s", org.OrgDomainRemovedEventType)
	}
	return crdb.NewDeleteStatement(
		e,
		[]handler.Condition{
			handler.NewCond(OrgDomainDomainCol, e.Domain),
			handler.NewCond(OrgDomainOrgIDCol, e.Aggregate().ID),
		},
	), nil
}
