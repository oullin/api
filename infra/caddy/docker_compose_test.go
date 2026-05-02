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

func TestComposeKeepsProxyAliasOnProdCaddy(t *testing.T) {
	content, err := os.ReadFile("../../docker-compose.yml")
	if err != nil {
		t.Fatalf("read docker-compose.yml: %v", err)
	}

	var compose yaml.Node
	if err := yaml.Unmarshal(content, &compose); err != nil {
		t.Fatalf("parse docker-compose.yml: %v", err)
	}

	root := compose.Content[0]
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
