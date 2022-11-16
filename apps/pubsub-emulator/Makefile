app := pubsub-emulator
image := gcr.io/inshur-prod0-repo0/$(app)
tag := latest

ci: docker ## Build in CI

ci-arm: docker-arm ## Build in CI for arm platform

docker: ## Build Docker image
	DOCKER_BUILDKIT=1 docker build --progress=plain \
		--tag $(image):latest \
		--tag $(image):$(tag) \
		.
#         PUBSUB_PROJECT1=inshur-example,integration-events;integration-events-sub;ie-ges-sub;global-in-subscription;ie-test-sub,event-store-seed;ies-ges-sub;seed-subscription,event-store-commands;esc-sub;eventstore-commands-subscription,policy-api-publishing-seed;pap-sub,passage-of-time;time-sub;platform-sidecar-sidecar-passage-of-time,action-triggers;platform-sidecar-sidecar-triggers

docker-arm: ## Build Docker image
	DOCKER_BUILDKIT=1 docker build --progress=plain \
		--tag $(image):latest \
		--tag $(image):$(tag) \
		--file Dockerfile_arm \
		.

start: docker ## Start the service
	docker run --env PUBSUB_PROJECT1="proj,topic1;topic1-sub1;topic1-sub2,topic2;topic2-sub1;topic2-sub2" $(image):$(tag)

push: ## Push images to registry
	docker push $(image):builder || true
	docker push $(image):latest
	docker push $(image):$(tag)
