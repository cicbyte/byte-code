-- 同 sqlite/75：项目移交邀请队列表（邀请制转交）
CREATE TABLE IF NOT EXISTS `project_transfers` (
    `id`             INT UNSIGNED NOT NULL AUTO_INCREMENT PRIMARY KEY,
    `project_id`     INT NOT NULL,
    `from_user_id`   INT NOT NULL,
    `to_user_id`    INT NOT NULL,
    `leave_project`  INT NOT NULL DEFAULT 0,
    `status`         VARCHAR(16) NOT NULL DEFAULT 'pending' CHECK(`status` IN ('pending','accepted','declined','cancelled')),
    `created_at`     DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    `resolved_at`    VARCHAR(19) NOT NULL DEFAULT '',
    `resolved_by`    INT NOT NULL DEFAULT 0,
    KEY `idx_pt_project_status` (`project_id`, `status`),
    KEY `idx_pt_to_user_status` (`to_user_id`, `status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
