-- ============================================================
--  YMMO — Migration : conversations bidirectionnelles
--  À exécuter sur une base existante (après schema.sql v1.0)
-- ============================================================
USE ymmo;

-- ── Ajout des colonnes pour gérer le fil de discussion ────────
-- client_id  : l'utilisateur "client" de la conversation (constant
--               sur tout le fil, même quand l'agent répond)
-- sender_role: qui a écrit CE message précis ('client' ou 'agent')
ALTER TABLE contact_messages
    ADD COLUMN client_id   INT UNSIGNED NULL                         AFTER property_id,
    ADD COLUMN sender_role ENUM('client','agent') NOT NULL
                            DEFAULT 'client'                          AFTER sender_id;

-- ── Rétro-compatibilité : les messages existants viennent tous
--    d'un client (l'ancien système n'avait que ce sens) ──────────
UPDATE contact_messages
SET client_id = sender_id,
    sender_role = 'client'
WHERE client_id IS NULL;

-- ── client_id devient obligatoire + clé étrangère ─────────────
ALTER TABLE contact_messages
    MODIFY COLUMN client_id INT UNSIGNED NOT NULL,
    ADD CONSTRAINT fk_messages_client
        FOREIGN KEY (client_id) REFERENCES users(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE;

-- ── Index pour retrouver rapidement un fil de discussion ──────
CREATE INDEX idx_messages_thread
    ON contact_messages (property_id, client_id, agent_id, created_at);

-- ── Index pour compter les messages non lus côté client ───────
CREATE INDEX idx_messages_client_unread
    ON contact_messages (client_id, sender_role, is_read);
