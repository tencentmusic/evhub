

CREATE DATABASE IF NOT EXISTS evhub;

USE evhub;

CREATE TABLE IF NOT EXISTS tx_event
(
    `sys_ctime`             DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ,
    `sys_utime`             DATETIME    NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP ,
    `event_id`              char(18) NOT NULL ,
    `app_id`                VARCHAR(128) NOT NULL COMMENT 'appID',
    `topic_id`              VARCHAR(128) NOT NULL COMMENT 'topicID',
    `retry_count`           INT(11)     NOT NULL DEFAULT 0 ,
    `cb_cf_type`            TINYINT     NOT NULL DEFAULT 0 ,
    `cb_pt_type`            TINYINT    NOT NULL DEFAULT 0 ,
    `cb_addr`               VARCHAR(128) NOT NULL DEFAULT 0 ,
    `cb_interval`           BIGINT(20)  ,
    `cb_timeout`            BIGINT(20)  ,
    `tx_status`             TINYINT     NOT NULL DEFAULT 0 ,
    `tx_trigger`            VARCHAR(1024) ,
    `tx_msg`                BLOB ,
    PRIMARY KEY (`event_id`),
    INDEX `idx_query` (`tx_status`, `sys_utime`)
) ENGINE = InnoDB
  DEFAULT CHARSET = utf8mb4;