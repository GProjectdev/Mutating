package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"

    admissionv1 "k8s.io/api/admission/v1"
    corev1 "k8s.io/api/core/v1"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1" // 🔧 추가
)

func main() {
    http.HandleFunc("/mutate", handleMutate)
    fmt.Println("Starting webhook server on :8443")
    if err := http.ListenAndServeTLS(":8443", "/tls/tls.crt", "/tls/tls.key", nil); err != nil {
        panic(err)
    }
}

func handleMutate(w http.ResponseWriter, r *http.Request) {
    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        http.Error(w, "could not read request", http.StatusBadRequest)
        return
    }

    var admissionReview admissionv1.AdmissionReview
    if err := json.Unmarshal(body, &admissionReview); err != nil {
        http.Error(w, "could not parse admission review", http.StatusBadRequest)
        return
    }

    pod := corev1.Pod{}
    if err := json.Unmarshal(admissionReview.Request.Object.Raw, &pod); err != nil {
        http.Error(w, "could not parse pod object", http.StatusBadRequest)
        return
    }

    var patches []map[string]interface{}
    if pod.Labels["change"] == "image" {
        for i, container := range pod.Spec.Containers {
            newImage := fmt.Sprintf("jeongseungjun/criu-test:%s", container.Image)
            patch := map[string]interface{}{
                "op":    "replace",
                "path":  fmt.Sprintf("/spec/containers/%d/image", i),
                "value": newImage,
            }
            patches = append(patches, patch)
        }
    }

    patchBytes, _ := json.Marshal(patches)

    response := admissionv1.AdmissionReview{
        TypeMeta: metav1.TypeMeta{ // 🔧 이 부분 추가
            APIVersion: "admission.k8s.io/v1",
            Kind:       "AdmissionReview",
        },
        Response: &admissionv1.AdmissionResponse{
            UID:     admissionReview.Request.UID,
            Allowed: true,
            Patch:   patchBytes,
            PatchType: func() *admissionv1.PatchType {
                pt := admissionv1.PatchTypeJSONPatch
                return &pt
            }(),
        },
    }

    // 🔧 Encoder 사용 (더 안정적)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}

