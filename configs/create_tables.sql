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
    KEY `idx_deleted` (`is_deleted`),
    KEY `idx_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

CREATE TABLE `alarm_rule` (
    `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` varchar(128) DEFAULT '',
    `type` int(6) NOT NULL,
    `query_range` int(11) NOT NULL DEFAULT '60',
    `contacts` varchar(256) NOT NULL,
    `level` int(6) NOT NULL COMMENT 'alarm level: error, warn, info',
    `message` varchar(512) DEFAULT '' COMMENT 'alarm message format',
    `data_source` text NOT NULL,
    `trigger` text NOT NULL,
    `deleted` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'has this rule be deleted',
    `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'create time',
    `updated` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'update time',
    PRIMARY KEY (`id`),
    KEY `idx_name` (`name`),
    KEY `idx_deleted` (`deleted`),
    KEY `idx_created` (`created`),
    KEY `idx_updated` (`updated`)
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
    `name` varchar(128) NOT NULL DEFAULT '',
    `type` tinyint(2) NOT NULL DEFAULT '0',
    `aggregation` varchar(32) NOT NULL DEFAULT '',
    `down_sample` varchar(32) NOT NULL DEFAULT '',
    `topn_limit` int(11) NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`),
    KEY `idx_deleted` (`is_deleted`),
    KEY `idx_name` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

CREATE TABLE `line` (
    `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `is_deleted` tinyint(4) unsigned NOT NULL DEFAULT '0',
    `name` varchar(128) NOT NULL DEFAULT '',
    `metric` varchar(255) NOT NULL DEFAULT '',
    `tags` varchar(1024) NOT NULL DEFAULT '',
    PRIMARY KEY (`id`),
    KEY `idx_deleted` (`is_deleted`),
    KEY `idx_name` (`name`),
    KEY `idx_metric` (`metric`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

CREATE TABLE `dashboard_chart_relation` (
    `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `is_deleted` tinyint(4) unsigned NOT NULL DEFAULT '0',
    `dashboard_id` int(10) UNSIGNED NOT NULL DEFAULT '0',
    `chart_id` int(10) UNSIGNED NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`),
    KEY `idx_deleted` (`is_deleted`),
    KEY `idx_dashboard_id` (`dashboard_id`),
    KEY `idx_chart_id` (`chart_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

CREATE TABLE `chart_line_relation` (
    `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT,
    `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `updated` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    `is_deleted` tinyint(4) unsigned NOT NULL DEFAULT '0',
    `dashboard_id` int(10) UNSIGNED NOT NULL DEFAULT '0',
    `chart_id` int(10) UNSIGNED NOT NULL DEFAULT '0',
    PRIMARY KEY (`id`),
    KEY `idx_deleted` (`is_deleted`),
    KEY `idx_dashboard_id` (`dashboard_id`),
    KEY `idx_chart_id` (`chart_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

INSERT INTO `alarm_contact` (`name`, `phone`, `mail`, `wechat`) VALUES ('eric', '13777820006', 'eric@gmail.com', 'eric@gmail.com');

INSERT INTO `alarm_rule` (`name`, `type`, `query_range`, `contacts`, `level`, `message`, `data_source`, `trigger`)
    VALUES ('memory usage alarm', 0, 60, 'eric', 1, '{time} memory monitor has no data',
            '{"metric": "sentry_sys_mem_usage", "tags":{"ip":"127.0.0.1"}, "aggregator": "max", "down_sample": 10}',
            '{"error_count": 1}');

INSERT INTO `alarm_rule` (`name`, `type`, `query_range`, `contacts`, `level`, `message`, `data_source`, `trigger`)
    VALUES ('memory usage alarm', 1, 60, 'eric', 1, '{time} memory usage reach {value}',
            '{"metric": "sentry_sys_mem_usage", "tags":{"ip":"*"}, "aggregator": "max", "down_sample": 10}',
            '{"greater_than": 12.0, "less_than": 1.0, "error_count": 2}');

INSERT INTO `alarm_rule` (`name`, `type`, `query_range`, `contacts`, `level`, `message`, `data_source`, `trigger`)
    VALUES ('memory usage alarm', 2, 60, 'eric', 1, '{time} memory usage reach {value}',
            '{"metric": "sentry_sys_mem_usage", "tags":{"ip":"*"}, "aggregator": "max", "down_sample": 10, "sort": "desc", "limit":10}',
            '{"greater_than": 12.0, "error_count": 1}');

INSERT INTO `alarm_rule` (`name`, `type`, `query_range`, `contacts`, `level`, `message`, `data_source`, `trigger`)
    VALUES ('memory usage alarm', 3, 60, 'eric', 1, '{time} memory usage reach {value}, history value: {history.value}, compare value: {compare.value}',
            '{"metric":"sentry_sys_mem_usage", "tags":{"ip":"127.0.0.1"}, "aggregator":"max", "down_sample":10, "compare_type":0, "compare_days_ago":0, "compare_seconds":60}',
            '{"less_than": -8.0, "greater_than": 8.0, "error_count": 1}');
