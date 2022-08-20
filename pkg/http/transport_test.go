package http

import (
	"bytes"
	"crypto/tls"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/gitops-tools/backstage-tenants-controller/test"
	"github.com/google/go-cmp/cmp"
)

func TestTransportFromSecret(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"testing": "value"}`)
	}))
	defer ts.Close()
	secret := &corev1.Secret{Data: secretDataFromTLSConfig(t, ts.TLS)}

	transport, err := TransportFromSecret(secret)
	if err != nil {
		t.Fatal(err)
	}

	client := http.Client{Transport: transport}
	resp, err := client.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff("{\"testing\": \"value\"}\n", string(b)); diff != "" {
		t.Fatalf("didn't get response: %s", diff)
	}
}

func TestTransportFromSecret_no_CA(t *testing.T) {
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"testing": "value"}`)
	}))
	defer ts.Close()
	secret := &corev1.Secret{Data: map[string][]byte{}}

	transport, err := TransportFromSecret(secret)
	if err != nil {
		t.Fatal(err)
	}

	client := http.Client{Transport: transport}
	_, err = client.Get(ts.URL)
	test.AssertErrorMatch(t, "certificate is not trusted", err)
}

func secretDataFromTLSConfig(t *testing.T, c *tls.Config) map[string][]byte {
	data := map[string][]byte{}
	b := bytes.Buffer{}
	if err := pem.Encode(&b, &pem.Block{Type: "CERTIFICATE", Bytes: c.Certificates[0].Certificate[0]}); err != nil {
		t.Fatal(err)
	}
	data[CACertKey] = b.Bytes()

	return data
}
