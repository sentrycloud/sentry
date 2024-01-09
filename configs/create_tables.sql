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
