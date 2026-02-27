-- Rollback Phase 8: Multi-Channel Notifications & Integrations
DROP TABLE IF EXISTS outbound_events;
DROP TABLE IF EXISTS email_digests;
DROP TABLE IF EXISTS notification_rules;
DROP TABLE IF EXISTS integration_connections;
