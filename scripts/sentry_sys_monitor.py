import json
import subprocess
import sys
import time
import psutil

METRIC_FORMAT = '"metric":"{}", "tags":{},"timestamp":{}, "value":{}'
TCP_STATUS_SET = ["ESTABLISHED", "SYN_SENT", "SYN_RECV", "FIN_WAIT1", "FIN_WAIT2", "TIME_WAIT", "CLOSE_WAIT",
                  "LAST_ACK", "LISTEN", "CLOSING", "CLOSED"]


def format_metric(metric, ts, value, tags=None):
    if tags is None:
        tags = {}
    metric = METRIC_FORMAT.format(metric, json.dumps(tags), ts, value)
    return "{" + metric + "}"


def run_shell(cmd):
    try:
        output = subprocess.check_output(cmd, shell=True)
        return output.decode("utf-8")
    except Exception as e:
        return ""


def sentry_time():
    print("10")


def collect():
    ts = int(round(int(time.time()) / 10) * 10)  # round to cron script interval
    metrics = []
    collect_cpu(metrics, ts)
    collect_load_avg(metrics, ts)
    collect_memory(metrics, ts)
    collect_disk_usage(metrics, ts)
    collect_net_io(metrics, ts)
    collect_tcp_status(metrics, ts)
    print("[" + ",".join(metrics) + "]")


def collect_cpu(metrics, ts):
    cpu_percent = psutil.cpu_percent(interval=0.5)
    metrics.append(format_metric("sentry_sys_cpu_usage", ts, cpu_percent))


def collect_load_avg(metrics, ts):
    load_avg = psutil.getloadavg()
    metrics.append(format_metric("sentry_sys_load_average", ts, load_avg[0]))  # load average for last 1 minute


def collect_memory(metrics, ts):
    mem_percent = psutil.virtual_memory().percent
    metrics.append(format_metric("sentry_sys_mem_usage", ts, mem_percent))


def collect_disk_usage(metrics, ts):
    partitions = psutil.disk_partitions(False)
    for disk in partitions:
        if disk.device != "overlay" and disk.device != "tmpfs" and disk.device != "shm":
            disk_usage = psutil.disk_usage(disk.mountpoint)
            metrics.append(format_metric("sentry_sys_disk_usage", ts, disk_usage.percent, {"device": disk.device}))


def collect_net_io(metrics, ts):
    netio = psutil.net_io_counters()
    metrics.append(format_metric("sentry_sys_net_bytes_sent", ts, netio.bytes_sent))
    metrics.append(format_metric("sentry_sys_net_bytes_recv", ts, netio.bytes_recv))


def collect_tcp_status(metrics, ts):
    result = run_shell("netstat -an | grep tcp | awk '{print $NF}' | sort | uniq -c")
    if result == "":
        return
    lines = result.split("\n")
    for line in lines:
        line = line.strip()
        if line != "":
            tcp_status = line.split(" ")
            if len(tcp_status) == 2 and tcp_status[1] in TCP_STATUS_SET:
                metrics.append(format_metric("sentry_sys_tcp_status", ts, tcp_status[0], {"status": tcp_status[1]}))


# psutil document: https://psutil.readthedocs.io/en/latest/
if __name__ == "__main__":
    if len(sys.argv) == 2 and sys.argv[1] == "sentry_time":
        sentry_time()
    else:
        collect()
