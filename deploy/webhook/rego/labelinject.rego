package lightpath.webhook

import future.keywords.in

req_uid := input.request.uid

mutate = {
	"apiVersion": "admission.k8s.io/v1",
	"kind": "AdmissionReview",
	"response": {
		"allowed": count(deny) == 0,
		"uid": req_uid,
		"patchType": "JSONPatch",
		"status": {"message": concat(", ", deny)},
		# Patch label
		"patch": base64.encode(json.marshal(patch)),
	},
}

# Exclude well-known namespaces required for bootstrapping
well_known_exclusions = [
	"lightpath-system",
	"kube-system",
	"cert-manager",
]

patch_conditions {
	# Only trigger on create or update
	input.request.operation == "CREATE"

  # match only TCP services
  some port in input.request.object.spec.ports
	not port.protocol != "TCP"

	# Only applies when the label does not exist
	not input.request.object.metadata.labels["service.kubernetes.io/service-proxy-name"]

	# FIXME change to opt-in
	# Only applies when it's not disabled
	not input.request.object.metadata.labels["lightpath.cloud/proxy"] == "disabled"

	# Exclude well-known namespaces we are relying on
	not input.object.metadata.namespace[well_known_exclusions]
}

patch[p] {
	# metadata.labels exists -> we have to patch it
	input.request.object.metadata.labels
	patch_conditions

	p := {
		"op": "add",
		# Well-known label as defined here: https://kubernetes.io/docs/reference/labels-annotations-taints/#servicekubernetesioservice-proxy-name
		"path": "/metadata/labels/service.kubernetes.io~1service-proxy-name",
		"value": "lightpath.cloud",
	}
}

patch[p] {
	# metadata.labels doesn't exist yet -> we have to create it
	not input.request.object.metadata.labels
	patch_conditions
	p := {
		"op": "add",
		# Well-known label as defined here: https://kubernetes.io/docs/reference/labels-annotations-taints/#servicekubernetesioservice-proxy-name
		"path": "/metadata/labels",
		"value": {"service.kubernetes.io/service-proxy-name": "lightpath.cloud"},
	}
}

deny["kind is not supported (requiring Service)"] {
	# Only trigger on Services
	not input.request.kind.kind == "Service"
}

deny["version is not supported (requiring v1)"] {
	# Only trigger on Services
	not input.request.kind.version == "v1"
}
