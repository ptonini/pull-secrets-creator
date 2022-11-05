/*
 * Kubernetes Admission Controller.
 * Copyright (C) 2022 Pedro Tonini
 * mailto:pedro DOT tonini AT hotmail DOT com
 *
 * Kubernetes Admission Controller is free software; you can redistribute it and/or
 * modify it under the terms of the GNU Lesser General Public
 * License as published by the Free Software Foundation; either
 * version 3 of the License, or (at your option) any later version.
 *
 * Kubernetes Admission Controller is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
 * Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with this program; if not, write to the Free Software Foundation,
 * Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
 */

package kac

import (
	"context"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"log"
)

var (
	podsGVR = metav1.GroupVersionResource{Version: "v1", Resource: "pods"}
	podGVK  = schema.GroupVersionKind{Version: "v1", Kind: "Pod"}
)

func validationReviewer(ctx context.Context, ar admissionv1.AdmissionReview) (*admissionv1.AdmissionResponse, error) {

	config, err := getConfig()
	if err != nil {
		return nil, err
	}

	clientSet, err := getKubernetesClientSet(ctx)
	if err != nil {
		return nil, err
	}

	// Deserialize and copy request object
	namespace := ar.Request.Namespace
	obj, err := validateAndDeserialize(ar, podsGVR, podGVK)
	if err != nil {
		return nil, err
	}
	pod := obj.(*corev1.Pod)

	for _, i := range pod.Spec.ImagePullSecrets {
		if val, ok := config.ImagePullSecret[i.Name]; ok {
			secret, _ := clientSet.CoreV1().Secrets(namespace).Get(ctx, i.Name, metav1.GetOptions{})
			if secret == nil || secret.Name == "" {
				log.Printf("Adding secret %s on %s", i.Name, namespace)
				_, err = clientSet.CoreV1().Secrets(namespace).Create(ctx, &corev1.Secret{
					TypeMeta: metav1.TypeMeta{},
					ObjectMeta: metav1.ObjectMeta{
						Name:      i.Name,
						Namespace: namespace,
					},
					Data: map[string][]byte{
						".dockerconfigjson": val,
					},
					Type: "kubernetes.io/dockerconfigjson",
				}, metav1.CreateOptions{})
				if err != nil {
					return nil, err
				}
			}
		}
	}

	pt := admissionv1.PatchTypeJSONPatch
	return &admissionv1.AdmissionResponse{Allowed: true, PatchType: &pt, Patch: []byte{}}, nil
}

func mutationReviewer(ctx context.Context, ar admissionv1.AdmissionReview) (*admissionv1.AdmissionResponse, error) {
	pt := admissionv1.PatchTypeJSONPatch
	return &admissionv1.AdmissionResponse{Allowed: true, PatchType: &pt, Patch: []byte{}}, nil

}
