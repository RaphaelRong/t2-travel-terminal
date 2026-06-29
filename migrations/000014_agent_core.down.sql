-- 000014_agent_core.down.sql
-- 回滚 Agent package 核心表

DROP TABLE IF EXISTS agent_user_dataset_rows CASCADE;
DROP TABLE IF EXISTS agent_user_datasets CASCADE;
DROP TABLE IF EXISTS agent_session_projects CASCADE;
DROP TABLE IF EXISTS agent_messages CASCADE;
DROP TABLE IF EXISTS agent_sessions CASCADE;
DROP TABLE IF EXISTS agent_memories CASCADE;
DROP TABLE IF EXISTS agent_user_profiles CASCADE;
DROP TABLE IF EXISTS agent_souls CASCADE;
DROP TABLE IF EXISTS god_configs CASCADE;
