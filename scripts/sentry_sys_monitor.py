import sys
import time
import psutil

METRIC_FORMAT = '"metric":"{}", "tags":{},"timestamp":{}, "value":{}'


def sentry_time():
    print("10")


def collect():
    ts = round(int(time.time()) / 10) * 10  # round to cron script interval
    print('[{},{}]'.format(collect_cpu(ts), collect_memory(ts)))


def collect_cpu(ts):
    cpu_percent = psutil.cpu_percent(interval=0.5)
    metric = METRIC_FORMAT.format("sentry_sys_cpu_usage", "{}", ts, cpu_percent)
    metric = "{" + metric + "}"
    return metric


def collect_memory(ts):
    mem_percent = psutil.virtual_memory().percent
    metric = METRIC_FORMAT.format("sentry_sys_mem_usage", "{}", ts, mem_percent)
    metric = "{" + metric + "}"
    return metric


if __name__ == "__main__":
    if len(sys.argv) == 2 and sys.argv[1] == "sentry_time":
        sentry_time()
    else:
        collect()

