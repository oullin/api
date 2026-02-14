---
name: deployment-pipeline-design
description: Design multi-stage CI/CD pipelines with approval gates, security checks, and deployment orchestration. Use when architecting deployment workflows, setting up continuous delivery, or implementing GitOps practices.
---

# Deployment Pipeline Design

Architecture patterns for multi-stage GitLab CI/CD pipelines with approval gates and Docker-based deployment strategies.

## Purpose

Design robust, secure deployment pipelines that balance speed with safety through proper stage organisation and approval workflows.

## When to Use

- Design CI/CD architecture
- Implement deployment gates
- Configure multi-environment pipelines
- Establish deployment best practices
- Implement progressive delivery

## Pipeline Stages

### Standard Pipeline Flow

```
┌─────────┐   ┌──────┐   ┌─────────┐   ┌────────┐   ┌──────────┐
│  Build  │ → │ Test │ → │ Staging │ → │ Approve│ → │Production│
└─────────┘   └──────┘   └─────────┘   └────────┘   └──────────┘
```

### Detailed Stage Breakdown

1. **Source** - Code checkout
2. **Build** - Compile, package, containerize
3. **Test** – Unit, integration, security scans
4. **Staging Deploy** - Deploy to staging environment
5. **Integration Tests** - E2E, smoke tests
6. **Approval Gate** - Manual approval required
7. **Production Deploy** – Blue-green, rolling via Docker Compose
8. **Verification** - Health checks, monitoring
9. **Rollback** – Automated rollback on failure

## Approval Gate Patterns

### Pattern 1: Manual Approval

```yaml
# GitLab CI
deploy:production:
  stage: deploy
  script:
    - docker compose -f docker-compose.production.yml up -d
  environment:
    name: production
    url: https://app.example.com
  when: manual
  only:
    - main
```

### Pattern 2: Time-Based Approval

```yaml
# GitLab CI
deploy:production:
  stage: deploy
  script:
    - docker compose -f docker-compose.production.yml up -d
  environment:
    name: production
  when: delayed
  start_in: 30 minutes
  only:
    - main
```

### Pattern 3: Multi-Stage Gate with Review

```yaml
# GitLab CI
stages:
  - build
  - test
  - staging
  - review
  - production

deploy:staging:
  stage: staging
  script:
    - docker compose -f docker-compose.staging.yml up -d
  environment:
    name: staging
    url: https://staging.example.com
  only:
    - main

review:staging:
  stage: review
  script:
    - echo "Staging deployed. Awaiting manual review."
    - curl -sf https://staging.example.com/health
  when: manual
  only:
    - main

deploy:production:
  stage: production
  needs: ["review:staging"]
  script:
    - docker compose -f docker-compose.production.yml up -d
  environment:
    name: production
    url: https://app.example.com
  when: manual
  only:
    - main
```

## Deployment Strategies

### 1. Rolling Deployment with Docker Compose

```yaml
# docker-compose.production.yml
services:
  app:
    image: ${CI_REGISTRY_IMAGE}:${CI_COMMIT_SHA}
    deploy:
      replicas: 3
      update_config:
        parallelism: 1
        delay: 10s
        order: start-first
      restart_policy:
        condition: on-failure
    healthcheck:
      test: ["CMD", "wget", "--spider", "-q", "http://localhost:8080/health"]
      interval: 10s
      timeout: 5s
      retries: 3
    ports:
      - "8080:8080"
```

**Characteristics:**

- Gradual rollout
- Zero downtime
- Easy rollback
- Best for most applications

### 2. Blue-Green Deployment with Docker

```yaml
# GitLab CI blue-green deployment
deploy:blue-green:
  stage: deploy
  image: docker:24
  services:
    - docker:24-dind
  script:
    - |
      # Determine current active colour
      CURRENT=$(docker inspect --format='{{index .Config.Labels "deploy.colour"}}' app-active 2>/dev/null || echo "blue")
      if [ "$CURRENT" = "blue" ]; then NEW="green"; else NEW="blue"; fi

      # Deploy new version
      docker run -d \
        --name app-$NEW \
        --label deploy.colour=$NEW \
        -p 8081:8080 \
        --health-cmd="wget --spider -q http://localhost:8080/health" \
        --health-interval=5s \
        $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA

      # Wait for health check
      for i in $(seq 1 30); do
        STATUS=$(docker inspect --format='{{.State.Health.Status}}' app-$NEW 2>/dev/null)
        if [ "$STATUS" = "healthy" ]; then break; fi
        sleep 2
      done

      # Switch traffic (update reverse proxy / load balancer)
      docker stop app-$CURRENT || true
      docker rename app-$CURRENT app-old-$CURRENT || true
      docker rename app-$NEW app-active

      # Cleanup old container
      docker rm app-old-$CURRENT || true
  environment:
    name: production
  when: manual
  only:
    - main
```

**Characteristics:**

- Instant switchover
- Easy rollback
- Doubles infrastructure cost temporarily
- Good for high-risk deployments

### 3. Canary Deployment with Docker

```yaml
# GitLab CI canary deployment
deploy:canary:
  stage: deploy
  script:
    - |
      # Deploy canary (1 instance alongside existing)
      docker run -d \
        --name app-canary \
        --label app.role=canary \
        -p 8081:8080 \
        $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA

      # Verify canary health
      sleep 30
      HEALTH=$(curl -sf http://localhost:8081/health || echo "unhealthy")
      if [ "$HEALTH" = "unhealthy" ]; then
        docker stop app-canary && docker rm app-canary
        echo "Canary failed health check. Rolling back."
        exit 1
      fi

      echo "Canary healthy. Promote manually with deploy:production."
  environment:
    name: canary
  only:
    - main

deploy:promote:
  stage: deploy
  script:
    - docker stop app-canary || true
    - docker rm app-canary || true
    - docker compose -f docker-compose.production.yml up -d
  environment:
    name: production
  when: manual
  needs: ["deploy:canary"]
  only:
    - main
```

**Characteristics:**

- Gradual traffic shift
- Risk mitigation
- Real user testing
- Requires load balancer configuration

### 4. Feature Flags

```go
package featureflags

import (
	"os"
	"strconv"
	"sync"
)

// Flags holds feature flag state
type Flags struct {
	mu    sync.RWMutex
	flags map[string]bool
}

// New creates a Flags instance from environment variables
func New() *Flags {
	return &Flags{
		flags: make(map[string]bool),
	}
}

// IsEnabled checks if a feature flag is enabled
func (f *Flags) IsEnabled(name string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if val, ok := f.flags[name]; ok {
		return val
	}

	// Fall back to environment variable
	envVal, _ := strconv.ParseBool(os.Getenv("FF_" + name))
	return envVal
}

// Usage:
// flags := featureflags.New()
// if flags.IsEnabled("NEW_CHECKOUT_FLOW") {
//     processCheckoutV2()
// } else {
//     processCheckoutV1()
// }
```

**Characteristics:**

- Deploy without releasing
- A/B testing
- Instant rollback
- Granular control

## Pipeline Orchestration

### Multi-Stage Pipeline Example

```yaml
stages:
  - build
  - test
  - staging
  - integration
  - production
  - verify

variables:
  GO_VERSION: "1.21"
  CGO_ENABLED: "0"

build:
  stage: build
  image: golang:${GO_VERSION}-alpine
  script:
    - go mod download
    - go build -ldflags="-w -s" -o app ./cmd/server
  artifacts:
    paths:
      - app

build-docker:
  stage: build
  image: docker:24
  services:
    - docker:24-dind
  before_script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
  script:
    - docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA

test:
  stage: test
  image: golang:${GO_VERSION}-alpine
  script:
    - go test -v -race -coverprofile=coverage.txt ./...
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml

security-scan:
  stage: test
  image: aquasec/trivy:latest
  script:
    - trivy image --exit-code 1 --severity HIGH,CRITICAL $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  allow_failure: true

deploy-staging:
  stage: staging
  script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
    - docker pull $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    - docker compose -f docker-compose.staging.yml up -d
  environment:
    name: staging
    url: https://staging.example.com
  only:
    - main

integration-test:
  stage: integration
  image: golang:${GO_VERSION}-alpine
  script:
    - go test -tags=integration -v ./test/integration/...
  needs: ["deploy-staging"]

deploy-production:
  stage: production
  script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
    - docker pull $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    - docker compose -f docker-compose.production.yml up -d
  environment:
    name: production
    url: https://app.example.com
  when: manual
  only:
    - main

verify:
  stage: verify
  needs: ["deploy-production"]
  script:
    - |
      for i in $(seq 1 10); do
        if curl -sf https://app.example.com/health; then
          echo "Health check passed"
          exit 0
        fi
        sleep 10
      done
      echo "Health check failed"
      exit 1
    - |
      curl -X POST "$SLACK_WEBHOOK" \
        -d '{"text":"Production deployment successful!"}'
```

## Pipeline Best Practices

1. **Fail fast** – Run quick tests first
2. **Parallel execution** - Run independent jobs concurrently
3. **Caching** - Cache dependencies between runs
4. **Artefact management** – Store build artefacts
5. **Environment parity** - Keep environments consistent
6. **Secrets management** - Use secret stores (Vault, etc.)
7. **Deployment windows** – Schedule deployments appropriately
8. **Monitoring integration** – Track deployment metrics
9. **Rollback automation** - Auto-rollback on failures
10. **Documentation** - Document pipeline stages

## Rollback Strategies

### Automated Rollback

```yaml
deploy-and-verify:
  stage: production
  script:
    - |
      # Tag current image as rollback target
      docker tag $CI_REGISTRY_IMAGE:latest $CI_REGISTRY_IMAGE:rollback
      docker push $CI_REGISTRY_IMAGE:rollback

      # Deploy new version
      docker compose -f docker-compose.production.yml up -d

      # Health check loop
      for i in $(seq 1 10); do
        if curl -sf https://app.example.com/health; then
          echo "Deployment verified"
          exit 0
        fi
        sleep 10
      done

      # Rollback on failure
      echo "Health check failed, rolling back..."
      export CI_COMMIT_SHA=rollback
      docker compose -f docker-compose.production.yml up -d
      exit 1
  environment:
    name: production
  when: manual
  only:
    - main
```

### Manual Rollback

```bash
# List available image tags
docker images $CI_REGISTRY_IMAGE --format "{{.Tag}}\t{{.CreatedAt}}"

# Rollback to previous version
export CI_COMMIT_SHA=<previous-tag>
docker compose -f docker-compose.production.yml up -d

# Verify rollback
curl -sf https://app.example.com/health
```

## Monitoring and Metrics

### Key Pipeline Metrics

- **Deployment Frequency** – How often deployments occur
- **Lead Time** – Time from commit to production
- **Change Failure Rate** – Percentage of failed deployments
- **Meantime to Recovery (MTTR)** – Time to recover from failure
- **Pipeline Success Rate** - Percentage of successful runs
- **Average Pipeline Duration** - Time to complete pipeline

### Integration with Monitoring

```yaml
post-deploy-verify:
  stage: verify
  script:
    - |
      # Wait for metrics stabilization
      sleep 60

      # Check error rate
      ERROR_RATE=$(curl -s "$PROMETHEUS_URL/api/v1/query?query=rate(http_errors_total[5m])" | jq '.data.result[0].value[1]')

      if (( $(echo "$ERROR_RATE > 0.01" | bc -l) )); then
        echo "Error rate too high: $ERROR_RATE"
        exit 1
      fi
```

## Related Skills

- `gitlab-ci-patterns` - For GitLab CI implementation
- `secrets-management` - For secrets handling
