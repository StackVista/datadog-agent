import os
import util

from testinfra.utils.ansible_runner import AnsibleRunner

testinfra_hosts = AnsibleRunner(os.environ['MOLECULE_INVENTORY_FILE']).get_hosts('kubernetes-cluster-agent')


def test_receiver_healthy(host):
    def assert_healthy():
        c = "curl -s -o /dev/null -w \"%{http_code}\" http://localhost:7077/health"
        assert host.check_output(c) == "200"

    util.wait_until(assert_healthy, 100, 5)


def test_node_agent_healthy(host):
    def assert_healthy():
        c = "kubectl wait --for=condition=available --timeout=30s daemonset/stackstate-agent"
        assert host.run(c).rc == 0

    util.wait_until(assert_healthy, 100, 5)


def test_cluster_agent_healthy(host):
    def assert_healthy():
        c = "kubectl wait --for=condition=available --timeout=30s deployment/stackstate-cluster-agent"
        assert host.run(c).rc == 0

    util.wait_until(assert_healthy, 100, 5)
