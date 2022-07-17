local k = import 'ksonnet-util/kausal.libsonnet';
local topologyYaml = importstr '../topology.yaml';

{
  _config+:: {
    topology: std.parseYaml(topologyYaml),
    image:: 'nginx',

    replicas_per_service: 2,
  },

  ns:
    k.core.v1.namespace.new('topology'),

  cm:
    k.core.v1.configMap.new('config')
    + k.core.v1.configMap.withData({
      'topology.yaml': importstr '../topology.yaml',
    }),

  container::
    k.core.v1.container.new(name='performance-test', image=$._config.image)
    + k.core.v1.container.withPorts([
      k.core.v1.containerPort.new('http', 8080),
    ])
    + k.util.resourcesRequests('100m', '100Mi')
    + k.util.resourcesLimits('200m', '150Mi'),

  services: {
    [svc.name]: {
      deployment:
        k.apps.v1.deployment.new(
          name=svc.name,
          replicas=$._config.replicas_per_service,
          containers=[
            $.container
            + k.core.v1.container.withEnvMap({
              SERVICE_NAME: svc.name,
            }),
          ],
        )
        + k.util.configMapVolumeMount($.cm, '/etc/config/service-graph.yaml'),
      service: k.util.serviceFor(self.deployment),
    }

    for svc in $._config.topology.services
  },


}
