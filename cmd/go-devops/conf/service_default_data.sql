-- 优先创建组件，后创建服务树
-- 为APP分类下的服务节点创建组件
INSERT INTO component (id, name, service_tree, owner, description, repo_url, repo_branch, created_at, updated_at) VALUES
(UUID(), 'push-server', 'DevOps/APP/push-server', 'admin', '推送服务组件，用于消息推送', '', 'main', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(UUID(), 'app-server', 'DevOps/APP/app-server', 'admin', '应用服务器组件，处理业务逻辑', '', 'main', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(UUID(), 'manager-server', 'DevOps/APP/manager-server', 'admin', '管理服务器组件，负责系统管理功能', '', 'main', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(UUID(), 'iam-server', 'DevOps/APP/iam-server', 'admin', '身份认证服务组件，管理用户认证和授权', '', 'main', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(UUID(), 'web-server', 'DevOps/APP/web-server', 'admin', 'Web服务器组件，提供前端页面服务', '', 'main', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(UUID(), 'keycloak', 'DevOps/APP/keycloak', 'admin', 'Keycloak认证服务组件，提供统一身份管理', '', 'main', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- 为PaaS-DB分类下的服务节点创建组件
INSERT INTO component (id, name, service_tree, owner, description, repo_url, repo_branch, created_at, updated_at) VALUES
(UUID(), 'S-APP-MysqlDB', 'DevOps/PaaS-DB/S-APP-MysqlDB', 'admin', 'MySQL数据库服务组件，提供数据存储服务', '', 'main', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(UUID(), 'S-APP-Redis', 'DevOps/PaaS-DB/S-APP-Redis', 'admin', 'Redis缓存服务组件，提供缓存支持', '', 'main', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- 为BasicService分类下的服务节点创建组件
INSERT INTO component (id, name, service_tree, owner, description, repo_url, repo_branch, created_at, updated_at) VALUES
(UUID(), 'PaaS-LVS', 'DevOps/BasicService/BasicService-LVS/PaaS-LVS', 'admin', 'LVS服务组件，提供负载均衡功能', '', 'main', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(UUID(), 'PaaS-LB', 'DevOps/BasicService/BasicService-LB/PaaS-LB', 'admin', '负载均衡服务组件，分发流量请求', '', 'main', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(UUID(), 'PaaS-DNS', 'DevOps/BasicService/BasicService-DNS/PaaS-DNS', 'admin', 'DNS服务组件，提供域名解析服务', '', 'main', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

SET @push_server_id = (SELECT id FROM component WHERE name = 'push-server');
SET @app_server_id = (SELECT id FROM component WHERE name = 'app-server');
SET @manager_server_id = (SELECT id FROM component WHERE name = 'manager-server');
SET @iam_server_id = (SELECT id FROM component WHERE name = 'iam-server');
SET @web_server_id = (SELECT id FROM component WHERE name = 'web-server');
SET @keycloak_id = (SELECT id FROM component WHERE name = 'keycloak');
SET @S_APP_MysqlDB_id = (SELECT id FROM component WHERE name = 'S-APP-MysqlDB');
SET @S_APP_Redis_id = (SELECT id FROM component WHERE name = 'S-APP-Redis');
SET @BasicService_LVS_id = (SELECT id FROM component WHERE name = 'BasicService-LVS');
SET @BasicService_LB_id = (SELECT id FROM component WHERE name = 'BasicService-LB');
SET @BasicService_DNS_id = (SELECT id FROM component WHERE name = 'BasicService-DNS');

-- 第一层: DevOps (根分类)
INSERT INTO service_tree (id, name, full_path, node_type, parent_id, level, service_id, description, created_at, updated_at)
VALUES (UUID(), 'DevOps', 'DevOps', 'category', 'root', 1, '', 'DevOps根分类', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
SET @devops_id = (SELECT id FROM service_tree WHERE name = 'DevOps');

-- 第二层: APP
INSERT INTO service_tree (id, name, full_path, node_type, parent_id, level, service_id, description, created_at, updated_at)
VALUES (UUID(), 'APP', 'DevOps/APP', 'category', @devops_id, 2, '', '应用服务分类', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
SET @app_id = (SELECT id FROM service_tree WHERE name = 'APP' AND parent_id = @devops_id);

-- APP下的服务节点
INSERT INTO service_tree (id, name, full_path, node_type, parent_id, level, service_id, description, created_at, updated_at) VALUES
(UUID(), 'push-server', 'DevOps/APP/push-server', 'service', @app_id, 3, @push_server_id, '推送服务', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(UUID(), 'app-server', 'DevOps/APP/app-server', 'service', @app_id, 3, @app_server_id, '应用服务器', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(UUID(), 'manager-server', 'DevOps/APP/manager-server', 'service', @app_id, 3, @manager_server_id, '管理服务器', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(UUID(), 'iam-server', 'DevOps/APP/iam-server', 'service', @app_id, 3, @iam_server_id, '身份认证服务', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(UUID(), 'web-server', 'DevOps/APP/web-server', 'service', @app_id, 3, @web_server_id, 'Web服务器', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(UUID(), 'keycloak', 'DevOps/APP/keycloak', 'service', @app_id, 3, @keycloak_id, 'Keycloak认证服务', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- 第二层: PaaS-DB
INSERT INTO service_tree (id, name, full_path, node_type, parent_id, level, service_id, description, created_at, updated_at)
VALUES (UUID(), 'PaaS-DB', 'DevOps/PaaS-DB', 'category', @devops_id, 2, '', '数据库服务分类', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
SET @paas_db_id = (SELECT id FROM service_tree WHERE name = 'PaaS-DB' AND parent_id = @devops_id);

-- PaaS-DB下的服务
INSERT INTO service_tree (id, name, full_path, node_type, parent_id, level, service_id, description, created_at, updated_at) VALUES
(UUID(), 'S-APP-MysqlDB', 'DevOps/PaaS-DB/S-APP-MysqlDB', 'service', @paas_db_id, 3, @S_APP_MysqlDB_id, 'MySQL数据库服务', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(UUID(), 'S-APP-Redis', 'DevOps/PaaS-DB/S-APP-Redis', 'service', @paas_db_id, 3, @S_APP_Redis_id, 'Redis缓存服务', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- 第二层: BasicService
INSERT INTO service_tree (id, name, full_path, node_type, parent_id, level, service_id, description, created_at, updated_at)
VALUES (UUID(), 'BasicService', 'DevOps/BasicService', 'category', @devops_id, 2, '', '基础服务分类', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);
SET @basic_service_id = (SELECT id FROM service_tree WHERE name = 'BasicService' AND parent_id = @devops_id);

-- BasicService下的三级分类
INSERT INTO service_tree (id, name, full_path, node_type, parent_id, level, service_id, description, created_at, updated_at) VALUES
(UUID(), 'BasicService-LVS', 'DevOps/BasicService/BasicService-LVS', 'category', @basic_service_id, 3, '', 'LVS服务分类', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(UUID(), 'BasicService-LB', 'DevOps/BasicService/BasicService-LB', 'category', @basic_service_id, 3, '', '负载均衡服务分类', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(UUID(), 'BasicService-DNS', 'DevOps/BasicService/BasicService-DNS', 'category', @basic_service_id, 3, '', 'DNS服务分类', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- 三级分类下的服务节点
SET @lvs_id = (SELECT id FROM service_tree WHERE name = 'BasicService-LVS' AND parent_id = @basic_service_id);
SET @lb_id = (SELECT id FROM service_tree WHERE name = 'BasicService-LB' AND parent_id = @basic_service_id);
SET @dns_id = (SELECT id FROM service_tree WHERE name = 'BasicService-DNS' AND parent_id = @basic_service_id);

INSERT INTO service_tree (id, name, full_path, node_type, parent_id, level, service_id, description, created_at, updated_at) VALUES
(UUID(), 'PaaS-LVS', 'DevOps/BasicService/BasicService-LVS/PaaS-LVS', 'service', @lvs_id, 4, @BasicService_LVS_id, 'LVS服务', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(UUID(), 'PaaS-LB', 'DevOps/BasicService/BasicService-LB/PaaS-LB', 'service', @lb_id, 4, @BasicService_LB_id, '负载均衡服务', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
(UUID(), 'PaaS-DNS', 'DevOps/BasicService/BasicService-DNS/PaaS-DNS', 'service', @dns_id, 4, @BasicService_DNS_id, 'DNS服务', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

