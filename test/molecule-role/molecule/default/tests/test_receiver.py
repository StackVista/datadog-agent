import os
import json
import util
from collections import defaultdict
from molecule.util import safe_load_file
from testinfra.utils.ansible_runner import AnsibleRunner

testinfra_hosts = AnsibleRunner(os.environ['MOLECULE_INVENTORY_FILE']).get_hosts('receiver_vm')


def test_etc_docker_directory(host):
    f = host.file('/etc/docker/')
    assert f.is_directory


def test_docker_compose_file(host):
    f = host.file('/home/ubuntu/docker-compose.yml')
    assert f.is_file


def test_receiver_ok(host):
    c = "curl -s -o /dev/null -w \"%{http_code}\" http://localhost:7077/health"
    assert host.check_output(c) == "200"


def test_generic_events(host):
    url = "http://localhost:7070/api/topic/sts_generic_events?offset=0&limit=40"

    def wait_for_metrics():
        data = host.check_output("curl \"%s\"" % url)
        json_data = json.loads(data)
        print(json.dumps(json_data))

        events = defaultdict(set)
        for message in json_data["messages"]:
            events[message["message"]["GenericEvent"]["host"]].add(message["message"]["GenericEvent"]["name"])

        print events
        assert events["agent-ubuntu"] == {"System.Agent Startup", "processStateEvent"}
        assert events["agent-fedora"] == {"System.Agent Startup", "processStateEvent"}
        assert events["agent-win"] == {"System.Agent Startup"}

    util.wait_until(wait_for_metrics, 30, 3)


def _get_instance_config(instance_name):
    instance_config_dict = safe_load_file(os.environ['MOLECULE_INSTANCE_CONFIG'])
    return next(item for item in instance_config_dict if item['instance'] == instance_name)


def _find_outgoing_connection(json_data, port, origin, dest):
    """Find Connection as seen from the sending endpoint"""
    return next(connection for message in json_data["messages"]
                for connection in message["message"]["Connections"]["connections"]
                if connection["remoteEndpoint"]["endpoint"]["port"] == port and
                connection["remoteEndpoint"]["endpoint"]["ip"]["address"] == dest and
                connection["localEndpoint"]["endpoint"]["ip"]["address"] == origin
                )


def _find_incoming_connection(json_data, port, origin, dest):
    """Find Connection as seen from the receiving endpoint"""
    return next(connection for message in json_data["messages"]
                for connection in message["message"]["Connections"]["connections"]
                if connection["localEndpoint"]["endpoint"]["port"] == port and
                connection["localEndpoint"]["endpoint"]["ip"]["address"] == dest and
                connection["remoteEndpoint"]["endpoint"]["ip"]["address"] == origin
                )


def test_created_connection_after_start_with_metrics(host):
    url = "http://localhost:7070/api/topic/sts_correlate_endpoints?limit=1000"

    conn_port = int(host.ansible("include_vars", "./common_vars.yml")["ansible_facts"]
                    ["test_connection_port_after_start"])

    ubuntu_public_ip = _get_instance_config("agent-ubuntu")["address"]
    ubuntu_private_ip = _get_instance_config("agent-ubuntu")["private_address"]
    fedora_public_ip = _get_instance_config("agent-fedora")["address"]
    fedora_private_ip = _get_instance_config("agent-fedora")["private_address"]
    windows_public_ip = _get_instance_config("agent-win")["address"]
    windows_private_ip = _get_instance_config("agent-win")["private_address"]

    def wait_for_connection():
        data = host.check_output("curl %s" % url)
        json_data = json.loads(data)
        print(json.dumps(json_data))

        outgoing_conn = _find_outgoing_connection(json_data, conn_port, fedora_private_ip, ubuntu_public_ip)
        print outgoing_conn
        assert outgoing_conn["direction"] == "OUTGOING"
        assert outgoing_conn["connectionType"] == "TCP"
        assert outgoing_conn["bytesSentPerSecond"] > 10.0
        assert outgoing_conn["bytesReceivedPerSecond"] == 0.0

        incoming_conn = _find_incoming_connection(json_data, conn_port, fedora_public_ip, ubuntu_private_ip)
        print incoming_conn
        assert incoming_conn["direction"] == "INCOMING"
        assert incoming_conn["connectionType"] == "TCP"
        assert incoming_conn["bytesSentPerSecond"] == 0.0
        assert incoming_conn["bytesReceivedPerSecond"] > 10.0

        outgoing_conn = _find_outgoing_connection(json_data, conn_port, windows_private_ip, ubuntu_public_ip)
        print outgoing_conn
        assert outgoing_conn["direction"] == "OUTGOING"
        assert outgoing_conn["connectionType"] == "TCP"
        assert outgoing_conn["bytesSentPerSecond"] == 0.0    # We don't collect metrics on Windows
        assert outgoing_conn["bytesReceivedPerSecond"] == 0.0

        incoming_conn = _find_incoming_connection(json_data, conn_port, windows_public_ip, ubuntu_private_ip)
        print incoming_conn
        assert incoming_conn["direction"] == "INCOMING"
        assert incoming_conn["connectionType"] == "TCP"
        assert incoming_conn["bytesSentPerSecond"] == 0.0      # TODO why > 0
        assert incoming_conn["bytesReceivedPerSecond"] > 10.0  # TODO why = 0

    util.wait_until(wait_for_connection, 30, 3)


def test_created_connection_before_start(host):
    url = "http://localhost:7070/api/topic/sts_correlate_endpoints?limit=1000"

    conn_port = int(host.ansible("include_vars", "./common_vars.yml")["ansible_facts"]
                    ["test_connection_port_before_start"])

    ubuntu_public_ip = _get_instance_config("agent-ubuntu")["address"]
    ubuntu_private_ip = _get_instance_config("agent-ubuntu")["private_address"]
    fedora_public_ip = _get_instance_config("agent-fedora")["address"]
    fedora_private_ip = _get_instance_config("agent-fedora")["private_address"]
    windows_public_ip = _get_instance_config("agent-win")["address"]
    windows_private_ip = _get_instance_config("agent-win")["private_address"]

    def wait_for_connection():
        data = host.check_output("curl %s" % url)
        json_data = json.loads(data)
        print(json.dumps(json_data))

        outgoing_conn = _find_outgoing_connection(json_data, conn_port, fedora_private_ip, ubuntu_public_ip)
        print outgoing_conn
        assert outgoing_conn["direction"] == "NONE"          # Outgoing gets no direction from Linux /proc scanning
        assert outgoing_conn["connectionType"] == "TCP"

        incoming_conn = _find_incoming_connection(json_data, conn_port, fedora_public_ip, ubuntu_private_ip)
        print incoming_conn
        assert incoming_conn["direction"] == "INCOMING"
        assert incoming_conn["connectionType"] == "TCP"

        outgoing_conn = _find_outgoing_connection(json_data, conn_port, windows_private_ip, ubuntu_public_ip)
        print outgoing_conn
        assert outgoing_conn["direction"] == "OUTGOING"
        assert outgoing_conn["connectionType"] == "TCP"

        incoming_conn = _find_incoming_connection(json_data, conn_port, windows_public_ip, ubuntu_private_ip)
        print incoming_conn
        assert incoming_conn["direction"] == "INCOMING"
        assert incoming_conn["connectionType"] == "TCP"

    util.wait_until(wait_for_connection, 30, 3)


def test_host_metrics(host):
    url = "http://localhost:7070/api/topic/sts_metrics?limit=1000"

    def wait_for_metrics():
        data = host.check_output("curl %s" % url)
        json_data = json.loads(data)
        print(json.dumps(json_data))

        metrics = {message["message"]["Metric"]["name"]: value["value"]
                   for message in json_data["messages"]
                   for value in message["message"]["Metric"]["value"]
                   }

        print metrics

        # These values are based on an ec2 micro instance
        # (as created by molecule.yml)

        # Same metrics we check in the backend e2e tests
        # https://stackvista.githost.io/StackVista/StackState/blob/master/stackstate-pm-test/src/test/scala/com/stackstate/it/e2e/ProcessAgentIntegrationE2E.scala#L17

        # No swap in these tests, we still wanna know whether it is reported
        assert metrics["system.swap.total"] == 0.0
        assert metrics["system.swap.pct_free"] == 1.0

        # Memory
        assert metrics["system.mem.total"] > 900.0
        assert metrics["system.mem.usable"] > 500.0
        assert metrics["system.mem.usable"] < 1000.0
        assert metrics["system.mem.pct_usable"] > 0.5
        assert metrics["system.mem.pct_usable"] < 1.0

        # Load
        assert metrics["system.load.norm.1"] > 0.0

        # CPU
        assert metrics["system.cpu.idle"] > 0.0
        assert metrics["system.cpu.iowait"] > 0.0
        assert metrics["system.cpu.system"] > 0.0
        assert metrics["system.cpu.user"] > 0.0

        # Inodes
        assert metrics["system.fs.file_handles.in_use"] > 0.0
        assert metrics["system.fs.file_handles.max"] > 10000.0

    util.wait_until(wait_for_metrics, 30, 3)


def test_process_metrics(host):
    url = "http://localhost:7070/api/topic/sts_multi_metrics?limit=1000"

    def wait_for_metrics():
        data = host.check_output("curl %s" % url)
        json_data = json.loads(data)
        print(json.dumps(json_data))

        metrics = next(set(message["message"]["MultiMetric"]["values"].keys())
                       for message in json_data["messages"]
                       if message["message"]["MultiMetric"]["name"] == "processMetrics"
                       )
        print metrics

        # Same metrics we check in the backend e2e tests
        # https://stackvista.githost.io/StackVista/StackState/blob/master/stackstate-pm-test/src/test/scala/com/stackstate/it/e2e/ProcessAgentIntegrationE2E.scala#L17

        expected = {"cpu_nice", "cpu_userPct", "cpu_userTime", "cpu_systemPct", "cpu_numThreads", "io_writeRate",
                    "io_writeBytesRate", "cpu_totalPct", "voluntaryCtxSwitches", "mem_dirty", "involuntaryCtxSwitches",
                    "io_readRate", "openFdCount", "mem_shared", "cpu_systemTime", "io_readBytesRate", "mem_data",
                    "mem_vms", "mem_lib", "mem_text", "mem_swap", "mem_rss"}

        assert metrics == expected

    util.wait_until(wait_for_metrics, 30, 3)
