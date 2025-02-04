package command

import (
	"context"

	"github.com/zitadel/zitadel/internal/domain"
	caos_errs "github.com/zitadel/zitadel/internal/errors"
	"github.com/zitadel/zitadel/internal/eventstore"
	"github.com/zitadel/zitadel/internal/repository/instance"
	"github.com/zitadel/zitadel/internal/telemetry/tracing"
)

func (c *Commands) AddDefaultLockoutPolicy(ctx context.Context, policy *domain.LockoutPolicy) (*domain.LockoutPolicy, error) {
	addedPolicy := NewInstanceLockoutPolicyWriteModel(ctx)
	instanceAgg := InstanceAggregateFromWriteModel(&addedPolicy.WriteModel)
	event, err := c.addDefaultLockoutPolicy(ctx, instanceAgg, addedPolicy, policy)
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

	return writeModelToLockoutPolicy(&addedPolicy.LockoutPolicyWriteModel), nil
}

func (c *Commands) addDefaultLockoutPolicy(ctx context.Context, instanceAgg *eventstore.Aggregate, addedPolicy *InstanceLockoutPolicyWriteModel, policy *domain.LockoutPolicy) (eventstore.Command, error) {
	err := c.eventstore.FilterToQueryReducer(ctx, addedPolicy)
	if err != nil {
		return nil, err
	}
	if addedPolicy.State == domain.PolicyStateActive {
		return nil, caos_errs.ThrowAlreadyExists(nil, "INSTANCE-0olDf", "Errors.IAM.LockoutPolicy.AlreadyExists")
	}

	return instance.NewLockoutPolicyAddedEvent(ctx, instanceAgg, policy.MaxPasswordAttempts, policy.ShowLockOutFailures), nil
}

func (c *Commands) ChangeDefaultLockoutPolicy(ctx context.Context, policy *domain.LockoutPolicy) (*domain.LockoutPolicy, error) {
	existingPolicy, err := c.defaultLockoutPolicyWriteModelByID(ctx)
	if err != nil {
		return nil, err
	}
	if existingPolicy.State == domain.PolicyStateUnspecified || existingPolicy.State == domain.PolicyStateRemoved {
		return nil, caos_errs.ThrowNotFound(nil, "INSTANCE-0oPew", "Errors.IAM.LockoutPolicy.NotFound")
	}

	instanceAgg := InstanceAggregateFromWriteModel(&existingPolicy.LockoutPolicyWriteModel.WriteModel)
	changedEvent, hasChanged := existingPolicy.NewChangedEvent(ctx, instanceAgg, policy.MaxPasswordAttempts, policy.ShowLockOutFailures)
	if !hasChanged {
		return nil, caos_errs.ThrowPreconditionFailed(nil, "INSTANCE-0psjF", "Errors.IAM.LockoutPolicy.NotChanged")
	}

	pushedEvents, err := c.eventstore.Push(ctx, changedEvent)
	if err != nil {
		return nil, err
	}
	err = AppendAndReduce(existingPolicy, pushedEvents...)
	if err != nil {
		return nil, err
	}
	return writeModelToLockoutPolicy(&existingPolicy.LockoutPolicyWriteModel), nil
}

func (c *Commands) defaultLockoutPolicyWriteModelByID(ctx context.Context) (policy *InstanceLockoutPolicyWriteModel, err error) {
	ctx, span := tracing.NewSpan(ctx)
	defer func() { span.EndWithError(err) }()

	writeModel := NewInstanceLockoutPolicyWriteModel(ctx)
	err = c.eventstore.FilterToQueryReducer(ctx, writeModel)
	if err != nil {
		return nil, err
	}
	return writeModel, nil
}
