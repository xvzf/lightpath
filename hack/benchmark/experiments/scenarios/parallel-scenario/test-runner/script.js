import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  discardResponseBodies: true,
  vus: 2,
  duration: '5m'
}

export default function () {

  // VU count starts at 1 -> one VU calls the kube proxy endpoint the other one lightpath-handled endpoint
  if (__VU == 1) {
    let resKubeProxy = http.post(`http://${__ENV.KUBE_PROXY_CLUSTERIP}:8080/`);
    check(resKubeProxy, { 'kube_proxy_success': (r) => r.status === 200 });
  }
  if (__VU == 2) {
    let resLightpath = http.post(`http://${__ENV.LIGHTPATH_CLUSTERIP}:8080/`);
    check(resLightpath, { 'lightpath_success': (r) => r.status === 200 });
  }

}

