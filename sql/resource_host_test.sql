
drop table resource_host_test;

CREATE TABLE `resource_host_test` (
                                      `id` int(11) NOT NULL AUTO_INCREMENT COMMENT '自增id',
                                      `name` varchar(200) NOT NULL COMMENT '资源名称',
                                      `tags` varchar(1024)  DEFAULT ''  COMMENT '标签map',
                                      `private_ips` varchar(1024)  DEFAULT ''  COMMENT '内网IP数组',
                                      PRIMARY KEY (`id`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4;

