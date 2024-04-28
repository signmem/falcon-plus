CREATE DATABASE vip_falcon_aggregator
  DEFAULT CHARACTER SET utf8
  DEFAULT COLLATE utf8_general_ci;
USE vip_falcon_aggregator;
SET NAMES utf8;

CREATE TABLE `host_metrics` (
 	`id`			bigint unsigned		NOT NULL 	AUTO_INCREMENT					COMMENT 'id',
 	`hostname`		varchar(255)		NOT NULL 	DEFAULT ''						COMMENT 'host name',
 	`metric`		varchar(128)		NOT NULL 	DEFAULT ''						COMMENT 'falcon metric name',
 	`tags`			varchar(255)		NOT NULL 	DEFAULT ''						COMMENT 'falcon tags',
 	`dstype`		varchar(16)			NOT NULL 	DEFAULT 'GAUGE'					COMMENT 'metric type, GAUGE|COUNTER|DERIVE',
 	`step`			int(11)				NOT NULL 	DEFAULT 60 						COMMENT 'interval, in second',
 	`create_time` 	timestamp 			NOT NULL 	DEFAULT '0000-00-00 00:00:00'	COMMENT 'create time',
 	`update_time`	timestamp 			NOT NULL 	DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP	COMMENT 'last modify time',
 	`is_deleted` 	tinyint 			NOT NULL 	DEFAULT 0						COMMENT 'delete status',
 	PRIMARY KEY (`id`),
 	UNIQUE KEY `idx_hostname_metric_tags` (`hostname`, `metric`, `tags`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='metric items table of falcon aggregator';

CREATE TABLE `trend` (
	`id`			bigint unsigned		NOT NULL 	AUTO_INCREMENT		COMMENT 'id',
	`metric_id`		bigint unsigned		NOT NULL 	DEFAULT 0			COMMENT 'id in host_metrics table',	
	`ts`			int(11)				NOT NULL 	DEFAULT 0			COMMENT 'timestamp',
	`min`		decimal(26,4)       NOT NULL 	DEFAULT 0.0000		COMMENT 'min value',
	`avg`           decimal(26,4)       NOT NULL 	DEFAULT 0.0000		COMMENT 'average value',
	`max`           decimal(26,4)       NOT NULL 	DEFAULT 0.0000		COMMENT 'max value',
	`num`			integer             NOT NULL 	DEFAULT 0			COMMENT 'the number of value',
	`create_time` 	timestamp 			NOT NULL 	DEFAULT '0000-00-00 00:00:00'	COMMENT 'create time',
	PRIMARY KEY (`id`),
	UNIQUE KEY `idx_metric_id_ts` (`metric_id`, `ts`),
	KEY `idx_ts`(`ts`)
) ENGINE=TokuDB DEFAULT CHARSET=utf8 COMMENT='trend data table of falcon aggregator'
PARTITION BY RANGE (`ts`)
(PARTITION p201801 VALUES LESS THAN (1517414400) ENGINE = TokuDB,
 PARTITION p201802 VALUES LESS THAN (1519833600) ENGINE = TokuDB,
 PARTITION p201803 VALUES LESS THAN (1522512000) ENGINE = TokuDB,
 PARTITION p201804 VALUES LESS THAN (1525104000) ENGINE = TokuDB,	
 PARTITION p201805 VALUES LESS THAN (1527782400) ENGINE = TokuDB,
 PARTITION p201806 VALUES LESS THAN (1530374400) ENGINE = TokuDB,
 PARTITION p201807 VALUES LESS THAN (1533052800) ENGINE = TokuDB,
 PARTITION p201808 VALUES LESS THAN (1535731200) ENGINE = TokuDB，
 PARTITION p201809 VALUES LESS THAN (1538323200) ENGINE = TokuDB,
 PARTITION p201810 VALUES LESS THAN (1541001600) ENGINE = TokuDB,
 PARTITION p201811 VALUES LESS THAN (1543593600) ENGINE = TokuDB,
 PARTITION p201812 VALUES LESS THAN (1546272000) ENGINE = TokuDB);
