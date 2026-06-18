package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestServer returns an httptest server whose handler is `h`, plus a Client
// pointed at it with a fixed token.
func newTestServer(t *testing.T, h http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return srv, NewClient(srv.URL, "test-token")
}

func TestDoSendsBearerAndParsesErrorEnvelope(t *testing.T) {
	var gotAuth string
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"error":{"code":"volume.busy","message":"Volume is busy"}}`))
	})
	_ = srv

	err := client.do(context.Background(), http.MethodGet, "/v1/anything", nil, nil)
	if err == nil {
		t.Fatal("expected an error for a 409 response")
	}
	if gotAuth != "Bearer test-token" {
		t.Fatalf("missing/incorrect auth header: %q", gotAuth)
	}
	if want := "Volume is busy"; !contains(err.Error(), want) {
		t.Fatalf("error should surface the API message %q, got: %v", want, err)
	}
	if !contains(err.Error(), "volume.busy") {
		t.Fatalf("error should surface the API code, got: %v", err)
	}
}

func TestSSHKeyRoundTrip(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/ssh-keys":
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["name"] != "k1" || body["publicKey"] != "ssh-ed25519 AAA" {
				t.Errorf("unexpected create body: %v", body)
			}
			writeJSON(w, 201, SSHKey{ID: "ssh_1", Name: "k1", PublicKey: "ssh-ed25519 AAA", Fingerprint: "SHA256:xyz"})
		case r.Method == http.MethodGet && r.URL.Path == "/v1/ssh-keys":
			writeJSON(w, 200, []SSHKey{{ID: "ssh_1", Name: "k1", Fingerprint: "SHA256:xyz"}})
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/ssh-keys/ssh_1":
			w.WriteHeader(204)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})
	_ = srv
	ctx := context.Background()

	created, err := client.CreateSSHKey(ctx, "k1", "ssh-ed25519 AAA")
	if err != nil || created.ID != "ssh_1" || created.Fingerprint != "SHA256:xyz" {
		t.Fatalf("create: %v / %+v", err, created)
	}
	got, err := client.GetSSHKey(ctx, "ssh_1")
	if err != nil || got == nil || got.ID != "ssh_1" {
		t.Fatalf("get: %v / %+v", err, got)
	}
	if err := client.DeleteSSHKey(ctx, "ssh_1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestGetSSHKeyNotFoundReturnsNil(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, []SSHKey{{ID: "ssh_other"}})
	})
	_ = srv
	got, err := client.GetSSHKey(context.Background(), "ssh_missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil for a missing key, got %+v", got)
	}
}

func TestVolumeResizeAndDelete(t *testing.T) {
	var resized int64
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/volumes":
			writeJSON(w, 201, Volume{ID: "vol_1", Name: "d", RegionCode: "fra", SizeGB: 50, Status: "AVAILABLE"})
		case r.Method == http.MethodPost && r.URL.Path == "/v1/volumes/vol_1/resize":
			var body map[string]int64
			_ = json.NewDecoder(r.Body).Decode(&body)
			resized = body["sizeGB"]
			w.WriteHeader(200)
		case r.Method == http.MethodDelete && r.URL.Path == "/v1/volumes/vol_1":
			w.WriteHeader(202)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})
	_ = srv
	ctx := context.Background()

	vol, err := client.CreateVolume(ctx, "d", "fra", 50)
	if err != nil || vol.ID != "vol_1" {
		t.Fatalf("create: %v / %+v", err, vol)
	}
	if err := client.ResizeVolume(ctx, "vol_1", 100); err != nil || resized != 100 {
		t.Fatalf("resize: %v / sent %d", err, resized)
	}
	if err := client.DeleteVolume(ctx, "vol_1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestNodeCreateOmitsEmptyOptionalsAndGetHandlesGone(t *testing.T) {
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/v1/nodes":
			raw := map[string]any{}
			_ = json.NewDecoder(r.Body).Decode(&raw)
			// omitempty: no sku_id / cloud_init / ssh keys were set.
			if _, ok := raw["skuId"]; ok {
				t.Errorf("skuId should be omitted when empty")
			}
			if _, ok := raw["cloudInit"]; ok {
				t.Errorf("cloudInit should be omitted when empty")
			}
			writeJSON(w, 202, Node{ID: "node_1", Hostname: "web", Status: "PROVISIONING"})
		case r.URL.Path == "/v1/nodes/node_gone":
			writeJSON(w, 404, map[string]any{"error": map[string]string{"code": "node.not_found", "message": "Node not found"}})
		case r.URL.Path == "/v1/nodes/node_destroyed":
			writeJSON(w, 200, Node{ID: "node_destroyed", Status: "DESTROYED"})
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
	})
	_ = srv
	ctx := context.Background()

	node, err := client.CreateNode(ctx, NodeCreate{Hostname: "web", RegionCode: "fra", ImageID: "ubuntu-24", CPU: 2, RAMGB: 4, DiskGB: 80, BillingMode: "HOURLY"})
	if err != nil || node.ID != "node_1" || node.Status != "PROVISIONING" {
		t.Fatalf("create: %v / %+v", err, node)
	}

	// A 404 and a DESTROYED node both read back as nil (gone).
	if got, err := client.GetNode(ctx, "node_gone"); err != nil || got != nil {
		t.Fatalf("expected nil for 404, got %+v err %v", got, err)
	}
	if got, err := client.GetNode(ctx, "node_destroyed"); err != nil || got != nil {
		t.Fatalf("expected nil for DESTROYED, got %+v err %v", got, err)
	}
}

func TestListPlansPassesRegionFilter(t *testing.T) {
	var gotQuery string
	srv, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		writeJSON(w, 200, []Plan{{ID: "vps-std-1-2", Family: "STANDARD", CPU: 1, RAMGB: 2, MonthlyCents: 599}})
	})
	_ = srv
	plans, err := client.ListPlans(context.Background(), "fra")
	if err != nil || len(plans) != 1 || plans[0].ID != "vps-std-1-2" {
		t.Fatalf("list plans: %v / %+v", err, plans)
	}
	if gotQuery != "regionCode=fra" {
		t.Fatalf("region filter not sent, got query %q", gotQuery)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
