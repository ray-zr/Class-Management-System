-- Class Management System v2.0.0 migration for MySQL 8.
-- Back up the database before running this script.
-- The script is idempotent and preserves score history.

SET NAMES utf8mb4;
SET @cms_schema := DATABASE();

-- Students are soft-deleted from v2.0.0 onward.
SET @cms_ddl := (
    SELECT IF(COUNT(*) = 0,
        'ALTER TABLE `students` ADD COLUMN `deleted_at` datetime(3) NULL',
        'DO 0')
    FROM information_schema.columns
    WHERE table_schema = @cms_schema
      AND table_name = 'students'
      AND column_name = 'deleted_at'
);
PREPARE cms_migration_stmt FROM @cms_ddl;
EXECUTE cms_migration_stmt;
DEALLOCATE PREPARE cms_migration_stmt;

SET @cms_ddl := (
    SELECT IF(COUNT(*) = 0,
        'ALTER TABLE `students` ADD INDEX `idx_students_deleted_at` (`deleted_at`)',
        'DO 0')
    FROM information_schema.statistics
    WHERE table_schema = @cms_schema
      AND table_name = 'students'
      AND index_name = 'idx_students_deleted_at'
);
PREPARE cms_migration_stmt FROM @cms_ddl;
EXECUTE cms_migration_stmt;
DEALLOCATE PREPARE cms_migration_stmt;

-- Score-entry idempotency key and immutable display snapshots.
SET @cms_ddl := (
    SELECT IF(COUNT(*) = 0,
        'ALTER TABLE `score_entries` ADD COLUMN `request_id` varchar(64) NOT NULL DEFAULT ''''',
        'DO 0')
    FROM information_schema.columns
    WHERE table_schema = @cms_schema
      AND table_name = 'score_entries'
      AND column_name = 'request_id'
);
PREPARE cms_migration_stmt FROM @cms_ddl;
EXECUTE cms_migration_stmt;
DEALLOCATE PREPARE cms_migration_stmt;

SET @cms_ddl := (
    SELECT IF(COUNT(*) = 0,
        'ALTER TABLE `score_entries` ADD COLUMN `student_no_snapshot` varchar(64) NOT NULL DEFAULT ''''',
        'DO 0')
    FROM information_schema.columns
    WHERE table_schema = @cms_schema
      AND table_name = 'score_entries'
      AND column_name = 'student_no_snapshot'
);
PREPARE cms_migration_stmt FROM @cms_ddl;
EXECUTE cms_migration_stmt;
DEALLOCATE PREPARE cms_migration_stmt;

SET @cms_ddl := (
    SELECT IF(COUNT(*) = 0,
        'ALTER TABLE `score_entries` ADD COLUMN `student_name_snapshot` varchar(64) NOT NULL DEFAULT ''''',
        'DO 0')
    FROM information_schema.columns
    WHERE table_schema = @cms_schema
      AND table_name = 'score_entries'
      AND column_name = 'student_name_snapshot'
);
PREPARE cms_migration_stmt FROM @cms_ddl;
EXECUTE cms_migration_stmt;
DEALLOCATE PREPARE cms_migration_stmt;

SET @cms_ddl := (
    SELECT IF(COUNT(*) = 0,
        'ALTER TABLE `score_entries` ADD COLUMN `group_name_snapshot` varchar(64) NOT NULL DEFAULT ''''',
        'DO 0')
    FROM information_schema.columns
    WHERE table_schema = @cms_schema
      AND table_name = 'score_entries'
      AND column_name = 'group_name_snapshot'
);
PREPARE cms_migration_stmt FROM @cms_ddl;
EXECUTE cms_migration_stmt;
DEALLOCATE PREPARE cms_migration_stmt;

SET @cms_ddl := (
    SELECT IF(COUNT(*) = 0,
        'ALTER TABLE `score_entries` ADD COLUMN `dimension_name_snapshot` varchar(64) NOT NULL DEFAULT ''''',
        'DO 0')
    FROM information_schema.columns
    WHERE table_schema = @cms_schema
      AND table_name = 'score_entries'
      AND column_name = 'dimension_name_snapshot'
);
PREPARE cms_migration_stmt FROM @cms_ddl;
EXECUTE cms_migration_stmt;
DEALLOCATE PREPARE cms_migration_stmt;

SET @cms_ddl := (
    SELECT IF(COUNT(*) = 0,
        'ALTER TABLE `score_entries` ADD COLUMN `score_item_name_snapshot` varchar(128) NOT NULL DEFAULT ''''',
        'DO 0')
    FROM information_schema.columns
    WHERE table_schema = @cms_schema
      AND table_name = 'score_entries'
      AND column_name = 'score_item_name_snapshot'
);
PREPARE cms_migration_stmt FROM @cms_ddl;
EXECUTE cms_migration_stmt;
DEALLOCATE PREPARE cms_migration_stmt;

SET @cms_ddl := (
    SELECT IF(COUNT(*) = 0,
        'ALTER TABLE `score_entries` ADD INDEX `idx_score_entries_request_id` (`request_id`)',
        'DO 0')
    FROM information_schema.statistics
    WHERE table_schema = @cms_schema
      AND table_name = 'score_entries'
      AND index_name = 'idx_score_entries_request_id'
);
PREPARE cms_migration_stmt FROM @cms_ddl;
EXECUTE cms_migration_stmt;
DEALLOCATE PREPARE cms_migration_stmt;

-- One operation row owns each client request ID.
CREATE TABLE IF NOT EXISTS `score_operations` (
    `id` bigint NOT NULL AUTO_INCREMENT,
    `created_at` datetime(3) NULL,
    `updated_at` datetime(3) NULL,
    `request_id` varchar(64) NOT NULL,
    `fingerprint` char(64) NOT NULL,
    `last_entry_id` bigint NOT NULL DEFAULT 0,
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_score_operations_request_id` (`request_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

SET @cms_ddl := (
    SELECT IF(COUNT(*) = 0,
        'ALTER TABLE `score_operations` ADD UNIQUE INDEX `idx_score_operations_request_id` (`request_id`)',
        'DO 0')
    FROM information_schema.statistics
    WHERE table_schema = @cms_schema
      AND table_name = 'score_operations'
      AND index_name = 'idx_score_operations_request_id'
);
PREPARE cms_migration_stmt FROM @cms_ddl;
EXECUTE cms_migration_stmt;
DEALLOCATE PREPARE cms_migration_stmt;

-- Keep one transient pick row before enforcing fair-round uniqueness.
DELETE duplicate
FROM `rollcall_picked` AS duplicate
JOIN `rollcall_picked` AS original
  ON duplicate.`round_id` = original.`round_id`
 AND duplicate.`student_id` = original.`student_id`
 AND duplicate.`id` > original.`id`;

SET @cms_ddl := (
    SELECT IF(COUNT(*) = 0,
        'ALTER TABLE `rollcall_picked` ADD UNIQUE INDEX `idx_rollcall_round_student` (`round_id`, `student_id`)',
        'DO 0')
    FROM information_schema.statistics
    WHERE table_schema = @cms_schema
      AND table_name = 'rollcall_picked'
      AND index_name = 'idx_rollcall_round_student'
);
PREPARE cms_migration_stmt FROM @cms_ddl;
EXECUTE cms_migration_stmt;
DEALLOCATE PREPARE cms_migration_stmt;

-- Backfill the best available names without overwriting existing snapshots.
UPDATE `score_entries` AS e
LEFT JOIN `students` AS s ON s.`id` = e.`student_id`
LEFT JOIN `groups` AS g ON g.`id` = e.`group_id`
LEFT JOIN `dimensions` AS d ON d.`id` = e.`dimension_id`
LEFT JOIN `score_items` AS si ON si.`id` = e.`score_item_id`
SET
    e.`student_no_snapshot` = IF(e.`student_no_snapshot` = '', COALESCE(s.`student_no`, ''), e.`student_no_snapshot`),
    e.`student_name_snapshot` = IF(e.`student_name_snapshot` = '', COALESCE(s.`name`, ''), e.`student_name_snapshot`),
    e.`group_name_snapshot` = IF(e.`group_name_snapshot` = '', COALESCE(g.`name`, ''), e.`group_name_snapshot`),
    e.`dimension_name_snapshot` = IF(e.`dimension_name_snapshot` = '', COALESCE(d.`name`, ''), e.`dimension_name_snapshot`),
    e.`score_item_name_snapshot` = IF(e.`score_item_name_snapshot` = '', COALESCE(si.`name`, ''), e.`score_item_name_snapshot`)
WHERE e.`student_no_snapshot` = ''
   OR e.`student_name_snapshot` = ''
   OR e.`group_name_snapshot` = ''
   OR e.`dimension_name_snapshot` = ''
   OR e.`score_item_name_snapshot` = '';

SET @cms_schema := NULL;
SET @cms_ddl := NULL;
