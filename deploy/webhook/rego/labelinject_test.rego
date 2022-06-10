package lightpath.webhook

test_create_patch_service_with_label {
	resp := mutate with input as {
		"kind": "AdmissionReview",
		"request": {
			"operation": "CREATE",
			"kind": {
				"kind": "Service",
				"version": "v1",
			},
			"object": {"metadata": {"labels": {"service.kubernetes.io/service-proxy-name": "kube-proxy"}}},
		},
	}

	resp
	payload := resp.response
	payload.patchType == "JSONPatch"
	patch := json.unmarshal(base64url.decode(payload.patch))
	patch == []
}

test_create_patch_service_without_label {
	resp := mutate with input as {
		"kind": "AdmissionReview",
		"request": {
			"operation": "CREATE",
			"kind": {
				"kind": "Service",
				"version": "v1",
			},
			"object": {"metadata": {}},
		},
	}

	resp
	payload = resp.response
	payload.allowed
	payload.patchType == "JSONPatch"
	patches = json.unmarshal(base64url.decode(payload.patch))
	patches[0].op == "add"
	patches[0].path == "/metadata/annotations/foo/service.kubernetes.io\\/service-proxy-name"
	patches[0].value == "lightpath"
}

test_create_patch_invalid_kind {
	resp := mutate with input as {
		"kind": "AdmissionReview",
		"request": {
			"operation": "CREATE",
			"kind": {
				"kind": "Pod",
				"version": "v1",
			},
			"object": {"metadata": {"labels": {"service.kubernetes.io/service-proxy-name": "kube-proxy"}}},
		},
	}

	resp
	payload := resp.response
	not payload.allowed
	contains(payload.status.message, "requiring Service")
}

test_create_patch_invalid_version {
	resp := mutate with input as {
		"kind": "AdmissionReview",
		"request": {
			"operation": "CREATE",
			"kind": {
				"kind": "Service",
				"version": "v2",
			},
			"object": {"metadata": {"labels": {"service.kubernetes.io/service-proxy-name": "kube-proxy"}}},
		},
	}

	resp
	payload := resp.response
	not payload.allowed
	contains(payload.status.message, "requiring v1")
}

test_create_patch_well_known_excluded_namespace {
	resp := mutate with input as {
		"kind": "AdmissionReview",
		"request": {
			"operation": "CREATE",
			"kind": {
				"kind": "Service",
				"version": "v1",
			},
			"object": {"metadata": {
				"namespace": "kube-system",
				"labels": {"service.kubernetes.io/service-proxy-name": "kube-proxy"},
			}},
		},
	}

	resp
	payload := resp.response
	payload.allowed
	count(patch) == 0
}
