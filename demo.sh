#!/usr/bin/env bash

# make sure to start the server first!

echo "Creating some pipelines..."
	ret1=$(./stagerunner client --token "secret" create \
	'{"name": "pipeline1", "repository": "repo1", "stages": {"run_stage": {"command": "go test ./..."}, "build_stage": {"dockerfile_path": "Dockerfile"}, "deploy_stage": {"cluster_name": "staging_eks_cluster", "manifest_path": "k8s/staging"}}}')
	
	pid1=$(echo $ret1 | cut -d':' -f2)

	ret2=$(./stagerunner client --token "secret" create \
	'{"name": "pipeline2", "repository": "repo2", "stages": {"run_stage": {"command": "go test ./..."}, "build_stage": {"dockerfile_path": "Dockerfile"}, "deploy_stage": {"cluster_name": "production_eks_cluster", "manifest_path": "k8s/production"}}}')
	
	pid2=$(echo $ret2 | cut -d':' -f2)

	echo "Triggering some pipeline runs for pipeline1..."
	ret3=$(./stagerunner client --token "secret" trigger $pid1 dev-branch)
	rid3=$(echo $ret3 | cut -d':' -f2)

	ret4=$(./stagerunner client --token "secret" trigger $pid1 feature-branch)
	rid4=$(echo $ret4 | cut -d':' -f2)

	ret5=$(./stagerunner client --token "secret" trigger $pid1 main)
	rid5=$(echo $ret5 | cut -d':' -f2)

	echo "Triggering some pipeline runs for pipeline2..."
	ret6=$(./stagerunner client --token "secret" trigger $pid2 dev-branch)
	rid6=$(echo $ret6 | cut -d':' -f2)

	ret7=$(./stagerunner client --token "secret" trigger $pid2 feature-branch)
	rid7=$(echo $ret7 | cut -d':' -f2)

	ret8=$(./stagerunner client --token "secret" trigger $pid2 main)
	rid8=$(echo $ret8 | cut -d':' -f2)


	echo
	for i in $(seq 1 10); do
		echo "Getting the run statuses for pipeline1..."
		./stagerunner client --token "secret" get-run $rid3
		./stagerunner client --token "secret" get-run $rid4
		./stagerunner client --token "secret" get-run $rid5

		echo
		echo
		echo "Getting the run statuses for pipeline2..."
		./stagerunner client --token "secret" get-run $rid6
		./stagerunner client --token "secret" get-run $rid7
		./stagerunner client --token "secret" get-run $rid8


		echo
		echo "sleeping 5 seconds..."
		echo
		sleep 5
	done

	echo "Done!"