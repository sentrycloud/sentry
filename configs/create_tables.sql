CREATE TABLE `alarm_contact` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
    `name` varchar(255) NOT NULL,
    `phone` varchar(255) DEFAULT NULL,
    `mail` varchar(255) DEFAULT NULL,
    `wechat` varchar(255) DEFAULT NULL,
    `deleted` tinyint(1) NOT NULL DEFAULT '0' COMMENT 'has this contact be deleted',
    `created` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'create time',
    `updated` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT 'update time',
    PRIMARY KEY (`id`),
    KEY `idx_name` (`name`),
    KEY `idx_created` (`created`)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;

CREATE TABLE `alarm_rule` (
    `id` int(11) NOT NULL AUTO_INCREMENT,
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
