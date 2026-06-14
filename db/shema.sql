-- ============================================================
--  YMMO — Schéma base de données MySQL
--  Version : 1.0
-- ============================================================

CREATE DATABASE IF NOT EXISTS ymmo CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE ymmo;

-- ── Users ────────────────────────────────────────────────────
CREATE TABLE users (
                       id         INT UNSIGNED    NOT NULL AUTO_INCREMENT,
                       first_name VARCHAR(100)    NOT NULL,
                       last_name  VARCHAR(100)    NOT NULL,
                       email      VARCHAR(255)    NOT NULL,
                       password   VARCHAR(255)    NOT NULL,               -- bcrypt hash
                       role       ENUM('client','agent','admin')
                               NOT NULL DEFAULT 'client',
                       phone      VARCHAR(20)     DEFAULT NULL,
                       created_at TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
                           ON UPDATE CURRENT_TIMESTAMP,

                       PRIMARY KEY (id),
                       UNIQUE KEY uq_users_email (email)
);

-- ── Properties ───────────────────────────────────────────────
CREATE TABLE properties (
                            id          INT UNSIGNED    NOT NULL AUTO_INCREMENT,
                            title       VARCHAR(255)    NOT NULL,
                            description TEXT            NOT NULL,
                            price       DECIMAL(12,2)   NOT NULL,
                            surface     DECIMAL(8,2)    NOT NULL,              -- m²
                            rooms       TINYINT UNSIGNED NOT NULL DEFAULT 1,
                            bedrooms    TINYINT UNSIGNED NOT NULL DEFAULT 0,
                            type        ENUM('apartment','house','office','commercial','land')
                                NOT NULL,
                            status      ENUM('available','sold','rented','pending')
                                NOT NULL DEFAULT 'available',
                            transaction ENUM('sale','rental')
                                NOT NULL,
                            address     VARCHAR(255)    NOT NULL,
                            city        VARCHAR(100)    NOT NULL,
                            zip_code    CHAR(5)         NOT NULL,
                            latitude    DECIMAL(10,7)   DEFAULT NULL,
                            longitude   DECIMAL(10,7)   DEFAULT NULL,
                            agent_id    INT UNSIGNED    NOT NULL,
                            view_count  INT UNSIGNED    NOT NULL DEFAULT 0,
                            created_at  TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP,
                            updated_at  TIMESTAMP       NOT NULL DEFAULT CURRENT_TIMESTAMP
                                ON UPDATE CURRENT_TIMESTAMP,

                            PRIMARY KEY (id),
                            CONSTRAINT fk_properties_agent
                                FOREIGN KEY (agent_id) REFERENCES users(id)
                                    ON DELETE RESTRICT                             -- on ne supprime pas un agent qui a des biens
                                    ON UPDATE CASCADE,

    -- Index pour les filtres les plus fréquents
                            INDEX idx_properties_city        (city),
                            INDEX idx_properties_type        (type),
                            INDEX idx_properties_status      (status),
                            INDEX idx_properties_transaction (transaction),
                            INDEX idx_properties_price       (price),
                            INDEX idx_properties_agent       (agent_id)
);

-- ── Property Images ──────────────────────────────────────────
CREATE TABLE property_images (
                                 id          INT UNSIGNED  NOT NULL AUTO_INCREMENT,
                                 property_id INT UNSIGNED  NOT NULL,
                                 url         VARCHAR(500)  NOT NULL,
                                 is_primary  TINYINT(1)    NOT NULL DEFAULT 0,
                                 created_at  TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,

                                 PRIMARY KEY (id),
                                 CONSTRAINT fk_images_property
                                     FOREIGN KEY (property_id) REFERENCES properties(id)
                                         ON DELETE CASCADE                              -- supprime les images si le bien est supprimé
                                         ON UPDATE CASCADE,

                                 INDEX idx_images_property (property_id)
);

-- ── Contact Messages ─────────────────────────────────────────
-- Une conversation = tous les messages partageant le même
-- triplet (property_id, client_id, agent_id). sender_role indique
-- qui a écrit CE message précis ('client' ou 'agent').
CREATE TABLE contact_messages (
                                  id          INT UNSIGNED  NOT NULL AUTO_INCREMENT,
                                  property_id INT UNSIGNED  NOT NULL,
                                  client_id   INT UNSIGNED  NOT NULL,             -- partie "client" du fil (constant)
                                  agent_id    INT UNSIGNED  NOT NULL,             -- partie "agent" du fil (constant)
                                  sender_id   INT UNSIGNED  NOT NULL,             -- auteur réel de CE message
                                  sender_role ENUM('client','agent') NOT NULL,    -- rôle de l'auteur pour ce message
                                  message     TEXT          NOT NULL,
                                  is_read     TINYINT(1)    NOT NULL DEFAULT 0,
                                  created_at  TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,

                                  PRIMARY KEY (id),

                                  CONSTRAINT fk_messages_property
                                      FOREIGN KEY (property_id) REFERENCES properties(id)
                                          ON DELETE CASCADE
                                          ON UPDATE CASCADE,
                                  CONSTRAINT fk_messages_client
                                      FOREIGN KEY (client_id) REFERENCES users(id)
                                          ON DELETE CASCADE
                                          ON UPDATE CASCADE,
                                  CONSTRAINT fk_messages_sender
                                      FOREIGN KEY (sender_id) REFERENCES users(id)
                                          ON DELETE CASCADE
                                          ON UPDATE CASCADE,
                                  CONSTRAINT fk_messages_agent
                                      FOREIGN KEY (agent_id) REFERENCES users(id)
                                          ON DELETE CASCADE
                                          ON UPDATE CASCADE,

                                  INDEX idx_messages_agent    (agent_id),
                                  INDEX idx_messages_property (property_id),
                                  INDEX idx_messages_thread   (property_id, client_id, agent_id, created_at),
                                  INDEX idx_messages_agent_unread  (agent_id,  sender_role, is_read),
                                  INDEX idx_messages_client_unread (client_id, sender_role, is_read)
);

-- ── Token Blacklist ──────────────────────────────────────────
-- Stocke les JWT invalidés (logout) jusqu'à leur expiration naturelle
CREATE TABLE token_blacklist (
                                 id         INT UNSIGNED  NOT NULL AUTO_INCREMENT,
                                 token_hash VARCHAR(64)   NOT NULL,                 -- SHA256 du token
                                 expires_at TIMESTAMP     NOT NULL,
                                 created_at TIMESTAMP     NOT NULL DEFAULT CURRENT_TIMESTAMP,

                                 PRIMARY KEY (id),
                                 UNIQUE KEY uq_token_hash (token_hash),
                                 INDEX idx_token_expires (expires_at)               -- pour purger les tokens expirés
);

-- ============================================================
--  Données initiales
-- ============================================================

-- Compte admin par défaut (mot de passe : Admin1234! — à changer)
-- Hash bcrypt généré avec cost=12
INSERT INTO users (first_name, last_name, email, password, role) VALUES
    ('Admin', 'Ymmo', 'admin@ymmo.fr',
     '$2a$12$placeholderHashToReplaceOnFirstRun000000000000000000000',
     'admin');