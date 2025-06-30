package main

import (
        "encoding/json"
        "io/ioutil"
        "log"
        "net/http"
        "os"

        admissionv1 "k8s.io/api/admission/v1"
        corev1 "k8s.io/api/core/v1"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func main() {
        certFile := os.Getenv("TLS_CERT_FILE")
        keyFile := os.Getenv("TLS_KEY_FILE")

        http.HandleFunc("/mutate", handleMutate)
        log.Println("Starting webhook server on port 8443...")
        err := http.ListenAndServeTLS(":8443", certFile, keyFile, nil)
        if err != nil {
                log.Fatal(err)
        }
}

func handleMutate(w http.ResponseWriter, r *http.Request) {
        body, err := ioutil.ReadAll(r.Body)
        if err != nil {
                http.Error(w, "could not read request", http.StatusBadRequest)
                return
        }

        var review admissionv1.AdmissionReview
        if err := json.Unmarshal(body, &review); err != nil {
                http.Error(w, "could not parse admission review", http.StatusBadRequest)
                return
        }

        pod := corev1.Pod{}
        if err := json.Unmarshal(review.Request.Object.Raw, &pod); err != nil {
                http.Error(w, "could not parse pod object", http.StatusBadRequest)
                return
        }

        var patch []map[string]interface{}

        // labels가 없으면 전체 labels map을 추가
        if pod.ObjectMeta.Labels == nil {
                patch = append(patch, map[string]interface{}{
                        "op":    "add",
                        "path":  "/metadata/labels",
                        "value": map[string]string{"injected": "true"},
                })
        } else {
                patch = append(patch, map[string]interface{}{
                        "op":    "add",
                        "path":  "/metadata/labels/injected",
                        "value": "true",
                })
        }

        patchBytes, err := json.Marshal(patch)
        if err != nil {
                http.Error(w, "could not marshal patch", http.StatusInternalServerError)
                return
        }

        reviewResponse := admissionv1.AdmissionReview{
                TypeMeta: metav1.TypeMeta{
                        APIVersion: "admission.k8s.io/v1",
                        Kind:       "AdmissionReview",
                },
                Response: &admissionv1.AdmissionResponse{
                        UID:       review.Request.UID,
                        Allowed:   true,
                        Patch:     patchBytes,
                        PatchType: func() *admissionv1.PatchType {
                                pt := admissionv1.PatchTypeJSONPatch
                                return &pt
                        }(),
                },
        }

        respBytes, err := json.Marshal(reviewResponse)
        if err != nil {
                http.Error(w, "could not marshal response", http.StatusInternalServerError)
                return
        }

        w.Header().Set("Content-Type", "application/json")
        w.Write(respBytes)
}
