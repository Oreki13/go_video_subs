CREATE DATABASE IF NOT EXISTS subscription_db
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE subscription_db;

CREATE TABLE users (
  id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name       VARCHAR(100)    NOT NULL,
  email      VARCHAR(150)    NOT NULL,
  password   VARCHAR(255)    NOT NULL,
  created_at TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (id),
  UNIQUE KEY uq_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE video_categories (
  id          BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  name        ENUM('gold', 'silver', 'bronze') NOT NULL,
  description VARCHAR(255)    NULL,
  created_at  TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY (id),
  UNIQUE KEY uq_video_categories_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO video_categories (name, description) VALUES
  ('gold',   'Exclusive content for Gold subscribers'),
  ('silver', 'Premium content for Silver and above'),
  ('bronze', 'Basic content accessible to all subscribers');

CREATE TABLE videos (
  id               BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  category_id      BIGINT UNSIGNED NOT NULL,
  title            VARCHAR(255)    NOT NULL,
  description      TEXT            NULL,
  url              VARCHAR(500)    NOT NULL,
  duration_seconds INT UNSIGNED    NOT NULL DEFAULT 0,
  created_at       TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at       TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (id),
  CONSTRAINT fk_videos_category
    FOREIGN KEY (category_id) REFERENCES video_categories (id)
    ON DELETE RESTRICT ON UPDATE CASCADE,
  KEY idx_videos_category_id (category_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE subscription_plans (
  id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  tier          ENUM('gold', 'silver', 'bronze') NOT NULL,
  price         DECIMAL(12, 2)  NOT NULL,
  duration_days INT UNSIGNED    NOT NULL,
  is_active     TINYINT(1)      NOT NULL DEFAULT 1,
  created_at    TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at    TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (id),
  UNIQUE KEY uq_plan_tier (tier)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO subscription_plans (tier, price, duration_days) VALUES
  ('gold',   150000.00, 30),
  ('silver',  99000.00, 30),
  ('bronze',  49000.00, 30);


CREATE TABLE subscriptions (
  id         BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id    BIGINT UNSIGNED NOT NULL,
  plan_id    BIGINT UNSIGNED NOT NULL,
  tier       ENUM('gold', 'silver', 'bronze') NOT NULL,
  status     ENUM('active', 'inactive', 'expired') NOT NULL DEFAULT 'inactive',
  started_at TIMESTAMP       NULL,
  expired_at TIMESTAMP       NULL,
  created_at TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (id),
  UNIQUE KEY uq_subscriptions_user_id (user_id),
  CONSTRAINT fk_subscriptions_user
    FOREIGN KEY (user_id) REFERENCES users (id)
    ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_subscriptions_plan
    FOREIGN KEY (plan_id) REFERENCES subscription_plans (id)
    ON DELETE RESTRICT ON UPDATE CASCADE,
  KEY idx_subscriptions_status (status),
  KEY idx_subscriptions_expired_at (expired_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


CREATE TABLE payment_transactions (
  id                  BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  user_id             BIGINT UNSIGNED NOT NULL,
  plan_id             BIGINT UNSIGNED NOT NULL,
  subscription_id     BIGINT UNSIGNED NULL,
  external_payment_id VARCHAR(255)    NOT NULL,
  tier                ENUM('gold', 'silver', 'bronze') NOT NULL,
  amount              DECIMAL(12, 2)  NOT NULL,
  status              ENUM('pending', 'success', 'failed') NOT NULL DEFAULT 'pending',
  payload_raw         TEXT            NULL,
  paid_at             TIMESTAMP       NULL,
  created_at          TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at          TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (id),
  UNIQUE KEY uq_payment_external_id (external_payment_id),
  CONSTRAINT fk_payment_user
    FOREIGN KEY (user_id) REFERENCES users (id)
    ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT fk_payment_plan
    FOREIGN KEY (plan_id) REFERENCES subscription_plans (id)
    ON DELETE RESTRICT ON UPDATE CASCADE,
  CONSTRAINT fk_payment_subscription
    FOREIGN KEY (subscription_id) REFERENCES subscriptions (id)
    ON DELETE SET NULL ON UPDATE CASCADE,
  KEY idx_payment_user_id (user_id),
  KEY idx_payment_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;