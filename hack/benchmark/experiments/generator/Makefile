gen:
	go run gen.go
	dot -Tpdf topology.dot > topology.pdf
	tk show tanka --dangerous-allow-redirect > k8s-deploy.yaml

export-%: gen
	mkdir -p ../scenarios/$*
	cp topology.* ../scenarios/$*
	cp k8s-deploy.yaml ../scenarios/$*
