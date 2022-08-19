local k = import 'ksonnet-util/kausal.libsonnet';
local topologyYaml = importstr '../topology.yaml';

{
  _config+:: {
    topology: std.parseYaml(topologyYaml),
    image: 'ghcr.io/xvzf/lightpath/isotope:latest@sha256:06a7ff55479c9b16de75f8d9285f72d89afdf05e1e807bda8a67ca751fe7b062',  // Build of isotope as the one used in istio test infra is not publicly served

    replicas_per_service: 3,
  },

  ns:
    k.core.v1.namespace.new('scenario'),

  cm:
    k.core.v1.configMap.new('config')
    + k.core.v1.configMap.withData({
      'service-graph.yaml': importstr '../topology.yaml',
    }),

  container::
    k.core.v1.container.new(name='performance-test', image=$._config.image)
    + k.core.v1.container.withImagePullPolicy('Always')
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
        + k.apps.v1.deployment.spec.template.metadata.withAnnotations({
          'prometheus.io/scrape': 'true',
          'prometheus.io/port': '8080',
          'prometheus.io/path': '/metrics',
        })
        + k.util.configMapVolumeMount($.cm, '/etc/config/'),

      service_lightpath:
        k.util.serviceFor(self.deployment, nameFormat='%(port)s')
        + self.entrypoint_mixin
        + self.lightpath_disable_mixin,

      entrypoint_mixin::
        if ($._config.topology.defaults + svc).isEntrypoint
        then
          k.core.v1.service.spec.withType('NodePort')
        else {},

      lightpath_disable_mixin::
        if std.extVar('LIGHTPATH_DISABLED') == 'true'
        then
          k.core.v1.service.metadata.withLabelsMixin({
            'lightpath.cloud/proxy': 'disabled',
          })
        else
          k.core.v1.service.metadata.withAnnotationsMixin({
            // disable access log; it decrases performance and kube-proxy does not log either
            'config.lightpath.cloud/http-access-log': 'disabled',
          }),
    }

    for svc in $._config.topology.services
  },

}
