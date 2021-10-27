package restore

import (
	"github.com/caos/orbos/pkg/labels"
	"github.com/caos/zitadel/operator/common"
	"github.com/caos/zitadel/operator/helpers"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getJob(
	namespace string,
	nameLabels *labels.Name,
	nodeselector map[string]string,
	tolerations []corev1.Toleration,
	secretName string,
	secretKey string,
	command string,
	image string,
	runAsUser int64,
) *batchv1.Job {
	return &batchv1.Job{
		ObjectMeta: v1.ObjectMeta{
			Name:      nameLabels.Name(),
			Namespace: namespace,
			Labels:    labels.MustK8sMap(nameLabels),
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					NodeSelector:  nodeselector,
					Tolerations:   tolerations,
					RestartPolicy: corev1.RestartPolicyNever,
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
						Name:  nameLabels.Name(),
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
							Name:      secretKey,
							SubPath:   secretKey,
							MountPath: secretPath,
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
						Name: secretKey,
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName:  secretName,
								DefaultMode: helpers.PointerInt32(0444),
							},
						},
					}, {
						Name: dbSecrets,
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					}},
				},
			},
		},
	}
}
