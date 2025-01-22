package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
)

const (
	defaultImage       = "container-registry.ubs.net/ubs/ibent/k8-hello-whisky:1.0.78-SNAPSHOT"
	defaultGateway     = "aks-istio-ingress/gateway-wildcard"
	clusterSuffixEnv   = "CLUSTER_SUFFIX"
	resourceGroupEnv   = "RESOURCE_GROUP_NAME"
	privateDNSZoneEnv  = "PRIVATE_DNS_ZONE_NAME"
	subscriptionIDEnv  = "AZURE_SUBSCRIPTION_ID"
	checkTimeout       = 2 * time.Minute
	dnsQueryTimeout    = 1 * time.Minute
)

type Config struct {
	Namespace     string
	AppName       string
	ContainerPort int32
	ServicePort   int32
	Replicas      int32
	Image         string
	Host          string
	Gateway       string
	WaitTime      time.Duration
}

func getUniqueHost(baseHost string) string {
	clusterSuffix := os.Getenv(clusterSuffixEnv)
	if clusterSuffix == "" {
		clusterSuffix = "default"
	}
	return fmt.Sprintf("%s-%s", baseHost, clusterSuffix)
}

func checkHostResponse(t *testing.T, host string) {
	t.Logf("Checking response from host: https://%s", host)
	start := time.Now()
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	for {
		select {
		case <-time.After(checkTimeout):
			t.Fatalf("Timed out waiting for host https://%s to respond with 200 OK", host)
		case <-ticker.C:
			resp, err := httpClient.Get(fmt.Sprintf("https://%s", host))
			if err != nil {
				t.Logf("Host https://%s not responding yet: %v", host, err)
				continue
			}
			if resp.StatusCode == http.StatusOK {
				t.Logf("Host https://%s responded with 200 OK after %s", host, time.Since(start))
				return
			}
			t.Logf("Host https://%s responded with status %d, retrying...", host, resp.StatusCode)
		}
	}
}

func checkPrivateDNSRecord(t *testing.T, dnsName string) {
	privateDNSZone := os.Getenv(privateDNSZoneEnv)
	resourceGroup := os.Getenv(resourceGroupEnv)
	subscriptionID := os.Getenv(subscriptionIDEnv)

	if privateDNSZone == "" || resourceGroup == "" || subscriptionID == "" {
		t.Fatalf("Environment variables %s, %s, and %s must be set", privateDNSZoneEnv, resourceGroupEnv, subscriptionIDEnv)
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		t.Fatalf("Failed to obtain Azure credential: %v", err)
	}

	client, err := armprivatedns.NewRecordSetsClient(subscriptionID, cred, nil)
	if err != nil {
		t.Fatalf("Failed to create Private DNS client: %v", err)
	}

	t.Logf("Checking Private DNS record: %s in zone: %s", dnsName, privateDNSZone)
	ctx, cancel := context.WithTimeout(context.Background(), dnsQueryTimeout)
	defer cancel()

	pager := client.NewListByTypePager(resourceGroup, privateDNSZone, armdns.RecordTypeA, nil)

	found := false
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			t.Fatalf("Failed to query DNS records: %v", err)
		}
		for _, record := range page.Value {
			if record.Name != nil && *record.Name == dnsName {
				t.Logf("Private DNS record %s exists", dnsName)
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if !found {
		t.Fatalf("Private DNS record %s not found", dnsName)
	}
}

func TestAKSHelloWorldDeploymentIstio(t *testing.T) {
	cfg := Config{
		Namespace:     "test-app2",
		AppName:       "whisky-app",
		ContainerPort: 8000,
		ServicePort:   80,
		Replicas:      3,
		Image:         defaultImage,
		Host:          getUniqueHost("whisky-app-testvs"),
		Gateway:       defaultGateway,
		WaitTime:      60 * time.Second,
	}

	t.Logf("Using host: https://%s", cfg.Host)

	// Check if the Private DNS record exists
	checkPrivateDNSRecord(t, cfg.Host)

	// Check if the host is responding
	checkHostResponse(t, cfg.Host)

	t.Log("Resources created successfully and host is responding with 200 OK")
}