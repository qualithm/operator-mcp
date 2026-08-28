package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// httpProvisioner claims devices against the provisioning API.
type httpProvisioner struct {
	baseURL string
	http    *http.Client
}

// claim implements provisioner: POST /provision/claim with the claim code.
func (p *httpProvisioner) claim(ctx context.Context, code, name string) (claimResult, error) {
	body, err := json.Marshal(map[string]string{"code": code, "name": name})
	if err != nil {
		return claimResult{}, fmt.Errorf("encode claim body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		strings.TrimRight(p.baseURL, "/")+"/provision/claim", bytes.NewReader(body))
	if err != nil {
		return claimResult{}, fmt.Errorf("build claim request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	res, err := p.http.Do(req)
	if err != nil {
		return claimResult{}, fmt.Errorf("claim request: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return claimResult{}, fmt.Errorf("read claim response: %w", err)
	}
	var env struct {
		Data    claimResult `json:"data"`
		Message string      `json:"message"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		return claimResult{}, fmt.Errorf("decode claim response: %w", err)
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return claimResult{}, fmt.Errorf("claim failed: HTTP %d: %s", res.StatusCode, env.Message)
	}
	return env.Data, nil
}

// mqttPublisher connects a claimed device to the gateway over MQTT/TLS and
// publishes one telemetry reading. The wire contract matches the device SDK:
// the MQTT client id is the device UUID, the secret is the password, and
// telemetry goes to the device-relative `telemetry` topic as
// {ts, metrics: {name: value}}.
type mqttPublisher struct {
	host string
	port int
}

// publishTelemetry implements publisher.
func (p *mqttPublisher) publishTelemetry(ctx context.Context, cred claimResult, metric string, value float64, ts int64) error {
	opts := mqtt.NewClientOptions().
		AddBroker(fmt.Sprintf("ssl://%s:%d", p.host, p.port)).
		SetClientID(cred.DeviceID).
		SetUsername(cred.DeviceID).
		SetPassword(cred.Secret).
		SetTLSConfig(&tls.Config{MinVersion: tls.VersionTLS13}).
		SetConnectTimeout(15 * time.Second)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	if !token.WaitTimeout(20 * time.Second) {
		return fmt.Errorf("mqtt connect timed out")
	}
	defer client.Disconnect(1000)
	if err := token.Error(); err != nil {
		return fmt.Errorf("mqtt connect: %w", err)
	}

	payload, err := json.Marshal(map[string]any{
		"ts":      ts,
		"metrics": map[string]float64{metric: value},
	})
	if err != nil {
		return fmt.Errorf("encode telemetry: %w", err)
	}
	pub := client.Publish("telemetry", 1, false, payload)
	if !pub.WaitTimeout(15 * time.Second) {
		return fmt.Errorf("mqtt publish timed out")
	}
	if err := pub.Error(); err != nil {
		return fmt.Errorf("mqtt publish: %w", err)
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
