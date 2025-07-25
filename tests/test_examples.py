import argparse
import subprocess
import time
import os
import threading

LOG_LEVEL = "debug"
START_TIMEOUT_S = 2
STOP_TIMEOUT_S = 5
WAIT_MESSAGE_TIMEOUT_S = 10
ERROR_MESSAGE = "ERROR"
STDBUF_CMD = ["stdbuf", "-o0"]
ZENOH_PORT = "7447"
DEFAULT_LISTEN_LOCATOR = f'tcp/0.0.0.0:{ZENOH_PORT}'
DEFAULT_CONNECT_LOCATOR = f'tcp/127.0.0.1:{ZENOH_PORT}'


def run_process(command, output_collector, process_list):
    env = os.environ.copy()
    env["RUST_LOG"] = LOG_LEVEL

    print(f"Run {command}")
    try:
        process = subprocess.Popen(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, env=env)
    except FileNotFoundError:
        print(f"Error: Command '{command[0]}' not found.")
        return
    process_list.append(process)
    for line in iter(process.stdout.readline, ''):
        print(f"-- [{process.pid}]:", line.strip())
        output_collector.append(line.strip())
    process.stdout.close()
    process.wait()

    if process.returncode != 0:
        stderr_output = process.stderr.read().strip()
        if stderr_output != "":
            raise Exception(f"[{process.pid}]: Terminated with error: {stderr_output}")


def run_background(command, output_collector, process_list):
    thread = threading.Thread(target=run_process, args=(command, output_collector, process_list))
    thread.start()


def terminate_processes(process_list):
    for process in process_list:
        process.terminate()
        try:
            process.wait(timeout=STOP_TIMEOUT_S)
        except subprocess.TimeoutExpired:
            process.kill()
    process_list.clear()


def wait_messages(client_output, messages):
    start_time = time.time()
    while time.time() - start_time < WAIT_MESSAGE_TIMEOUT_S:
        
        if all(any((message in line) for line in client_output) for message in messages):
            return True
        time.sleep(1)
    return False


def check_errors(output):
    for line in output:
        if ERROR_MESSAGE in line:
            print(line)
            raise Exception("Router have an error.")


def test_examples(peer_command, peer_expect, client_command, client_expect):
    print(f"Test pair: \"{' '.join(peer_command)}\", \"{' '.join(client_command)}\"")
    peer_output = []
    client_output = []
    process_list = []
    try:
        run_background(peer_command, peer_output, process_list)
        time.sleep(START_TIMEOUT_S)
        run_background(client_command, client_output, process_list)

        if wait_messages(peer_output, peer_expect):
            print(f'{peer_command} got "{peer_expect}"')
        else:
            raise Exception(f'{peer_command} FAILED to get "{peer_expect}"')

        if wait_messages(client_output, client_expect):
            print(f'{client_command} got "{client_expect}"')
        else:
            raise Exception(f'{client_command} FAILED to get "{client_expect}"')

        check_errors(peer_output)
        check_errors(client_output)
    finally:
        terminate_processes(process_list)


def __parse_args():
    parser = argparse.ArgumentParser(description="Parse command line arguments.")

    parser.add_argument('-l', '--listen', type=str, default=DEFAULT_LISTEN_LOCATOR, help='The listen locator address')
    parser.add_argument('-e', '--connect', type=str, default=DEFAULT_CONNECT_LOCATOR, help='The connect locator address')
    parser.add_argument('-c', '--config', type=str, help='The path to the config file')
    parser.add_argument('examples_dir', type=str, help='The directory containing examples')

    return parser.parse_args()


def main():
    args = __parse_args()

    config = []
    if args.config is not None:
        config = ["-c", args.config]

    def build_peer_cmd(cmd):
        return STDBUF_CMD + [os.path.join(args.examples_dir, cmd), "-l", args.listen, "-m", "peer"] + config

    def build_client_cmd(cmd):
        return STDBUF_CMD + [os.path.join(args.examples_dir, cmd), "-e", args.connect, "-m", "client"] + config

    test_examples(
        build_peer_cmd("z_sub"),
        ["Received PUT ('demo/example/zenoh-go-put"],
        build_client_cmd("z_put"),
        ["Putting Data ('demo/example/zenoh-go-put'"],
    )

    test_examples(
        build_peer_cmd("z_sub"),
        ["Received DELETE ('demo/example/zenoh-go-put'"],
        build_client_cmd("z_delete"),
        ["Deleting resources matching 'demo/example/zenoh-go-put'"],
    )

    test_examples(
        build_peer_cmd("z_pub"),
        ["Putting Data ('demo/example/zenoh-go-pub'"],
        build_client_cmd("z_sub"),
        ["Received PUT ('demo/example/zenoh-go-pub"],
    )

    test_examples(
        build_peer_cmd("z_queryable"),
        ["Responding ('demo/example/zenoh-go-queryable': 'Queryable from Go!')"],
        build_client_cmd("z_get"),
        ["Received ('demo/example/zenoh-go-queryable': 'Queryable from Go!')"],
    )

    test_examples(
        build_peer_cmd("z_sub_liveliness"),
        ["New alive token ('group1/**')"],
        build_client_cmd("z_liveliness"),
        ["Declaring liveliness token 'group1/**'"],
    )

    test_examples(
        build_peer_cmd("z_liveliness"),
        ["Declaring liveliness token 'group1/**'"],
        build_client_cmd("z_get_liveliness"),
        ["Alive token ('group1/**')"],
    )

    test_examples(
        build_peer_cmd("z_queryable"),
        [
            "Received Query 'demo/example/**?' with value '[   1] '",
            "Responding ('demo/example/zenoh-go-queryable': 'Queryable from Go!')"
        ],
        build_client_cmd("z_querier"),
        [
            "Querying 'demo/example/**' with payload '[   1] '...",
            "Received ('demo/example/zenoh-go-queryable': 'Queryable from Go!')"
        ]
    )


if __name__ == "__main__":
    main()
