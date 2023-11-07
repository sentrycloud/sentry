echo "query metrics" 
curl -d '{"metric": "sentry"}' http://127.0.0.1:51001/server/api/metrics

echo "query tag keys"
curl -d '{"metric": "sentry_sys_mem_usage"}' http://127.0.0.1:51001/server/api/tagKeys

echo "query tag values"
curl -d '{"metric": "sentry_sys_mem_usage", "tags":{"ip":"*"}}' http://127.0.0.1:51001/server/api/tagValues

echo "query curves"
curl -d '{"metric": "sentry_sys_mem_usage", "tags":{"ip":"*"}}' http://127.0.0.1:51001/server/api/curves

echo "query range data"
curl -d '{"token":"123456", "last":3600, "aggregator":"max","down_sample":10, "metrics":[{"metric": "sentry_sys_cpu_usage", "tags":{"ip":"127.0.0.1"}}]}' http://127.0.0.1:51001/server/api/range

echo "query topN data"
curl -d '{"token":"123456", "last":3600, "aggregator":"max","down_sample":10, "limit":10, "order":"desc", "metric": "sentry_sys_cpu_usage", "tags":{"ip":"*"}}' http://127.0.0.1:51001/server/api/topn
