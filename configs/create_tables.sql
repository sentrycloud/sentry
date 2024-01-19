CREATE DATABASE IF NOT EXISTS sentry;
USE sentry;

CREATE TABLE `alarm_contact` (
    `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `is_deleted` tinyint(4) unsigned NOT NULL DEFAULT '0',
    `name` varchar(255) NOT NULL,
    `phone` varchar(255) DEFAULT NULL,
    `mail` varchar(255) DEFAULT NULL,
    `wechat` varchar(255) DEFAULT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_updated` (`updated`),
    KEY `idx_deleted` (`is_deleted`),
    KEY `idx_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

CREATE TABLE `alarm_rule` (
    `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `is_deleted` tinyint(4) unsigned NOT NULL DEFAULT '0',
    `name` varchar(128) DEFAULT '',
    `type` int(6) NOT NULL,
    `query_range` int(11) NOT NULL DEFAULT '60',
    `contacts` varchar(256) NOT NULL,
    `level` int(6) NOT NULL COMMENT 'alarm level: error, warn, info',
    `message` varchar(512) DEFAULT '' COMMENT 'alarm message format',
    `data_source` text NOT NULL,
    `trigger` text NOT NULL,
    PRIMARY KEY (`id`),
    KEY `idx_updated` (`updated`),
    KEY `idx_deleted` (`is_deleted`),
    KEY `idx_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

CREATE TABLE `metric_white_list` (
    `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `is_deleted` tinyint(4) unsigned NOT NULL DEFAULT '0',
    `metric` varchar(255) NOT NULL,
    `creator` varchar(255) NOT NULL DEFAULT '',
    `app_name` varchar(255) NOT NULL DEFAULT '',
    PRIMARY KEY (`id`),
    KEY `idx_deleted` (`is_deleted`),
    KEY `idx_metric` (`metric`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

CREATE TABLE `dashboard` (
    `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `is_deleted` tinyint(4) unsigned NOT NULL DEFAULT '0',
    `name` varchar(128) NOT NULL DEFAULT '',
    `creator` varchar(255) NOT NULL DEFAULT '',
    `app_name` varchar(255) NOT NULL DEFAULT '',
    `chart_layout` varchar(4096) NOT NULL DEFAULT '',
    `tag_filter` varchar(1024) NOT NULL DEFAULT '',
    `saved_status` varchar(1024) NOT NULL DEFAULT '',
     PRIMARY KEY (`id`),
     KEY `idx_deleted` (`is_deleted`),
     KEY `idx_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

CREATE TABLE `chart` (
    `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `is_deleted` tinyint(4) unsigned NOT NULL DEFAULT '0',
    `dashboard_id` int(10) UNSIGNED NOT NULL DEFAULT '0',
    `name` varchar(128) NOT NULL DEFAULT '',
    `type` varchar(32) NOT NULL DEFAULT '',
    `aggregation` varchar(32) NOT NULL DEFAULT '',
    `down_sample` varchar(32) NOT NULL DEFAULT '',
    `topn_limit` int(11) NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`),
    KEY `idx_deleted` (`is_deleted`),
    KEY `idx_deleted_dashboard_id` (`is_deleted`, `dashboard_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

CREATE TABLE `line` (
    `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `is_deleted` tinyint(4) unsigned NOT NULL DEFAULT '0',
    `chart_id` int(10) UNSIGNED NOT NULL DEFAULT '0',
    `name` varchar(128) NOT NULL DEFAULT '',
    `metric` varchar(255) NOT NULL DEFAULT '',
    `tags` varchar(4096) NOT NULL DEFAULT '',
    `offset` int NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`),
    KEY `idx_deleted` (`is_deleted`),
    KEY `idx_deleted_chart_id` (`is_deleted`,`chart_id`),
    KEY `idx_name` (`name`),
    KEY `idx_metric` (`metric`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

INSERT INTO `alarm_contact` (`name`, `phone`, `mail`, `wechat`) VALUES ('eric', '13777820006', 'eric@gmail.com', 'eric@gmail.com');

INSERT INTO `alarm_rule` (`name`, `type`, `query_range`, `contacts`, `level`, `message`, `data_source`, `trigger`)
    VALUES ('memory heartbeat alarm', 0, 60, 'eric', 1, '{time} memory monitor has no data',
            '{"metric": "sentry_sys_mem_usage", "tags":{"ip":"127.0.0.1"}, "aggregation": "max", "down_sample": 10}',
            '{"error_count": 1}');

INSERT INTO `alarm_rule` (`name`, `type`, `query_range`, `contacts`, `level`, `message`, `data_source`, `trigger`)
    VALUES ('memory threshold alarm', 1, 60, 'eric', 1, '{time} memory usage reach {value}',
            '{"metric": "sentry_sys_mem_usage", "tags":{"ip":"*"}, "aggregation": "max", "down_sample": 10}',
            '{"greater_than": 12.0, "less_than": 1.0, "error_count": 2}');

INSERT INTO `alarm_rule` (`name`, `type`, `query_range`, `contacts`, `level`, `message`, `data_source`, `trigger`)
    VALUES ('memory topN alarm', 2, 60, 'eric', 1, '{time} memory usage reach {value}',
            '{"metric": "sentry_sys_mem_usage", "tags":{"ip":"*"}, "aggregation": "max", "down_sample": 10, "sort": "desc", "limit":10}',
            '{"greater_than": 12.0, "error_count": 1}');

INSERT INTO `alarm_rule` (`name`, `type`, `query_range`, `contacts`, `level`, `message`, `data_source`, `trigger`)
    VALUES ('memory compare alarm', 3, 60, 'eric', 1, '{time} memory usage reach {value}, history value: {history.value}, compare value: {compare.value}',
            '{"metric":"sentry_sys_mem_usage", "tags":{"ip":"127.0.0.1"}, "aggregation":"max", "down_sample":10, "compare_type":0, "compare_days_ago":0, "compare_seconds":60}',
            '{"less_than": -8.0, "greater_than": 8.0, "error_count": 1}');

/* create dashboard for sentry_server monitor */
INSERT INTO `dashboard` (`name`, `creator`, `app_name`, `chart_layout`, `tag_filter`)
    VALUES ('sentry_server monitor', 'eric', 'sentry_server',
            '[{"w":6,"h":4,"x":0,"y":0,"i":"1","moved":false,"static":true},{"w":6,"h":4,"x":6,"y":0,"i":"2","moved":false,"static":true},{"w":4,"h":4,"x":0,"y":4,"i":"3","moved":false,"static":true},{"w":4,"h":4,"x":4,"y":4,"i":"4","moved":false,"static":true},{"w":4,"h":4,"x":8,"y":4,"i":"5","moved":false,"static":true},{"w":4,"h":4,"x":0,"y":8,"i":"6","moved":false,"static":true},{"w":4,"h":4,"x":4,"y":8,"i":"7","moved":false,"static":true},{"w":4,"h":4,"x":8,"y":8,"i":"8","moved":false,"static":true}]',
            '{"metric":"sentry_server_http_qps","tags":["sentryIP"]}');

INSERT INTO `chart` (`dashboard_id`, `name`, `type`, `aggregation`, `down_sample`, `topn_limit`)
    VALUES (1, 'http QPS', 'line', 'sum', '10s', 10);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (1, 'qps', 'sentry_server_http_qps', '{"api":"*"}', 0);

INSERT INTO `chart` (`dashboard_id`, `name`, `type`, `aggregation`, `down_sample`, `topn_limit`)
    VALUES (1, 'http RT', 'line', 'avg', '10s', 10);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (2, 'rt', 'sentry_server_http_rt', '{"api":"*"}', 0);

INSERT INTO `chart` (`dashboard_id`, `name`, `type`, `aggregation`, `down_sample`, `topn_limit`)
    VALUES (1, 'goroutine number', 'line', 'avg', '10s', 10);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (3, 'number', 'sentry_go_num', '{}', 0);

INSERT INTO `chart` (`dashboard_id`, `name`, `type`, `aggregation`, `down_sample`, `topn_limit`)
    VALUES (1, 'gc number', 'line', 'avg', '10s', 10);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (4, 'number', 'sentry_gc_num', '{}', 0);

INSERT INTO `chart` (`dashboard_id`, `name`, `type`, `aggregation`, `down_sample`, `topn_limit`)
    VALUES (1, 'gc pause time(ms)', 'line', 'avg', '10s', 10);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (5, 'pause', 'sentry_gc_pause', '{}', 0);

INSERT INTO `chart` (`dashboard_id`, `name`, `type`, `aggregation`, `down_sample`, `topn_limit`)
    VALUES (1, 'agent number', 'line', 'sum', '10s', 10);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (6, 'agent number', 'sentry_server_agent_count', '{}', 0);

INSERT INTO `chart` (`dashboard_id`, `name`, `type`, `aggregation`, `down_sample`, `topn_limit`)
    VALUES (1, 'data points QPS', 'line', 'sum', '10s', 10);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (7, 'qps', 'sentry_server_data_point', '{}', 0);

INSERT INTO `chart` (`dashboard_id`, `name`, `type`, `aggregation`, `down_sample`, `topn_limit`)
    VALUES (1, 'merge chan size', 'line', 'max', '10s', 10);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (8, 'chan size', 'sentry_server_chan_size', '{"chan":"*"}', 0);

/* create dashboard for machine monitor (cpu, memory, disk, network ...) */
INSERT INTO `dashboard` (`name`, `creator`, `app_name`, `chart_layout`, `tag_filter`)
    VALUES ('machine monitor', 'eric', 'sentry_agent',
            '[{"w":4,"h":4,"x":0,"y":0,"i":"9","moved":false,"static":true},{"w":4,"h":4,"x":4,"y":0,"i":"10","moved":false,"static":true},{"w":4,"h":4,"x":8,"y":0,"i":"11","moved":false,"static":true},{"w":4,"h":4,"x":0,"y":4,"i":"12","moved":false,"static":true},{"w":4,"h":4,"x":4,"y":4,"i":"13","moved":false,"static":true},{"w":4,"h":4,"x":8,"y":4,"i":"14","moved":false,"static":true},{"w":4,"h":4,"x":0,"y":8,"i":"15","moved":false,"static":true},{"w":4,"h":4,"x":4,"y":8,"i":"16","moved":false,"static":true},{"w":4,"h":4,"x":8,"y":8,"i":"17","moved":false,"static":true},{"w":6,"h":4,"x":0,"y":12,"i":"18","moved":false,"static":true},{"w":6,"h":4,"x":6,"y":12,"i":"19","moved":false,"static":true}]',
            '{"metric":"sentry_sys_cpu_usage","tags":["sentryIP"]}');

INSERT INTO `chart` (`dashboard_id`, `name`, `type`, `aggregation`, `down_sample`, `topn_limit`)
    VALUES (2, 'cpu usage', 'line', 'max', '10s', 10);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (9, 'cpu', 'sentry_sys_cpu_usage', '{}', 0);

INSERT INTO `chart` (`dashboard_id`, `name`, `type`, `aggregation`, `down_sample`, `topn_limit`)
    VALUES (2, 'load average', 'line', 'max', '10s', 10);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (10, 'load', 'sentry_sys_load_average', '{}', 0);

INSERT INTO `chart` (`dashboard_id`, `name`, `type`, `aggregation`, `down_sample`, `topn_limit`)
    VALUES (2, 'memory usage', 'line', 'max', '10s', 10);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (11, 'memory', 'sentry_sys_mem_usage', '{}', 0);

INSERT INTO `chart` (`dashboard_id`, `name`, `type`, `aggregation`, `down_sample`, `topn_limit`)
    VALUES (2, 'disk usage', 'line', 'max', '10s', 10);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (12, 'disk usage', 'sentry_sys_disk_usage', '{"device":"*"}', 0);

INSERT INTO `chart` (`dashboard_id`, `name`, `type`, `aggregation`, `down_sample`, `topn_limit`)
    VALUES (2, 'disk io wait', 'line', 'max', '10s', 10);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (13, 'io wait', 'sentry_sys_io_wait', '{}', 0);

INSERT INTO `chart` (`dashboard_id`, `name`, `type`, `aggregation`, `down_sample`, `topn_limit`)
    VALUES (2, 'disk io util', 'line', 'max', '10s', 10);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (14, 'io util', 'sentry_sys_io_util', '{}', 0);

INSERT INTO `chart` (`dashboard_id`, `name`, `type`, `aggregation`, `down_sample`, `topn_limit`)
    VALUES (2, 'disk io stats', 'line', 'max', '10s', 10);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (15, 'read bytes/s', 'sentry_sys_io_read', '{}', 0);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (15, 'write bytes/s', 'sentry_sys_io_write', '{}', 0);

INSERT INTO `chart` (`dashboard_id`, `name`, `type`, `aggregation`, `down_sample`, `topn_limit`)
    VALUES (2, 'net stats in bytes', 'line', 'max', '10s', 10);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (16, 'bytes sent/s', 'sentry_sys_net_bytes_sent', '{}', 0);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (16, 'bytes recv/s', 'sentry_sys_net_bytes_recv', '{}', 0);

INSERT INTO `chart` (`dashboard_id`, `name`, `type`, `aggregation`, `down_sample`, `topn_limit`)
    VALUES (2, 'net stats in packets', 'line', 'max', '10s', 10);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (17, 'packets sent/s', 'sentry_sys_net_packets_sent', '{}', 0);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (17, 'packets recv/s', 'sentry_sys_net_packets_recv', '{}', 0);

INSERT INTO `chart` (`dashboard_id`, `name`, `type`, `aggregation`, `down_sample`, `topn_limit`)
    VALUES (2, 'tcp status', 'line', 'max', '10s', 10);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (18, 'status', 'sentry_sys_tcp_status', '{"status":"*"}', 0);

INSERT INTO `chart` (`dashboard_id`, `name`, `type`, `aggregation`, `down_sample`, `topn_limit`)
    VALUES (2, 'process number', 'line', 'max', '10s', 10);
INSERT INTO `line` (`chart_id`, `name`, `metric`, `tags`, `offset`)
    VALUES (19, 'number', 'sentry_sys_process_number', '{}', 0);
