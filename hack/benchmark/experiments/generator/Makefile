gen:
	go run gen.go
	dot -Tpdf topology.dot > topology.pdf
	tk show tanka -V LIGHTPATH_DISABLED=false --dangerous-allow-redirect > k8s-lightpath-deploy.yaml
	tk show tanka -V LIGHTPATH_DISABLED=true --dangerous-allow-redirect > k8s-kube-proxy-deploy.yaml

export-%: gen
	mkdir -p ../scenarios/$*
	cp topology.* ../scenarios/$*
	cp k8s-*-deploy.yaml ../scenarios/$*
