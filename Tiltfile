load('ext://restart_process', 'docker_build_with_restart')

docker_build_with_restart('ctlptl-registry:5005/my-image',
    context='.',
    entrypoint="/usr/local/bin/cnetmon",
    only=["./src"],
    live_update=[
        sync('./src', '/usr/src/app'),
        run(
            'go build -o /usr/local/bin/cnetmon .'
        )
    ]
)

k8s_yaml(['k8s/daemonset.yaml', 'k8s/service.yaml'])
k8s_resource(workload='cnetmon', port_forwards=7777)
k8s_resource(workload='cnetmon', port_forwards=2808)
local_resource('ensure-cluster', cmd="ctlptl apply -f ctlptl/setup.yaml")
