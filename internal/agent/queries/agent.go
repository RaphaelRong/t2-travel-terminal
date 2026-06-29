package queries

// agent_souls
const (
	SoulInsert = `
		INSERT INTO agent_souls (scope, user_id, name, identity_text, voice_text, values_text, allowed_domains, forbidden_domains, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	SoulUpdate = `
		UPDATE agent_souls
		SET name = $1, identity_text = $2, voice_text = $3, values_text = $4,
		    allowed_domains = $5, forbidden_domains = $6, metadata = $7, updated_at = now()
		WHERE id = $8
	`

	SoulDelete = `DELETE FROM agent_souls WHERE id = $1`

	SoulSelectByID = `
		SELECT id, scope, user_id, name, identity_text, voice_text, values_text,
		       allowed_domains, forbidden_domains, metadata, created_at, updated_at
		FROM agent_souls
		WHERE id = $1
	`

	SoulsSelectByUser = `
		SELECT id, scope, user_id, name, identity_text, voice_text, values_text,
		       allowed_domains, forbidden_domains, metadata, created_at, updated_at
		FROM agent_souls
		WHERE user_id = $1
		ORDER BY updated_at DESC
	`

	SoulsSelectSystem = `
		SELECT id, scope, user_id, name, identity_text, voice_text, values_text,
		       allowed_domains, forbidden_domains, metadata, created_at, updated_at
		FROM agent_souls
		WHERE scope = 'system'
		ORDER BY updated_at DESC
	`
)

// agent_memories
const (
	MemoryInsert = `
		INSERT INTO agent_memories (user_id, category, content, source_session_id, source_message_id, confidence, expires_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	MemoryUpdate = `
		UPDATE agent_memories
		SET category = $1, content = $2, confidence = $3, expires_at = $4, metadata = $5, updated_at = now()
		WHERE id = $6
	`

	MemoryDelete = `DELETE FROM agent_memories WHERE id = $1`

	MemorySelectByID = `
		SELECT id, user_id, category, content, source_session_id, source_message_id,
		       confidence, expires_at, metadata, created_at, updated_at
		FROM agent_memories
		WHERE id = $1
	`

	MemoriesSelectByUser = `
		SELECT id, user_id, category, content, source_session_id, source_message_id,
		       confidence, expires_at, metadata, created_at, updated_at
		FROM agent_memories
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2
	`

	MemoriesSearch = `
		SELECT id, user_id, category, content, source_session_id, source_message_id,
		       confidence, expires_at, metadata, created_at, updated_at
		FROM agent_memories
		WHERE user_id = $1
		  AND content ILIKE '%' || $2 || '%'
		ORDER BY updated_at DESC
		LIMIT $3
	`
)

// agent_sessions
const (
	SessionInsert = `
		INSERT INTO agent_sessions (user_id, tenant_id, title, status, parent_session_id, context_summary, context_summary_at, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	SessionUpdate = `
		UPDATE agent_sessions
		SET title = $1, status = $2, context_summary = $3, context_summary_at = $4, metadata = $5, updated_at = now()
		WHERE id = $6
	`

	SessionSelectByID = `
		SELECT id, user_id, tenant_id, title, status, parent_session_id, context_summary,
		       context_summary_at, metadata, created_at, updated_at
		FROM agent_sessions
		WHERE id = $1
	`

	SessionsSelectByUser = `
		SELECT id, user_id, tenant_id, title, status, parent_session_id, context_summary,
		       context_summary_at, metadata, created_at, updated_at
		FROM agent_sessions
		WHERE user_id = $1
		  AND status = $2
		ORDER BY updated_at DESC
		LIMIT $3
	`
)

// agent_messages
const (
	MessageInsert = `
		INSERT INTO agent_messages (session_id, role, content, tool_calls, tool_call_id, tool_name, tool_result, reasoning_content, token_count, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`

	MessagesSelectBySession = `
		SELECT id, session_id, role, content, tool_calls, tool_call_id, tool_name, tool_result,
		       reasoning_content, token_count, metadata, created_at
		FROM agent_messages
		WHERE session_id = $1
		ORDER BY created_at ASC
		LIMIT $2
	`
)

// agent_session_projects
const (
	SessionProjectInsert = `
		INSERT INTO agent_session_projects (session_id, project_id)
		VALUES ($1, $2)
		ON CONFLICT (session_id, project_id) DO NOTHING
	`

	SessionProjectDelete = `
		DELETE FROM agent_session_projects
		WHERE session_id = $1 AND project_id = $2
	`

	SessionProjectSelectBySession = `
		SELECT project_id FROM agent_session_projects WHERE session_id = $1
	`
)

// agent_user_profiles
const (
	UserProfileInsert = `
		INSERT INTO agent_user_profiles (user_id, soul_id, default_llm_profile_id, status)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id) DO UPDATE SET
			soul_id = EXCLUDED.soul_id,
			default_llm_profile_id = EXCLUDED.default_llm_profile_id,
			status = EXCLUDED.status,
			updated_at = now()
	`

	UserProfileSelectByUser = `
		SELECT id, user_id, soul_id, default_llm_profile_id, status, created_at, updated_at
		FROM agent_user_profiles
		WHERE user_id = $1
	`
)

// god_configs
const (
	GodConfigInsert = `
		INSERT INTO god_configs (name, is_active, allowed_domains, forbidden_domains, allowed_tools, forbidden_tools,
		                         require_approval_tools, max_iterations, can_delegate, can_run_workflow, rules)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	GodConfigUpdate = `
		UPDATE god_configs
		SET is_active = $1, allowed_domains = $2, forbidden_domains = $3, allowed_tools = $4,
		    forbidden_tools = $5, require_approval_tools = $6, max_iterations = $7,
		    can_delegate = $8, can_run_workflow = $9, rules = $10, updated_at = now()
		WHERE id = $11
	`

	GodConfigSelectByName = `
		SELECT id, name, is_active, allowed_domains, forbidden_domains, allowed_tools, forbidden_tools,
		       require_approval_tools, max_iterations, can_delegate, can_run_workflow, rules, created_at, updated_at
		FROM god_configs
		WHERE name = $1
	`

	GodConfigSelectActive = `
		SELECT id, name, is_active, allowed_domains, forbidden_domains, allowed_tools, forbidden_tools,
		       require_approval_tools, max_iterations, can_delegate, can_run_workflow, rules, created_at, updated_at
		FROM god_configs
		WHERE is_active = true
		LIMIT 1
	`

	GodConfigList = `
		SELECT id, name, is_active, allowed_domains, forbidden_domains, allowed_tools, forbidden_tools,
		       require_approval_tools, max_iterations, can_delegate, can_run_workflow, rules, created_at, updated_at
		FROM god_configs
		ORDER BY updated_at DESC
	`
)

// projects / project_capabilities / project_integrations（Agent 视角）
const (
	ProjectsSelectBySession = `
		SELECT p.id, p.tenant_id, p.source_scope, p.kind, p.status, p.source_type,
		       p.name, p.description, p.endpoint_url, p.request_method, p.request_path,
		       p.request_headers, p.request_body_template, p.auth_type, p.auth_config,
		       p.capability_summary, p.created_by, p.created_at, p.updated_at, p.last_published_at
		FROM projects p
		JOIN agent_session_projects asp ON asp.project_id = p.id
		WHERE asp.session_id = $1
		ORDER BY p.name
	`

	ProjectCapabilitiesSelectBySession = `
		SELECT c.id, c.project_id, c.integration_id, c.kind, c.name, c.external_name, c.description,
		       c.status, c.request_method, c.request_path, c.input_schema, c.output_schema, c.metadata,
		       c.created_at, c.updated_at,
		       i.kind AS integration_kind,
		       p.endpoint_url AS project_endpoint_url,
		       p.request_headers AS project_request_headers,
		       p.auth_type AS project_auth_type,
		       p.auth_config AS project_auth_config,
		       i.endpoint_url AS integration_endpoint_url,
		       i.request_headers AS integration_request_headers,
		       i.auth_type AS integration_auth_type,
		       i.auth_config AS integration_auth_config
		FROM project_capabilities c
		JOIN agent_session_projects asp ON asp.project_id = c.project_id
		JOIN projects p ON p.id = c.project_id
		LEFT JOIN project_integrations i ON i.id = c.integration_id
		WHERE asp.session_id = $1
		ORDER BY c.name
	`

	ProjectCapabilitySelectByID = `
		SELECT c.id, c.project_id, c.integration_id, c.kind, c.name, c.external_name, c.description,
		       c.status, c.request_method, c.request_path, c.input_schema, c.output_schema, c.metadata,
		       c.created_at, c.updated_at,
		       i.kind AS integration_kind,
		       p.endpoint_url AS project_endpoint_url,
		       p.request_headers AS project_request_headers,
		       p.auth_type AS project_auth_type,
		       p.auth_config AS project_auth_config,
		       i.endpoint_url AS integration_endpoint_url,
		       i.request_headers AS integration_request_headers,
		       i.auth_type AS integration_auth_type,
		       i.auth_config AS integration_auth_config
		FROM project_capabilities c
		JOIN projects p ON p.id = c.project_id
		LEFT JOIN project_integrations i ON i.id = c.integration_id
		WHERE c.id = $1
	`

	ProjectSelectByID = `
		SELECT id, tenant_id, source_scope, kind, status, source_type,
		       name, description, endpoint_url, request_method, request_path,
		       request_headers, request_body_template, auth_type, auth_config,
		       capability_summary, created_by, created_at, updated_at, last_published_at
		FROM projects
		WHERE id = $1
	`

	ProjectIntegrationSelectByID = `
		SELECT id, project_id, kind, name, description, status,
		       endpoint_url, documentation_url, transport, auth_type, request_headers,
		       auth_config, metadata, last_synced_at, sync_status, sync_error,
		       created_at, updated_at
		FROM project_integrations
		WHERE id = $1
	`
)

// agent_user_datasets
const (
	UserDatasetInsert = `
		INSERT INTO agent_user_datasets (user_id, name, description, schema, row_count, source, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	UserDatasetUpdate = `
		UPDATE agent_user_datasets
		SET name = $1, description = $2, schema = $3, row_count = $4, source = $5, metadata = $6, updated_at = now()
		WHERE id = $7
	`

	UserDatasetDelete = `DELETE FROM agent_user_datasets WHERE id = $1`

	UserDatasetSelectByID = `
		SELECT id, user_id, name, description, schema, row_count, source, metadata, created_at, updated_at
		FROM agent_user_datasets
		WHERE id = $1
	`

	UserDatasetSelectByName = `
		SELECT id, user_id, name, description, schema, row_count, source, metadata, created_at, updated_at
		FROM agent_user_datasets
		WHERE user_id = $1 AND name = $2
	`

	UserDatasetsSelectByUser = `
		SELECT id, user_id, name, description, schema, row_count, source, metadata, created_at, updated_at
		FROM agent_user_datasets
		WHERE user_id = $1
		ORDER BY updated_at DESC
		LIMIT $2
	`

	UserDatasetRowsInsert = `
		INSERT INTO agent_user_dataset_rows (dataset_id, user_id, row_index, data)
		VALUES ($1, $2, $3, $4)
	`

	UserDatasetRowsDeleteByDataset = `
		DELETE FROM agent_user_dataset_rows WHERE dataset_id = $1
	`
)
