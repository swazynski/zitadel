package projection

import (
	"context"
	"database/sql"
	"time"

	"github.com/zitadel/zitadel/internal/crypto"
	"github.com/zitadel/zitadel/internal/eventstore"
	"github.com/zitadel/zitadel/internal/eventstore/handler"
	"github.com/zitadel/zitadel/internal/eventstore/handler/crdb"
)

const (
	CurrentSeqTable   = "projections.current_sequences"
	LocksTable        = "projections.locks"
	FailedEventsTable = "projections.failed_events"
)

func Start(ctx context.Context, sqlClient *sql.DB, es *eventstore.Eventstore, config Config, keyEncryptionAlgorithm crypto.EncryptionAlgorithm) error {
	projectionConfig := crdb.StatementHandlerConfig{
		ProjectionHandlerConfig: handler.ProjectionHandlerConfig{
			HandlerConfig: handler.HandlerConfig{
				Eventstore: es,
			},
			RequeueEvery:     config.RequeueEvery,
			RetryFailedAfter: config.RetryFailedAfter,
		},
		Client:            sqlClient,
		SequenceTable:     CurrentSeqTable,
		LockTable:         LocksTable,
		FailedEventsTable: FailedEventsTable,
		MaxFailureCount:   config.MaxFailureCount,
		BulkLimit:         config.BulkLimit,
	}

	NewOrgProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["orgs"]))
	NewActionProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["actions"]))
	NewFlowProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["flows"]))
	NewProjectProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["projects"]))
	NewPasswordComplexityProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["password_complexities"]))
	NewPasswordAgeProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["password_age_policy"]))
	NewLockoutPolicyProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["lockout_policy"]))
	NewPrivacyPolicyProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["privacy_policy"]))
	NewDomainPolicyProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["org_iam_policy"]))
	NewLabelPolicyProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["label_policy"]))
	NewProjectGrantProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["project_grants"]))
	NewProjectRoleProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["project_roles"]))
	NewOrgDomainProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["org_domains"]))
	NewLoginPolicyProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["login_policies"]))
	NewIDPProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["idps"]))
	NewAppProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["apps"]))
	NewIDPUserLinkProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["idp_user_links"]))
	NewIDPLoginPolicyLinkProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["idp_login_policy_links"]))
	NewMailTemplateProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["mail_templates"]))
	NewMessageTextProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["message_texts"]))
	NewCustomTextProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["custom_texts"]))
	NewUserProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["users"]))
	NewLoginNameProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["login_names"]))
	NewOrgMemberProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["org_members"]))
	NewInstanceDomainProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["instance_domains"]))
	NewInstanceMemberProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["iam_members"]))
	NewProjectMemberProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["project_members"]))
	NewProjectGrantMemberProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["project_grant_members"]))
	NewAuthNKeyProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["authn_keys"]))
	NewPersonalAccessTokenProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["personal_access_tokens"]))
	NewUserGrantProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["user_grants"]))
	NewUserMetadataProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["user_metadata"]))
	NewUserAuthMethodProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["user_auth_method"]))
	NewInstanceProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["instances"]))
	NewSecretGeneratorProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["secret_generators"]))
	NewSMTPConfigProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["smtp_configs"]))
	NewSMSConfigProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["sms_config"]))
	NewOIDCSettingsProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["oidc_settings"]))
	NewDebugNotificationProviderProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["debug_notification_provider"]))
	NewKeyProjection(ctx, applyCustomConfig(projectionConfig, config.Customizations["keys"]), keyEncryptionAlgorithm)
	return nil
}

func applyCustomConfig(config crdb.StatementHandlerConfig, customConfig CustomConfig) crdb.StatementHandlerConfig {
	if customConfig.BulkLimit != nil {
		config.BulkLimit = *customConfig.BulkLimit
	}
	if customConfig.MaxFailureCount != nil {
		config.MaxFailureCount = *customConfig.MaxFailureCount
	}
	if customConfig.RequeueEvery != nil {
		config.RequeueEvery = *customConfig.RequeueEvery
	}
	if customConfig.RetryFailedAfter != nil {
		config.RetryFailedAfter = *customConfig.RetryFailedAfter
	}

	return config
}

func iteratorPool(workerCount int) chan func() {
	if workerCount <= 0 {
		return nil
	}

	queue := make(chan func())
	for i := 0; i < workerCount; i++ {
		go func() {
			for iteration := range queue {
				iteration()
				time.Sleep(2 * time.Second)
			}
		}()
	}
	return queue
}
