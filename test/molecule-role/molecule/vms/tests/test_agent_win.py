import os
import re
import util
from testinfra.utils.ansible_runner import AnsibleRunner

testinfra_hosts = AnsibleRunner(os.environ["MOLECULE_INVENTORY_FILE"]).get_hosts("agent_win_vm")


def test_stackstate_agent_is_installed(host, ansible_var):
    pkg = "StackState Agent"
    # res = host.ansible("win_shell", "Get-Package \"{}\"".format(pkg), check=False)
    res = host.ansible("win_shell", " Get-WmiObject -Class Win32_Product | where name -eq \"{}\" | select Name, Version ".format(pkg), check=False)
    print(res)
    expected_major_version = ansible_var("major_version")
    assert re.search(".*{} {}\\.".format(pkg, expected_major_version), res["stdout"], re.I)


def test_stackstate_agent_running_and_enabled(host):
    def check(name, deps, depended_by):
        service = host.ansible("win_service", "name={}".format(name))
        print(service)
        assert service["exists"]
        assert not service["changed"]
        assert service["state"] == "running"
        assert service["dependencies"] == deps
        assert service["depended_by"] == depended_by

    check("stackstateagent", [], ["stackstate-process-agent", "stackstate-trace-agent"])
    check("stackstate-trace-agent", ["stackstateagent"], [])
    check("stackstate-process-agent", ["stackstateagent"], [])


def test_stackstate_agent_log(host, hostname):
    agent_log_path = "c:\\programdata\\stackstate\\logs\\agent.log"

    # Check for presence of success
    def wait_for_check_successes():
        agent_log = host.ansible("win_shell", "cat \"{}\"".format(agent_log_path), check=False)["stdout"]
        print(agent_log)
        assert re.search("Successfully posted payload to.*stsAgent/intake", agent_log)

    util.wait_until(wait_for_check_successes, 30, 3)

    agent_log = host.ansible("win_shell", "cat \"{}\"".format(agent_log_path), check=False)["stdout"]
    with open("./{}-agent.log".format(hostname), 'wb') as f:
        f.write(agent_log.encode('utf-8'))

    # Check for errors
    for line in agent_log.splitlines():
        print("Considering: %s" % line)
        assert not re.search("\\| error \\|", line, re.IGNORECASE)


def test_stackstate_process_agent_no_log_errors(host, hostname):
    process_agent_log_path = "c:\\programdata\\stackstate\\logs\\process-agent.log"

    # Check for presence of success
    def wait_for_check_successes():
        process_agent_log = host.ansible("win_shell", "cat \"{}\"".format(process_agent_log_path), check=False)["stdout"]
        print(process_agent_log)

        assert re.search("Finished check #1", process_agent_log)
        assert re.search("starting network tracer locally", process_agent_log)

    util.wait_until(wait_for_check_successes, 30, 3)

    process_agent_log = host.ansible("win_shell", "cat \"{}\"".format(process_agent_log_path), check=False)["stdout"]
    with open("./{}-process.log".format(hostname), 'wb') as f:
        f.write(process_agent_log.encode('utf-8'))

    # Check for errors
    for line in process_agent_log.splitlines():
        print("Considering: %s" % line)
        assert not re.search("error", line, re.IGNORECASE)


def test_stackstate_trace_agent_log(host, hostname):
    trace_agent_log_path = "c:\\programdata\\stackstate\\logs\\trace-agent.log"

    # Check for presence of success
    def wait_for_check_successes():
        trace_agent_log = host.ansible("win_shell", "cat \"{}\"".format(trace_agent_log_path), check=False)["stdout"]
        print(trace_agent_log)
        assert re.search("Trace agent running on host", trace_agent_log)
        assert re.search("Listening for traces at", trace_agent_log)
        assert re.search("No data received", trace_agent_log)

    util.wait_until(wait_for_check_successes, 30, 3)

    agent_log = host.ansible("win_shell", "cat \"{}\"".format(trace_agent_log_path), check=False)["stdout"]
    with open("./{}-trace.log".format(hostname), 'wb') as f:
        f.write(agent_log.encode('utf-8'))

    # Check for errors
    for line in agent_log.splitlines():
        print("Considering: %s" % line)
        assert not re.search("\\| error \\|", line, re.IGNORECASE)
