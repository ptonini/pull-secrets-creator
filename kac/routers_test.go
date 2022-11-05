package kac

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	//configMapsGVR =
	pod = corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod"},
		Spec: corev1.PodSpec{
			ImagePullSecrets: []corev1.LocalObjectReference{
				{Name: "test-credentials"},
			},
		},
	}
)

func admissionReviewFactory(gvr metav1.GroupVersionResource, obj interface{}) string {
	rawObject, _ := json.Marshal(obj)
	a, err := json.Marshal(admissionv1.AdmissionReview{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AdmissionReview",
			APIVersion: "admission.k8s.io/v1",
		},
		Request: &admissionv1.AdmissionRequest{
			Namespace: "default",
			Resource:  gvr,
			Object: runtime.RawExtension{
				Raw: rawObject,
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	return string(a)
}

func fakeRequest(ctx context.Context, r *gin.Engine, method string, route string, rawBody string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(method, route, strings.NewReader(rawBody))
	req = req.WithContext(ctx)
	r.ServeHTTP(w, req)
	return w
}

func Test_HealthcheckRoute(t *testing.T) {
	router := NewRouter()
	w := fakeRequest(context.Background(), router, http.MethodGet, "/health", "")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, `{"status":"ok"}`, w.Body.String())
}

func Test_ReviewerRoutes(t *testing.T) {

	ctx := context.Background()
	router := NewRouter()
	_ = readConfig("../config.yaml")

	for _, route := range []string{"/mutate", "/validate"} {
		t.Run("route "+route+" with nil body", func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, route, nil)
			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
		t.Run("route "+route+" with empty body", func(t *testing.T) {
			w := fakeRequest(ctx, router, http.MethodPost, route, "")
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
		t.Run("route "+route+" with invalid body", func(t *testing.T) {
			invalidBody, _ := json.Marshal(pod)
			w := fakeRequest(ctx, router, http.MethodPost, route, string(invalidBody))
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}

	t.Run("route /validate with invalid admission request resource", func(t *testing.T) {
		body := admissionReviewFactory(metav1.GroupVersionResource{Version: "v1", Resource: "ConfigMaps"}, &corev1.ConfigMap{})
		w := fakeRequest(ctx, router, http.MethodPost, "/validate", body)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("route /validate with invalid admission request resource kind", func(t *testing.T) {
		body := admissionReviewFactory(podsGVR, &corev1.ConfigMap{})
		w := fakeRequest(ctx, router, http.MethodPost, "/validate", body)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	ctx = context.WithValue(ctx, keyFakeClientSet, true)
	t.Run("route /validate with valid request", func(t *testing.T) {
		body := admissionReviewFactory(podsGVR, pod)
		w := fakeRequest(ctx, router, http.MethodPost, "/validate", body)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	ctx = context.WithValue(ctx, keyFakeClientSet, true)
	t.Run("route /mutate with valid request", func(t *testing.T) {
		body := admissionReviewFactory(podsGVR, pod)
		w := fakeRequest(ctx, router, http.MethodPost, "/mutate", body)
		assert.Equal(t, http.StatusOK, w.Code)
	})

}
