package lightpath.webhook

test_create_patch_service_with_label {
	resp := main with input as {
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
	resp := main with input as {
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
	resp := main with input as {
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
	resp := main with input as {
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
