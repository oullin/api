---
name: gitlab-ci-patterns
description: Build GitLab CI/CD pipelines with multi-stage workflows, caching, and distributed runners for scalable automation. Use when implementing GitLab CI/CD, optimizing pipeline performance, or setting up automated testing and deployment.
---

# GitLab CI Patterns

Comprehensive GitLab CI/CD pipeline patterns for automated testing, building, and deployment.

## Purpose

Create efficient GitLab CI pipelines with proper stage organisation, caching, and deployment strategies.

## When to Use

- Automate GitLab-based CI/CD
- Implement multi-stage pipelines
- Configure GitLab Runners
- Deploy Docker containers from GitLab
- Implement GitOps workflows

## Basic Pipeline Structure

```yaml
stages:
  - build
  - test
  - deploy

variables:
  DOCKER_DRIVER: overlay2
  DOCKER_TLS_CERTDIR: "/certs"
  GO_VERSION: "1.21"
  CGO_ENABLED: "0"
  GOOS: "linux"
  GOARCH: "amd64"

build:
  stage: build
  image: golang:${GO_VERSION}-alpine
  before_script:
    - apk add --no-cache git make
  script:
    - go mod download
    - go build -ldflags="-w -s" -o app ./cmd/main.go
  artifacts:
    paths:
      - app
      - go.mod
      - go.sum
    expire_in: 1 hour
  cache:
    key: ${CI_COMMIT_REF_SLUG}-go
    paths:
      - .go/pkg/mod/

test:
  stage: test
  image: golang:${GO_VERSION}-alpine
  before_script:
    - apk add --no-cache git make gcc musl-dev
  script:
    - go mod download
    - go fmt ./...
    - go vet ./...
    - go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...
    - go tool cover -html=coverage.txt -o coverage.html
  coverage: '/total:\s+\(statements\)\s+(\d+\.\d+)%/'
  artifacts:
    reports:
      coverage_report:
        coverage_format: cobertura
        path: coverage.xml
    paths:
      - coverage.txt
      - coverage.html

deploy:
  stage: deploy
  image: docker:24
  services:
    - docker:24-dind
  before_script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
  script:
    - docker pull $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    - docker stop my-app || true
    - docker rm my-app || true
    - |
      docker run -d \
        --name my-app \
        --restart unless-stopped \
        -p 8080:8080 \
        $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  only:
    - main
  environment:
    name: production
    url: https://app.example.com
```

## Docker Build and Push

```yaml
build-docker:
  stage: build
  image: docker:24
  services:
    - docker:24-dind
  before_script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY
  script:
    - |
      docker build \
        --build-arg GO_VERSION=1.21 \
        --target production \
        -t $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA \
        -t $CI_REGISTRY_IMAGE:latest \
        -f Dockerfile .
    - docker push $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    - docker push $CI_REGISTRY_IMAGE:latest
  only:
    - main
    - tags
```

## Multi-Environment Deployment

```yaml
.deploy_template: &deploy_template
  image: docker:24
  services:
    - docker:24-dind
  before_script:
    - docker login -u $CI_REGISTRY_USER -p $CI_REGISTRY_PASSWORD $CI_REGISTRY

deploy:staging:
  <<: *deploy_template
  stage: deploy
  script:
    - docker pull $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    - docker compose -f docker-compose.staging.yml up -d
  environment:
    name: staging
    url: https://staging.example.com
  only:
    - develop

deploy:production:
  <<: *deploy_template
  stage: deploy
  script:
    - docker pull $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
    - docker compose -f docker-compose.production.yml up -d
  environment:
    name: production
    url: https://app.example.com
  when: manual
  only:
    - main
```

## Terraform Pipeline

```yaml
stages:
  - validate
  - plan
  - apply

variables:
  TF_ROOT: ${CI_PROJECT_DIR}/terraform
  TF_VERSION: "1.6.0"

before_script:
  - cd ${TF_ROOT}
  - terraform --version

validate:
  stage: validate
  image: hashicorp/terraform:${TF_VERSION}
  script:
    - terraform init -backend=false
    - terraform validate
    - terraform fmt -check

plan:
  stage: plan
  image: hashicorp/terraform:${TF_VERSION}
  script:
    - terraform init
    - terraform plan -out=tfplan
  artifacts:
    paths:
      - ${TF_ROOT}/tfplan
    expire_in: 1 day

apply:
  stage: apply
  image: hashicorp/terraform:${TF_VERSION}
  script:
    - terraform init
    - terraform apply -auto-approve tfplan
  dependencies:
    - plan
  when: manual
  only:
    - main
```

## Security Scanning

```yaml
include:
  - template: Security/SAST.gitlab-ci.yml
  - template: Security/Dependency-Scanning.gitlab-ci.yml
  - template: Security/Container-Scanning.gitlab-ci.yml

trivy-scan:
  stage: test
  image: aquasec/trivy:latest
  script:
    - trivy image --exit-code 1 --severity HIGH,CRITICAL $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA
  allow_failure: true
```

## Caching Strategies

```yaml
# Cache Go modules
build:
  cache:
    key: ${CI_COMMIT_REF_SLUG}-go
    paths:
      - .go/pkg/mod/
      - /go/pkg/mod/
    policy: pull-push

# Global cache
cache:
  key: ${CI_COMMIT_REF_SLUG}-go
  paths:
    - .cache/
    - vendor/
    - /go/pkg/mod/

# Separate cache per job
build-job:
  cache:
    key: build-cache-go
    paths:
      - bin/
      - .go/pkg/mod/

test-job:
  cache:
    key: test-cache-go
    paths:
      - .go/pkg/mod/
      - test-results/
```

## Dynamic Child Pipelines

```yaml
generate-pipeline:
  stage: build
  script:
    - python generate_pipeline.py > child-pipeline.yml
  artifacts:
    paths:
      - child-pipeline.yml

trigger-child:
  stage: deploy
  trigger:
    include:
      - artifact: child-pipeline.yml
        job: generate-pipeline
    strategy: depend
```

## Reference Files

- `assets/gitlab-ci.yml.template` - Complete pipeline template
- `references/pipeline-stages.md` - Stage organization patterns

## Best Practices

1. **Use specific image tags** (golang:1.25-alpine, not golang:latest)
2. **Cache Go modules** appropriately using GOMODCACHE
3. **Use artefacts** for compiled binaries
4. **Implement manual gates** for production
5. **Use environments** for deployment tracking
6. **Enable merge request pipelines**
7. **Use pipeline schedules** for recurring jobs
8. **Implement security scanning** with gosec and trivy
9. **Use CI/CD variables** for secrets
10. **Monitor pipeline performance**
11. **Use multi-stage Docker builds** for smaller images
12. **Set CGO_ENABLED=0** for static binaries

## Related Skills

- `secrets-management` - For secrets handling
