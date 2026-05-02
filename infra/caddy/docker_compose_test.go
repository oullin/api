package caddy_test

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

func mappingValue(node *yaml.Node, key string) *yaml.Node {
	if node == nil || node.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i < len(node.Content)-1; i += 2 {
		if node.Content[i].Value == key {
			return node.Content[i+1]
		}
	}

	return nil
}

func readComposeRoot(t *testing.T) *yaml.Node {
	t.Helper()

	content, err := os.ReadFile("../../docker-compose.yml")
	if err != nil {
		t.Fatalf("read docker-compose.yml: %v", err)
	}

	var compose yaml.Node
	if err := yaml.Unmarshal(content, &compose); err != nil {
		t.Fatalf("parse docker-compose.yml: %v", err)
	}

	return compose.Content[0]
}

func sequenceContains(node *yaml.Node, value string) bool {
	if node == nil || node.Kind != yaml.SequenceNode {
		return false
	}

	for _, item := range node.Content {
		if item.Value == value {
			return true
		}
	}

	return false
}

func TestComposeKeepsProxyAliasOnProdCaddy(t *testing.T) {
	root := readComposeRoot(t)
	services := mappingValue(root, "services")
	caddyProd := mappingValue(services, "caddy_prod")
	if caddyProd == nil {
		t.Fatal("expected caddy_prod service to exist")
	}

	networks := mappingValue(caddyProd, "networks")
	caddyNet := mappingValue(networks, "caddy_net")
	if caddyNet == nil {
		t.Fatal("expected caddy_prod to be attached to caddy_net")
	}

	aliases := mappingValue(caddyNet, "aliases")
	if aliases == nil || aliases.Kind != yaml.SequenceNode {
		t.Fatal("expected caddy_prod caddy_net aliases to be a sequence")
	}

	for _, alias := range aliases.Content {
		if alias.Value == "proxy" {
			return
		}
	}

	t.Fatal("expected caddy_prod caddy_net aliases to include proxy")
}

func TestComposeProdCaddyWaitsForHealthyAPI(t *testing.T) {
	root := readComposeRoot(t)
	services := mappingValue(root, "services")
	caddyProd := mappingValue(services, "caddy_prod")
	dependsOn := mappingValue(caddyProd, "depends_on")
	api := mappingValue(dependsOn, "api")
	condition := mappingValue(api, "condition")

	if condition == nil || condition.Value != "service_healthy" {
		t.Fatalf("expected caddy_prod to wait for healthy api, got %#v", condition)
	}
}

func TestComposeAPIHasHealthcheckAndPersistentLogs(t *testing.T) {
	root := readComposeRoot(t)
	services := mappingValue(root, "services")
	api := mappingValue(services, "api")
	if api == nil {
		t.Fatal("expected api service to exist")
	}

	healthcheck := mappingValue(api, "healthcheck")
	test := mappingValue(healthcheck, "test")
	if !sequenceContains(test, "wget --no-verbose --tries=1 --spider http://localhost:$${ENV_HTTP_PORT:-8080}/health") {
		t.Fatalf("expected api healthcheck to call /health")
	}

	volumes := mappingValue(api, "volumes")
	if !sequenceContains(volumes, "${API_LOGS_PATH:-./storage/logs/api}:/app/storage/logs") {
		t.Fatalf("expected api logs to be persisted on the host")
	}

	environment := mappingValue(api, "environment")
	logsDir := mappingValue(environment, "ENV_APP_LOGS_DIR")
	if logsDir == nil || logsDir.Value != "/app/storage/logs/logs_%s.log" {
		t.Fatalf("expected api logs dir to target persisted mount, got %#v", logsDir)
	}
}
