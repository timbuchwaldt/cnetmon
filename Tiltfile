load('ext://restart_process', 'docker_build_with_restart')

docker_build_with_restart('ctlptl-registry:5005/my-image',
    context='.',
    dockerfile='Dockerfile.tilt',
    entrypoint="/usr/local/bin/cnetmon",
    only=["./out"],

    live_update=[
        sync('./out/cnetmon', '/usr/local/bin/cnetmon'),
    ]
)

k8s_yaml(['k8s/ns.yaml', 'k8s/daemonset.yaml', 'k8s/service.yaml'])
k8s_resource(workload='cnetmon', port_forwards=7777)
k8s_resource(workload='cnetmon', port_forwards=2808)
local_resource('ensure-cluster', cmd="ctlptl apply -f ctlptl/setup.yaml")

local_resource(
  'build go binary',
  'go build -ldflags="-w -s" -o ../out/cnetmon .',
  deps=['./src'],
  dir="src",
  env={
    'GOOS': 'linux',
    'GOARCH': 'amd64',
    'CGO_ENABLED': '0'
  }
)
