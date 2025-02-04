package command

import (
	"context"

	"github.com/zitadel/zitadel/internal/domain"
	caos_errs "github.com/zitadel/zitadel/internal/errors"
	"github.com/zitadel/zitadel/internal/eventstore"
	"github.com/zitadel/zitadel/internal/repository/instance"
	"github.com/zitadel/zitadel/internal/telemetry/tracing"
)

func (c *Commands) AddDefaultPasswordAgePolicy(ctx context.Context, policy *domain.PasswordAgePolicy) (*domain.PasswordAgePolicy, error) {
	addedPolicy := NewInstancePasswordAgePolicyWriteModel(ctx)
	instanceAgg := InstanceAggregateFromWriteModel(&addedPolicy.WriteModel)
	event, err := c.addDefaultPasswordAgePolicy(ctx, instanceAgg, addedPolicy, policy)
	if err != nil {
		return nil, err
	}

	pushedEvents, err := c.eventstore.Push(ctx, event)
	if err != nil {
		return nil, err
	}
	err = AppendAndReduce(addedPolicy, pushedEvents...)
	if err != nil {
		return nil, err
	}
	return writeModelToPasswordAgePolicy(&addedPolicy.PasswordAgePolicyWriteModel), nil
}

func (c *Commands) addDefaultPasswordAgePolicy(ctx context.Context, instanceAgg *eventstore.Aggregate, addedPolicy *InstancePasswordAgePolicyWriteModel, policy *domain.PasswordAgePolicy) (eventstore.Command, error) {
	err := c.eventstore.FilterToQueryReducer(ctx, addedPolicy)
	if err != nil {
		return nil, err
	}
	if addedPolicy.State == domain.PolicyStateActive {
		return nil, caos_errs.ThrowAlreadyExists(nil, "INSTANCE-Lk0dS", "Errors.IAM.PasswordAgePolicy.AlreadyExists")
	}

	return instance.NewPasswordAgePolicyAddedEvent(ctx, instanceAgg, policy.ExpireWarnDays, policy.MaxAgeDays), nil

}

func (c *Commands) ChangeDefaultPasswordAgePolicy(ctx context.Context, policy *domain.PasswordAgePolicy) (*domain.PasswordAgePolicy, error) {
	existingPolicy, err := c.defaultPasswordAgePolicyWriteModelByID(ctx)
	if err != nil {
		return nil, err
	}
	if existingPolicy.State == domain.PolicyStateUnspecified || existingPolicy.State == domain.PolicyStateRemoved {
		return nil, caos_errs.ThrowNotFound(nil, "INSTANCE-0oPew", "Errors.IAM.PasswordAgePolicy.NotFound")
	}

	instanceAgg := InstanceAggregateFromWriteModel(&existingPolicy.PasswordAgePolicyWriteModel.WriteModel)
	changedEvent, hasChanged := existingPolicy.NewChangedEvent(ctx, instanceAgg, policy.ExpireWarnDays, policy.MaxAgeDays)
	if !hasChanged {
		return nil, caos_errs.ThrowPreconditionFailed(nil, "INSTANCE-180sf", "Errors.IAM.PasswordAgePolicy.NotChanged")
	}

	pushedEvents, err := c.eventstore.Push(ctx, changedEvent)
	if err != nil {
		return nil, err
	}
	err = AppendAndReduce(existingPolicy, pushedEvents...)
	if err != nil {
		return nil, err
	}

	return writeModelToPasswordAgePolicy(&existingPolicy.PasswordAgePolicyWriteModel), nil
}

func (c *Commands) defaultPasswordAgePolicyWriteModelByID(ctx context.Context) (policy *InstancePasswordAgePolicyWriteModel, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	writeModel := NewInstancePasswordAgePolicyWriteModel(ctx)
	err = c.eventstore.FilterToQueryReducer(ctx, writeModel)
	if err != nil {
		return nil, err
	}
	return writeModel, nil
}
