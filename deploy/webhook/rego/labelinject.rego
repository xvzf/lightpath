package lightpath.webhook

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
		"patch": base64url.encode(json.marshal(patch)),
	},
}

# Exclude well-known namespaces required for bootstrapping
well_known_exclusions = [
	"lightpath-system",
	"kube-system",
	"cert-manager",
]

patch[p] {
	# Only trigger on create or update
	input.request.operation == "CREATE"

	# Only applies when the label does not exist
	not input.request.object.metadata.labels["service.kubernetes.io/service-proxy-name"]

	# Exclude well-known namespaces we are relying on
	not input.object.metadata.namespace[well_known_exclusions]

	p := {
		"op": "add",
		# Well-known label as defined here: https://kubernetes.io/docs/reference/labels-annotations-taints/#servicekubernetesioservice-proxy-name
		"path": "/metadata/annotations/foo/service.kubernetes.io\\/service-proxy-name",
		"value": "lightpath",
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
