package queries

const (
	// ProjectsListAccessible 查询当前租户可访问的项目列表，包含 online 系统项目和租户项目。
	ProjectsListAccessible = `SELECT id, tenant_id, source_scope, kind, status, source_type,
	        name, description, endpoint_url, request_method, request_path,
	        request_headers, request_body_template, auth_type, auth_config,
	        capability_summary, created_by, created_at, updated_at, last_published_at
	 FROM projects
	 WHERE tenant_id = $1 OR source_scope = 'system'
	 ORDER BY source_scope, created_at DESC`

	// ProjectsInsert 创建项目。
	ProjectsInsert = `INSERT INTO projects (
	     tenant_id, source_scope, kind, status, source_type,
	     name, description, endpoint_url, request_method, request_path,
	     request_headers, request_body_template, auth_type, auth_config,
	     capability_summary, created_by, last_published_at
	 )
	 VALUES (
	     $1::uuid, $2, 'data_source', $3, $4,
	     $5, NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''), NULLIF($9, ''),
	     $10, $11, NULLIF($12, ''), $13,
	     NULLIF($14, ''), $15,
	     CASE WHEN $3 = 'online' THEN now() ELSE NULL END
	 )
	 RETURNING id`

	// ProjectsSelectByID 查询单个项目。
	ProjectsSelectByID = `SELECT id, tenant_id, source_scope, kind, status, source_type,
	        name, description, endpoint_url, request_method, request_path,
	        request_headers, request_body_template, auth_type, auth_config,
	        capability_summary, created_by, created_at, updated_at, last_published_at
	 FROM projects
	 WHERE id = $1`

	// ProjectsUpdate 更新项目。
	ProjectsUpdate = `UPDATE projects
	 SET status = COALESCE(NULLIF($1, ''), status),
	     source_type = COALESCE(NULLIF($2, ''), source_type),
	     name = COALESCE(NULLIF($3, ''), name),
	     description = COALESCE(NULLIF($4, ''), description),
	     endpoint_url = COALESCE(NULLIF($5, ''), endpoint_url),
	     request_method = COALESCE(NULLIF($6, ''), request_method),
	     request_path = COALESCE(NULLIF($7, ''), request_path),
	     request_headers = COALESCE($8, request_headers),
	     request_body_template = COALESCE($9, request_body_template),
	     auth_type = COALESCE(NULLIF($10, ''), auth_type),
	     auth_config = COALESCE($11, auth_config),
	     capability_summary = COALESCE(NULLIF($12, ''), capability_summary),
	     last_published_at = CASE
	         WHEN COALESCE(NULLIF($1, ''), status) = 'online' AND status <> 'online' THEN now()
	         ELSE last_published_at
	     END,
	     updated_at = now()
	 WHERE id = $13`

	// ProjectsDelete 删除项目。
	ProjectsDelete = `DELETE FROM projects WHERE id = $1`

	// ProjectCapabilitiesList 查询项目能力列表。
	ProjectCapabilitiesList = `SELECT id, project_id, kind, name, description, status,
	        integration_id, request_method, request_path, input_schema, output_schema, metadata,
	        created_at, updated_at
	 FROM project_capabilities
	 WHERE project_id = $1
	 ORDER BY kind, name`

	// ProjectCapabilitiesDeleteByProject 删除某个项目下全部能力。
	ProjectCapabilitiesDeleteByProject = `DELETE FROM project_capabilities WHERE project_id = $1`

	// ProjectCapabilitiesSelectByID 查询单个项目能力。
	ProjectCapabilitiesSelectByID = `SELECT id, project_id, kind, name, description, status,
	        integration_id, request_method, request_path, input_schema, output_schema, metadata,
	        created_at, updated_at
	 FROM project_capabilities
	 WHERE project_id = $1 AND id = $2`

	// ProjectCapabilitiesInsert 创建项目能力。
	ProjectCapabilitiesInsert = `INSERT INTO project_capabilities (
	     project_id, kind, name, description, status, request_method, request_path,
	     input_schema, output_schema, metadata
	 )
	 VALUES (
	     $1, $2, $3, NULLIF($4, ''), COALESCE(NULLIF($5, ''), 'active'),
	     NULLIF($6, ''), NULLIF($7, ''), $8, $9, $10
	 )
	 RETURNING id`

	// ProjectCapabilitiesInsertForIntegration 创建来源同步能力。
	ProjectCapabilitiesInsertForIntegration = `INSERT INTO project_capabilities (
	     project_id, integration_id, kind, name, external_name, description, status,
	     request_method, request_path, input_schema, output_schema, metadata
	 )
	 VALUES (
	     $1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''),
	     COALESCE(NULLIF($7, ''), 'active'), NULLIF($8, ''), NULLIF($9, ''),
	     $10, $11, $12
	 )
	 RETURNING id`

	// ProjectCapabilitiesDeleteByIntegration 删除某个来源同步出的能力。
	ProjectCapabilitiesDeleteByIntegration = `DELETE FROM project_capabilities WHERE project_id = $1 AND integration_id = $2`

	// ProjectCapabilitiesUpdate 更新项目能力。
	ProjectCapabilitiesUpdate = `UPDATE project_capabilities
	 SET kind = COALESCE(NULLIF($1, ''), kind),
	     name = COALESCE(NULLIF($2, ''), name),
	     description = COALESCE(NULLIF($3, ''), description),
	     status = COALESCE(NULLIF($4, ''), status),
	     request_method = COALESCE(NULLIF($5, ''), request_method),
	     request_path = COALESCE(NULLIF($6, ''), request_path),
	     input_schema = COALESCE($7, input_schema),
	     output_schema = COALESCE($8, output_schema),
	     metadata = COALESCE($9, metadata),
	     updated_at = now()
	 WHERE project_id = $10 AND id = $11`

	// ProjectCapabilitiesDelete 删除项目能力。
	ProjectCapabilitiesDelete = `DELETE FROM project_capabilities WHERE project_id = $1 AND id = $2`

	// ProjectsListSystem 查询系统项目（SuperAdmin）。
	ProjectsListSystem = `SELECT id, tenant_id, source_scope, kind, status, source_type,
	        name, description, endpoint_url, request_method, request_path,
	        request_headers, request_body_template, auth_type, auth_config,
	        capability_summary, created_by, created_at, updated_at, last_published_at
	 FROM projects
	 WHERE source_scope = 'system'
	 ORDER BY created_at DESC`

	// ProjectIntegrationsList 查询项目能力来源。
	ProjectIntegrationsList = `SELECT id, project_id, kind, name, description, status,
	        endpoint_url, documentation_url, transport, auth_type, request_headers,
	        auth_config, metadata, last_synced_at, sync_status, sync_error,
	        created_at, updated_at
	 FROM project_integrations
	 WHERE project_id = $1
	 ORDER BY kind, name`

	// ProjectIntegrationsSelectByID 查询单个能力来源。
	ProjectIntegrationsSelectByID = `SELECT id, project_id, kind, name, description, status,
	        endpoint_url, documentation_url, transport, auth_type, request_headers,
	        auth_config, metadata, last_synced_at, sync_status, sync_error,
	        created_at, updated_at
	 FROM project_integrations
	 WHERE project_id = $1 AND id = $2`

	// ProjectIntegrationsInsert 创建能力来源。
	ProjectIntegrationsInsert = `INSERT INTO project_integrations (
	     project_id, kind, name, description, status, endpoint_url, documentation_url,
	     transport, auth_type, request_headers, auth_config, metadata
	 )
	 VALUES (
	     $1, $2, $3, NULLIF($4, ''), COALESCE(NULLIF($5, ''), 'active'),
	     NULLIF($6, ''), NULLIF($7, ''), COALESCE(NULLIF($8, ''), 'http'),
	     COALESCE(NULLIF($9, ''), 'inherit'), $10, $11, $12
	 )
	 RETURNING id`

	// ProjectIntegrationsUpdate 更新能力来源。
	ProjectIntegrationsUpdate = `UPDATE project_integrations
	 SET kind = COALESCE(NULLIF($1, ''), kind),
	     name = COALESCE(NULLIF($2, ''), name),
	     description = COALESCE(NULLIF($3, ''), description),
	     status = COALESCE(NULLIF($4, ''), status),
	     endpoint_url = COALESCE(NULLIF($5, ''), endpoint_url),
	     documentation_url = COALESCE(NULLIF($6, ''), documentation_url),
	     transport = COALESCE(NULLIF($7, ''), transport),
	     auth_type = COALESCE(NULLIF($8, ''), auth_type),
	     request_headers = COALESCE($9, request_headers),
	     auth_config = COALESCE($10, auth_config),
	     metadata = COALESCE($11, metadata),
	     updated_at = now()
	 WHERE project_id = $12 AND id = $13`

	// ProjectIntegrationsDelete 删除能力来源。
	ProjectIntegrationsDelete = `DELETE FROM project_integrations WHERE project_id = $1 AND id = $2`

	// ProjectIntegrationsUpdateSyncResult 更新同步状态。
	ProjectIntegrationsUpdateSyncResult = `UPDATE project_integrations
	 SET last_synced_at = now(),
	     sync_status = $1,
	     sync_error = NULLIF($2, ''),
	     updated_at = now()
	 WHERE project_id = $3 AND id = $4`
)
