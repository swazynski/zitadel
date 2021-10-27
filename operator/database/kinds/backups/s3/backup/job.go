package backup

import (
	"github.com/caos/orbos/pkg/labels"
	"github.com/caos/zitadel/operator/common"
	"github.com/caos/zitadel/operator/helpers"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getCronJob(
	namespace string,
	nameLabels *labels.Name,
	cron string,
	jobSpecDef batchv1.JobSpec,
) *v1beta1.CronJob {
	return &v1beta1.CronJob{
		ObjectMeta: v1.ObjectMeta{
			Name:      nameLabels.Name(),
			Namespace: namespace,
			Labels:    labels.MustK8sMap(nameLabels),
		},
		Spec: v1beta1.CronJobSpec{
			Schedule:          cron,
			ConcurrencyPolicy: v1beta1.ForbidConcurrent,
			JobTemplate: v1beta1.JobTemplateSpec{
				Spec: jobSpecDef,
			},
		},
	}
}

func getJob(
	namespace string,
	nameLabels *labels.Name,
	jobSpecDef batchv1.JobSpec,
) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      nameLabels.Name(),
			Namespace: namespace,
			Labels:    labels.MustK8sMap(nameLabels),
		},
		Spec: jobSpecDef,
	}
}

func getJobSpecDef(
	nodeselector map[string]string,
	tolerations []corev1.Toleration,
	accessKeyIDName string,
	accessKeyIDKey string,
	secretAccessKeyName string,
	secretAccessKeyKey string,
	sessionTokenName string,
	sessionTokenKey string,
	backupName string,
	command string,
	image string,
	runAsUser int64,
) batchv1.JobSpec {
	return batchv1.JobSpec{
		Template: corev1.PodTemplateSpec{
			Spec: corev1.PodSpec{
				RestartPolicy: corev1.RestartPolicyNever,
				NodeSelector:  nodeselector,
				Tolerations:   tolerations,
				InitContainers: []corev1.Container{
					common.GetInitContainer(
						"backup",
						internalSecretName,
						dbSecrets,
						[]string{"root"},
						runAsUser,
						image,
					),
				},
				Containers: []corev1.Container{{
					Name:  backupName,
					Image: image,
					Command: []string{
						"/bin/bash",
						"-c",
						command,
					},
					VolumeMounts: []corev1.VolumeMount{{
						Name:      dbSecrets,
						MountPath: certPath,
					}, {
						Name:      accessKeyIDKey,
						SubPath:   accessKeyIDKey,
						MountPath: accessKeyIDPath,
					}, {
						Name:      secretAccessKeyKey,
						SubPath:   secretAccessKeyKey,
						MountPath: secretAccessKeyPath,
					}, {
						Name:      sessionTokenKey,
						SubPath:   sessionTokenKey,
						MountPath: sessionTokenPath,
					}},
					ImagePullPolicy: corev1.PullIfNotPresent,
				}},
				Volumes: []corev1.Volume{{
					Name: internalSecretName,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  rootSecretName,
							DefaultMode: helpers.PointerInt32(0444),
						},
					},
				}, {
					Name: accessKeyIDKey,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  accessKeyIDName,
							DefaultMode: helpers.PointerInt32(0444),
						},
					},
				}, {
					Name: secretAccessKeyKey,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  secretAccessKeyName,
							DefaultMode: helpers.PointerInt32(0444),
						},
					},
				}, {
					Name: sessionTokenKey,
					VolumeSource: corev1.VolumeSource{
						Secret: &corev1.SecretVolumeSource{
							SecretName:  sessionTokenName,
							DefaultMode: helpers.PointerInt32(0444),
						},
					},
				}, {
					Name: dbSecrets,
					VolumeSource: corev1.VolumeSource{
						EmptyDir: &corev1.EmptyDirVolumeSource{},
					},
				},
				},
			},
		},
	}
}
